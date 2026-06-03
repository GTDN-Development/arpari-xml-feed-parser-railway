package hon

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/fanda/arpari-xml-feed-parser-railway/internal/shoptet"
)

const ProductsURL = "https://www.webshop.officepro-brno.cz/import/HONClientFeed/HONClientFeed.xml"

const supplierName = "HON"

type Downloader interface {
	Download(ctx context.Context, url string) (io.ReadCloser, error)
}

type HTTPDownloader struct {
	Client *http.Client
}

func (downloader HTTPDownloader) Download(ctx context.Context, url string) (io.ReadCloser, error) {
	client := downloader.Client
	if client == nil {
		client = &http.Client{Timeout: 2 * time.Minute}
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create HON request: %w", err)
	}

	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("download HON feed: %w", err)
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		_ = response.Body.Close()
		return nil, fmt.Errorf("download HON feed: unexpected HTTP status %s", response.Status)
	}

	return response.Body, nil
}

type ProductsOptions struct {
	MaxProducts int
}

type ProductsStats struct {
	ProductsRead      int
	ProductsEmitted   int
	ProductsSkipped   int
	ItemsWithVariants int
	VariantsEmitted   int
}

func ParseProducts(ctx context.Context, r io.Reader, options ProductsOptions) (shoptet.Feed, ProductsStats, error) {
	decoder := xml.NewDecoder(r)
	var stats ProductsStats
	var entries []productEntry

	for {
		if err := ctx.Err(); err != nil {
			return shoptet.Feed{}, stats, err
		}

		token, err := decoder.Token()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return emitProducts(entries, options, &stats), stats, nil
			}
			return shoptet.Feed{}, stats, fmt.Errorf("parse HON products XML: %w", err)
		}

		start, ok := token.(xml.StartElement)
		if !ok || start.Name.Local != "SHOPITEM" {
			continue
		}

		var source sourceItem
		if err := decoder.DecodeElement(&source, &start); err != nil {
			return shoptet.Feed{}, stats, fmt.Errorf("decode HON SHOPITEM: %w", err)
		}

		stats.ProductsRead++
		item, ok := transformProduct(source)
		if !ok {
			stats.ProductsSkipped++
			continue
		}

		entries = append(entries, productEntry{Source: source, Item: item})
	}
}

type productEntry struct {
	Source sourceItem
	Item   shoptet.Item
}

func transformProduct(source sourceItem) (shoptet.Item, bool) {
	code := strings.TrimSpace(source.PartNumber)
	if code == "" {
		code = strings.TrimSpace(source.ID)
	}
	name := transformName(source.Product, source.Description)
	if code == "" || name == "" {
		slog.Warn("skipping HON product without required identity", "code", code, "name", name)
		return shoptet.Item{}, false
	}

	categories, defaultCategory := transformCategory(source.MainCategory)
	return shoptet.Item{
		Code:                  code,
		Name:                  name,
		Description:           strings.TrimSpace(source.Description),
		Manufacturer:          supplierName,
		Supplier:              supplierName,
		PriceVAT:              strings.TrimSpace(source.PriceVAT),
		Stock:                 normalizeNumber(source.Stock),
		Availability:          strings.TrimSpace(source.Availability),
		Categories:            categories,
		DefaultCategory:       defaultCategory,
		Images:                transformImages(source.Images),
		InformationParameters: transformParameters(source.Parameters),
	}, true
}

func emitProducts(entries []productEntry, options ProductsOptions, stats *ProductsStats) shoptet.Feed {
	var result shoptet.Feed
	if len(entries) == 0 {
		return result
	}

	groups := variantGroups(entries)
	emittedGroups := make(map[string]struct{}, len(groups))
	parentCodes := make(map[string]struct{}, len(groups))
	for _, entry := range entries {
		key := variantGroupKey(entry.Source)
		if group, ok := groups[key]; ok {
			if _, emitted := emittedGroups[key]; emitted {
				continue
			}
			emittedGroups[key] = struct{}{}
			item := transformVariantGroup(group, uniqueParentCode(entry.Source.Product, parentCodes))
			stats.ItemsWithVariants++
			stats.VariantsEmitted += len(item.Variants)
			stats.ProductsEmitted++
			result.Items = append(result.Items, item)
		} else {
			stats.ProductsEmitted++
			result.Items = append(result.Items, entry.Item)
		}

		if options.MaxProducts > 0 && stats.ProductsEmitted >= options.MaxProducts {
			return result
		}
	}
	return result
}

func variantGroups(entries []productEntry) map[string][]productEntry {
	candidates := make(map[string][]productEntry)
	for _, entry := range entries {
		key := variantGroupKey(entry.Source)
		if key == "" {
			continue
		}
		candidates[key] = append(candidates[key], entry)
	}

	groups := make(map[string][]productEntry)
	for key, entries := range candidates {
		if len(entries) < 2 {
			continue
		}
		if _, ok := variantValues(entries); !ok {
			continue
		}
		groups[key] = entries
	}
	return groups
}

func variantGroupKey(source sourceItem) string {
	product := strings.TrimSpace(source.Product)
	category := strings.TrimSpace(source.MainCategory)
	if product == "" || category == "" {
		return ""
	}
	return strings.ToLower(product) + "\x00" + strings.ToLower(category)
}

type variantCandidate struct {
	Values []string
	Common string
	Score  int
}

func variantValues(entries []productEntry) ([]string, bool) {
	candidate, ok := bestVariantCandidate(entries)
	if !ok {
		return nil, false
	}
	return candidate.Values, true
}

func bestVariantCandidate(entries []productEntry) (variantCandidate, bool) {
	descriptions := make([]string, 0, len(entries))
	for _, entry := range entries {
		description := strings.TrimSpace(entry.Source.Description)
		if description == "" {
			return variantCandidate{}, false
		}
		descriptions = append(descriptions, description)
	}

	var candidates []variantCandidate
	if allDescriptionsContainComma(descriptions) {
		candidates = append(candidates, commaVariantCandidates(descriptions, len(entries))...)
	}
	if len(candidates) == 0 {
		if candidate, ok := wholeDescriptionVariantCandidate(descriptions); ok {
			candidates = append(candidates, candidate)
		}
	}
	if len(candidates) == 0 {
		return variantCandidate{}, false
	}

	best := candidates[0]
	for _, candidate := range candidates[1:] {
		if candidate.Score < best.Score {
			best = candidate
		}
	}
	return best, true
}

func allDescriptionsContainComma(descriptions []string) bool {
	for _, description := range descriptions {
		if !strings.Contains(description, ",") {
			return false
		}
	}
	return true
}

func commaVariantCandidates(descriptions []string, groupSize int) []variantCandidate {
	parts := make([][]string, 0, len(descriptions))
	maxParts := 0
	for _, description := range descriptions {
		descriptionParts := splitDescriptionParts(description)
		parts = append(parts, descriptionParts)
		if len(descriptionParts) > maxParts {
			maxParts = len(descriptionParts)
		}
	}

	var candidates []variantCandidate
	for splitIndex := 1; splitIndex < maxParts; splitIndex++ {
		var leftValues []string
		var rightValues []string
		validSplit := true
		for _, descriptionParts := range parts {
			if len(descriptionParts) <= splitIndex {
				validSplit = false
				break
			}
			leftValues = append(leftValues, strings.Join(descriptionParts[:splitIndex], ", "))
			rightValues = append(rightValues, strings.Join(descriptionParts[splitIndex:], ", "))
		}
		if !validSplit {
			continue
		}

		if variantValuesAreUnique(leftValues) && commonValuesAreStable(rightValues, groupSize) {
			candidates = append(candidates, newVariantCandidate(leftValues, rightValues))
		}
		if variantValuesAreUnique(rightValues) && commonValuesAreStable(leftValues, groupSize) {
			candidates = append(candidates, newVariantCandidate(rightValues, leftValues))
		}
	}
	return candidates
}

func splitDescriptionParts(description string) []string {
	rawParts := strings.Split(description, ",")
	parts := make([]string, 0, len(rawParts))
	for _, part := range rawParts {
		part = strings.TrimSpace(part)
		if part != "" {
			parts = append(parts, part)
		}
	}
	return parts
}

func wholeDescriptionVariantCandidate(descriptions []string) (variantCandidate, bool) {
	values := make([]string, 0, len(descriptions))
	for _, description := range descriptions {
		description = strings.TrimSpace(description)
		if strings.Contains(description, ",") || len([]rune(description)) > 32 || len(strings.Fields(description)) > 3 {
			return variantCandidate{}, false
		}
		values = append(values, description)
	}
	if !variantValuesAreUnique(values) {
		return variantCandidate{}, false
	}
	return newVariantCandidate(values, nil), true
}

func newVariantCandidate(values, commonValues []string) variantCandidate {
	return variantCandidate{
		Values: values,
		Common: firstNonEmptyValue(commonValues),
		Score:  averageRuneLength(values),
	}
}

func variantValuesAreUnique(values []string) bool {
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		key := normalizeKey(value)
		if key == "" {
			return false
		}
		if _, ok := seen[key]; ok {
			return false
		}
		seen[key] = struct{}{}
	}
	return true
}

func commonValuesAreStable(values []string, groupSize int) bool {
	if len(values) == 0 {
		return true
	}
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		key := normalizeKey(value)
		if key != "" {
			seen[key] = struct{}{}
		}
	}
	return len(seen) <= commonValueLimit(groupSize)
}

func commonValueLimit(groupSize int) int {
	if groupSize >= 4 {
		return 2
	}
	return 1
}

func normalizeKey(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	var builder strings.Builder
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r > 127 {
			builder.WriteRune(r)
		}
	}
	return builder.String()
}

func firstNonEmptyValue(values []string) string {
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			return value
		}
	}
	return ""
}

func averageRuneLength(values []string) int {
	if len(values) == 0 {
		return 0
	}
	total := 0
	for _, value := range values {
		total += len([]rune(strings.TrimSpace(value)))
	}
	return total / len(values)
}

func transformVariantGroup(entries []productEntry, parentCode string) shoptet.Item {
	first := entries[0].Item
	candidate, _ := bestVariantCandidate(entries)
	variants := make([]shoptet.Variant, 0, len(entries))
	for index, entry := range entries {
		variantValue := candidate.Values[index]
		variants = append(variants, shoptet.Variant{
			Code:         entry.Item.Code,
			PriceVAT:     entry.Item.PriceVAT,
			Stock:        entry.Item.Stock,
			Availability: entry.Item.Availability,
			ImageRef:     firstImageURL(entry.Item.Images),
			Parameters: []shoptet.Parameter{{
				Name:  "Provedení",
				Value: truncateVariantValue(variantValue),
			}},
		})
	}

	return shoptet.Item{
		Code:                  parentCode,
		Name:                  strings.TrimSpace(entries[0].Source.Product),
		Description:           candidate.Common,
		Manufacturer:          first.Manufacturer,
		Supplier:              first.Supplier,
		Categories:            first.Categories,
		DefaultCategory:       first.DefaultCategory,
		Images:                mergeImages(entries),
		InformationParameters: mergeInformationParameters(entries),
		Variants:              variants,
	}
}

func uniqueParentCode(product string, used map[string]struct{}) string {
	base := "HON-" + codeSlug(product)
	if base == "HON-" {
		base = "HON-PRODUKT"
	}
	code := base
	for index := 2; ; index++ {
		if _, ok := used[code]; !ok {
			used[code] = struct{}{}
			return code
		}
		code = base + "-" + strconv.Itoa(index)
	}
}

func codeSlug(value string) string {
	value = strings.ToUpper(strings.TrimSpace(value))
	var builder strings.Builder
	lastDash := false
	for _, r := range value {
		if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			builder.WriteRune(r)
			lastDash = false
			continue
		}
		if !lastDash && builder.Len() > 0 {
			builder.WriteByte('-')
			lastDash = true
		}
	}
	return strings.Trim(builder.String(), "-")
}

func truncateVariantValue(value string) string {
	value = strings.TrimSpace(value)
	if len([]rune(value)) <= 128 {
		return value
	}
	return string([]rune(value)[:128])
}

func firstImageURL(images []shoptet.Image) string {
	if len(images) == 0 {
		return ""
	}
	return strings.TrimSpace(images[0].URL)
}

func mergeImages(entries []productEntry) []shoptet.Image {
	var result []shoptet.Image
	seen := make(map[string]struct{})
	for _, entry := range entries {
		for _, image := range entry.Item.Images {
			url := strings.TrimSpace(image.URL)
			if url == "" {
				continue
			}
			if _, ok := seen[url]; ok {
				continue
			}
			seen[url] = struct{}{}
			result = append(result, shoptet.Image{URL: url})
		}
	}
	return result
}

func mergeInformationParameters(entries []productEntry) []shoptet.Parameter {
	var result []shoptet.Parameter
	seen := make(map[string]struct{})
	for _, entry := range entries {
		for _, parameter := range entry.Item.InformationParameters {
			name := strings.TrimSpace(parameter.Name)
			value := strings.TrimSpace(parameter.Value)
			if name == "" || value == "" {
				continue
			}
			key := strings.ToLower(name) + "\x00" + strings.ToLower(value)
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			result = append(result, shoptet.Parameter{Name: name, Value: value})
		}
	}
	return result
}

func transformName(product, description string) string {
	product = strings.TrimSpace(product)
	description = strings.TrimSpace(description)
	if product == "" {
		return description
	}
	if description == "" {
		return product
	}
	return product + " - " + description
}

func transformImages(urls []string) []shoptet.Image {
	result := make([]shoptet.Image, 0, len(urls))
	seen := make(map[string]struct{}, len(urls))
	for _, rawURL := range urls {
		url := strings.TrimSpace(rawURL)
		if url == "" {
			continue
		}
		if _, ok := seen[url]; ok {
			continue
		}
		seen[url] = struct{}{}
		result = append(result, shoptet.Image{URL: url})
	}
	return result
}

func transformParameters(parameters []sourceParameter) []shoptet.Parameter {
	if len(parameters) == 0 {
		return nil
	}

	result := make([]shoptet.Parameter, 0, len(parameters))
	seen := make(map[string]struct{}, len(parameters))
	for _, parameter := range parameters {
		name := strings.TrimSpace(parameter.Name)
		value := normalizeNumber(parameter.Value)
		if name == "" || value == "" {
			continue
		}
		key := strings.ToLower(name) + "\x00" + strings.ToLower(value)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, shoptet.Parameter{Name: name, Value: value})
	}
	return result
}

func transformCategory(mainCategory string) ([]shoptet.Category, *shoptet.Category) {
	name := strings.ToLower(strings.TrimSpace(mainCategory))
	var category shoptet.Category

	switch {
	case strings.Contains(name, "kancelářsk"):
		category = shoptet.Category{ID: "881", Path: "KANCELÁŘSKÉ ŽIDLE A KŘESLA"}
	case strings.Contains(name, "jednací"):
		category = shoptet.Category{ID: "1146", Path: "ŽIDLE > KONFERENČNÍ ŽIDLE"}
	case strings.Contains(name, "židle"):
		category = shoptet.Category{ID: "902", Path: "ŽIDLE"}
	default:
		category = shoptet.Category{ID: "1173", Path: "BYTOVÉ DOPLŇKY"}
	}

	return []shoptet.Category{category}, &category
}

func normalizeNumber(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return value
	}
	if parsed == float64(int64(parsed)) {
		return strconv.FormatInt(int64(parsed), 10)
	}
	return strconv.FormatFloat(parsed, 'f', -1, 64)
}

type sourceItem struct {
	ID           string            `xml:"ID"`
	MainCategory string            `xml:"MAIN_CATEGORY"`
	Product      string            `xml:"PRODUCT"`
	PriceVAT     string            `xml:"PRICE_VAT"`
	Availability string            `xml:"DOSTUPNOST"`
	Stock        string            `xml:"STOCK"`
	PartNumber   string            `xml:"PART_NUMBER"`
	Description  string            `xml:"DESCRIPTION"`
	Images       []string          `xml:"IMGURL>IMGURL"`
	Parameters   []sourceParameter `xml:"PARAM"`
}

type sourceParameter struct {
	Name  string `xml:"PARAM_NAME"`
	Value string `xml:"VAL"`
}

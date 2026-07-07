package sakypaky

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"html"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/fanda/arpari-xml-feed-parser-railway/internal/shoptet"
)

const ProductsURL = "https://www.sakypaky.cz/export/b2b_partners_cs.xml"

const (
	defaultVAT           = "21"
	maxVariantValueRunes = 128
	supplierName         = "Sakypaky"
	variantParameterName = "Barva"
)

var (
	beanBagCategory         = shoptet.Category{ID: "914", Path: "SEDACÍ VAKY"}
	footstoolCategory       = shoptet.Category{ID: "1155", Path: "ŽIDLE > TABURETY"}
	gardenAccessoryCategory = shoptet.Category{ID: "1227", Path: "ZAHRADNÍ NÁBYTEK > ZAHRADNÍ DOPLŃKY"}
	sideTableCategory       = shoptet.Category{ID: "1269", Path: "STOLY > ODKLÁDACÍ A PŘÍSTAVNÉ STOLKY"}
	homeAccessoryCategory   = shoptet.Category{ID: "1173", Path: "BYTOVÉ DOPLŇKY"}
)

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
		return nil, fmt.Errorf("create Sakypaky request: %w", err)
	}

	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("download Sakypaky feed: %w", err)
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		_ = response.Body.Close()
		return nil, fmt.Errorf("download Sakypaky feed: unexpected HTTP status %s", response.Status)
	}

	return response.Body, nil
}

type ProductsOptions struct {
	MaxProducts           int
	MaxVariantsPerProduct int
	PreferVariantItems    bool
}

type ProductsStats struct {
	ProductsRead      int
	ProductsEmitted   int
	ProductsSkipped   int
	ProductsTrimmed   int
	ItemsWithVariants int
	VariantsEmitted   int
	VariantsSkipped   int
	VariantsTrimmed   int
}

func ParseProducts(ctx context.Context, r io.Reader, options ProductsOptions) (shoptet.Feed, ProductsStats, error) {
	maxVariants := options.MaxVariantsPerProduct
	if maxVariants <= 0 {
		maxVariants = shoptet.DefaultMaxVariantsPerItem
	}

	decoder := xml.NewDecoder(r)
	var stats ProductsStats
	var entries []productEntry
	groups := make(map[string][]variantEntry)

	for {
		if err := ctx.Err(); err != nil {
			return shoptet.Feed{}, stats, err
		}

		token, err := decoder.Token()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return emitProducts(entries, groups, options, maxVariants, &stats), stats, nil
			}
			return shoptet.Feed{}, stats, fmt.Errorf("parse Sakypaky products XML: %w", err)
		}

		start, ok := token.(xml.StartElement)
		if !ok || start.Name.Local != "SHOPITEM" {
			continue
		}

		var source sourceItem
		if err := decoder.DecodeElement(&source, &start); err != nil {
			return shoptet.Feed{}, stats, fmt.Errorf("decode Sakypaky SHOPITEM: %w", err)
		}

		stats.ProductsRead++
		item, ok := transformProduct(source)
		if !ok {
			stats.ProductsSkipped++
			continue
		}

		groupID := strings.TrimSpace(source.ItemGroupID)
		if groupID == "" {
			entries = append(entries, productEntry{Item: item, HasItem: true})
			continue
		}

		if _, exists := groups[groupID]; !exists {
			entries = append(entries, productEntry{GroupID: groupID})
		}
		groups[groupID] = append(groups[groupID], variantEntry{
			GroupID:    groupID,
			Color:      source.Parameters.Value(variantParameterName),
			Category:   item.Categories[0],
			Item:       item,
			SourceName: item.Name,
		})
	}
}

func transformProduct(source sourceItem) (shoptet.Item, bool) {
	code := strings.TrimSpace(source.Code)
	name := strings.TrimSpace(source.ProductName)
	if code == "" || name == "" {
		slog.Warn("skipping Sakypaky product without required identity", "code", code, "name", name)
		return shoptet.Item{}, false
	}

	category, ok := targetCategory(source)
	if !ok {
		return shoptet.Item{}, false
	}

	priceVAT := shoptet.FormatWholePrice(source.PriceVAT)
	currency := ""
	vat := ""
	if priceVAT != "" {
		currency = "CZK"
		vat = defaultVAT
	}

	return shoptet.Item{
		Code:            code,
		Name:            name,
		Description:     normalizeText(source.Description),
		Manufacturer:    transformManufacturer(source.Manufacturer),
		Supplier:        supplierName,
		PriceVAT:        priceVAT,
		VAT:             vat,
		Currency:        currency,
		Availability:    transformDeliveryDate(source.DeliveryDate),
		EAN:             strings.TrimSpace(source.EAN),
		Categories:      []shoptet.Category{category},
		DefaultCategory: &category,
		Images:          transformImages(source),
	}, true
}

func emitProducts(entries []productEntry, groups map[string][]variantEntry, options ProductsOptions, maxVariants int, stats *ProductsStats) shoptet.Feed {
	var result shoptet.Feed
	for _, entry := range orderedEntries(entries, groups, options) {
		var item shoptet.Item
		var ok bool
		if entry.GroupID != "" {
			item, ok = transformVariantGroup(groups[entry.GroupID], maxVariants, stats)
		} else if entry.HasItem {
			item, ok = entry.Item, true
		}
		if !ok {
			stats.ProductsSkipped++
			continue
		}

		result.Items = append(result.Items, item)
		stats.ProductsEmitted++
		if len(item.Variants) > 0 {
			stats.ItemsWithVariants++
			stats.VariantsEmitted += len(item.Variants)
		}
		if options.MaxProducts > 0 && stats.ProductsEmitted >= options.MaxProducts {
			return result
		}
	}
	return result
}

func orderedEntries(entries []productEntry, groups map[string][]variantEntry, options ProductsOptions) []productEntry {
	if !options.PreferVariantItems {
		return entries
	}

	var variantsFirst []productEntry
	var simpleItems []productEntry
	for _, entry := range entries {
		if entry.GroupID != "" && len(groups[entry.GroupID]) > 0 {
			variantsFirst = append(variantsFirst, entry)
			continue
		}
		simpleItems = append(simpleItems, entry)
	}
	return append(variantsFirst, simpleItems...)
}

func transformVariantGroup(entries []variantEntry, maxVariants int, stats *ProductsStats) (shoptet.Item, bool) {
	if len(entries) == 0 {
		return shoptet.Item{}, false
	}

	displayColors, parentColors := variantColorPlan(entries)
	colorCounts := make(map[string]int, len(entries))
	for _, color := range displayColors {
		colorCounts[normalizeKey(color)]++
	}

	var variants []shoptet.Variant
	for index, entry := range entries {
		if len(variants) >= maxVariants {
			stats.VariantsTrimmed++
			continue
		}

		item := entry.Item
		variants = append(variants, shoptet.Variant{
			Code:         item.Code,
			EAN:          item.EAN,
			PriceVAT:     item.PriceVAT,
			VAT:          item.VAT,
			Currency:     item.Currency,
			Availability: item.Availability,
			ImageRef:     firstImageURL(item.Images),
			Parameters: []shoptet.Parameter{{
				Name:  variantParameterName,
				Value: variantValue(displayColors[index], item.Code, colorCounts),
			}},
		})
	}
	if len(variants) == 0 {
		return shoptet.Item{}, false
	}
	if len(entries) > len(variants) {
		stats.ProductsTrimmed++
		slog.Warn(
			"trimmed Sakypaky variants to Shoptet limit",
			"groupID", entries[0].GroupID,
			"kept", len(variants),
			"trimmed", len(entries)-len(variants),
		)
	}

	first := entries[0].Item
	category := entries[0].Category
	return shoptet.Item{
		Code:            "SAKYPAKY-" + entries[0].GroupID,
		Name:            parentName(entries, parentColors),
		Description:     first.Description,
		Manufacturer:    first.Manufacturer,
		Supplier:        supplierName,
		Categories:      []shoptet.Category{category},
		DefaultCategory: &category,
		Images:          mergeGroupImages(entries),
		Variants:        variants,
	}, true
}

type productEntry struct {
	Item    shoptet.Item
	HasItem bool
	GroupID string
}

type variantEntry struct {
	GroupID    string
	Color      string
	Category   shoptet.Category
	Item       shoptet.Item
	SourceName string
}

func targetCategory(source sourceItem) (shoptet.Category, bool) {
	name := strings.ToLower(normalizeText(source.ProductName))
	category := strings.ToLower(normalizeText(source.CategoryText))
	combined := name + " " + category

	if containsAny(combined, "pelech", "pro psy", "chovatel", "etiket", "jmenovk", "obalové materiály", "obalove materialy") {
		return shoptet.Category{}, false
	}

	switch {
	case containsAny(name, "houpač", "houpac", "houpací", "houpaci") || containsAny(category, "houpač", "houpac", "houpací", "houpaci"):
		return gardenAccessoryCategory, true
	case containsAny(name, "stolek", "stolku", "stolky", "skládací", "skladaci", "záklopová", "zaklopova"):
		return sideTableCategory, true
	case containsAny(name, "taburet") || categoryIsSpecificTaburet(category):
		return footstoolCategory, true
	case containsAny(name, "náplň", "napln", "oprav", "servis") || containsAny(category, "náhradní náplně", "nahradni naplne", "opravné sety", "opravne sety"):
		return beanBagCategory, true
	case containsAny(name, "sedací vak", "sedaci vak", "sedací pyt", "sedaci pyt") || containsAny(category, "sedací vaky", "sedaci vaky", "sedací pytle", "sedaci pytle"):
		return beanBagCategory, true
	case containsAny(name, "set") || containsAny(category, "výhodné sety", "vyhodne sety"):
		return beanBagCategory, true
	case containsAny(category, "ostatní produkty", "ostatni produkty"):
		return homeAccessoryCategory, true
	default:
		return shoptet.Category{}, false
	}
}

func categoryIsSpecificTaburet(category string) bool {
	category = strings.TrimSpace(category)
	return strings.HasSuffix(category, "taburety") || strings.Contains(category, "| taburety") || strings.Contains(category, "> taburety")
}

func containsAny(value string, needles ...string) bool {
	for _, needle := range needles {
		if strings.Contains(value, needle) {
			return true
		}
	}
	return false
}

func transformManufacturer(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return supplierName
	}
	return value
}

func normalizeText(value string) string {
	value = html.UnescapeString(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, "\u00a0", " ")
	return strings.Join(strings.Fields(value), " ")
}

func transformDeliveryDate(value string) string {
	value = strings.TrimSpace(value)
	if value == "0" {
		return "Skladem"
	}
	if value == "" {
		return ""
	}
	return "Dodání " + value + " dnů"
}

func transformImages(source sourceItem) []shoptet.Image {
	urls := make([]string, 0, 1+len(source.AlternativeImages))
	urls = append(urls, source.ImageURL)
	urls = append(urls, source.AlternativeImages...)

	images := make([]shoptet.Image, 0, len(urls))
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
		images = append(images, shoptet.Image{URL: url})
	}
	return images
}

func variantColor(entry variantEntry) string {
	if color := strings.TrimSpace(entry.Color); color != "" {
		return color
	}
	return strings.TrimSpace(entry.Item.Code)
}

func variantValue(color, code string, colorCounts map[string]int) string {
	color = strings.TrimSpace(color)
	if color == "" {
		return truncateVariantValue(code)
	}
	if colorCounts[normalizeKey(color)] <= 1 {
		return truncateVariantValue(color)
	}
	return truncateVariantValue(color + " (" + strings.TrimSpace(code) + ")")
}

func variantColorPlan(entries []variantEntry) ([]string, []string) {
	displayColors := make([]string, len(entries))
	parentColors := make([]string, len(entries))
	for index, entry := range entries {
		displayColors[index] = variantColor(entry)
		parentColors[index] = strings.TrimSpace(entry.Color)
	}
	if !isCoherentDusinkaVariantGroup(entries) {
		return displayColors, parentColors
	}

	for index, entry := range entries {
		suffix, _ := dusinkaVariantSuffix(entry.Color)
		displayColors[index] = suffix
		parentColors[index] = suffix
	}
	return displayColors, parentColors
}

func isCoherentDusinkaVariantGroup(entries []variantEntry) bool {
	if len(entries) < 2 {
		return false
	}

	var groupBase string
	for _, entry := range entries {
		suffix, ok := dusinkaVariantSuffix(entry.Color)
		if !ok {
			return false
		}

		base := trimTrailingValue(entry.SourceName, suffix)
		if base == entry.SourceName || !lastWordEqual(base, "Dušinka") {
			return false
		}
		if groupBase == "" {
			groupBase = base
			continue
		}
		if normalizeKey(groupBase) != normalizeKey(base) {
			return false
		}
	}
	return true
}

func dusinkaVariantSuffix(color string) (string, bool) {
	parts := strings.Fields(strings.TrimSpace(color))
	if len(parts) < 2 || !strings.EqualFold(parts[0], "Dušinka") {
		return "", false
	}
	return strings.Join(parts[1:], " "), true
}

func lastWordEqual(value, word string) bool {
	parts := strings.Fields(strings.TrimSpace(value))
	return len(parts) > 0 && strings.EqualFold(parts[len(parts)-1], word)
}

func parentName(entries []variantEntry, colors []string) string {
	names := make([][]string, 0, len(entries))
	for index, entry := range entries {
		color := strings.TrimSpace(entry.Color)
		if index < len(colors) {
			color = colors[index]
		}
		name := productBaseName(entry.SourceName, color)
		parts := strings.Fields(name)
		if len(parts) > 0 {
			names = append(names, parts)
		}
	}
	if len(names) == 0 {
		return entries[0].SourceName
	}

	common := append([]string(nil), names[0]...)
	for _, parts := range names[1:] {
		limit := min(len(common), len(parts))
		index := 0
		for index < limit && normalizeKey(common[index]) == normalizeKey(parts[index]) {
			index++
		}
		common = common[:index]
		if len(common) == 0 {
			break
		}
	}
	if len(common) > 0 {
		return strings.Join(common, " ")
	}
	return strings.Join(names[0], " ")
}

func productBaseName(name, color string) string {
	name = strings.TrimSpace(name)
	result := trimTrailingValue(name, color)
	if result == name {
		for _, suffix := range colorSuffixCandidates(color) {
			result = trimTrailingValue(name, suffix)
			if result != name {
				break
			}
		}
	}
	if result == "" {
		return name
	}
	return result
}

func colorSuffixCandidates(color string) []string {
	parts := strings.Fields(strings.NewReplacer("-", " ", "/", " ").Replace(strings.TrimSpace(color)))
	candidates := make([]string, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))
	for _, part := range parts {
		part = strings.Trim(part, " ,.-–")
		if len([]rune(part)) < 3 {
			continue
		}
		key := normalizeKey(part)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		candidates = append(candidates, part)
	}
	return candidates
}

func trimTrailingValue(value, suffix string) string {
	value = strings.TrimSpace(value)
	suffix = strings.TrimSpace(suffix)
	if value == "" || suffix == "" || !strings.HasSuffix(strings.ToLower(value), strings.ToLower(suffix)) {
		return value
	}
	trimmed := strings.TrimSpace(value[:len(value)-len(suffix)])
	return strings.TrimRight(trimmed, " ,-–")
}

func mergeGroupImages(entries []variantEntry) []shoptet.Image {
	var images []shoptet.Image
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
			images = append(images, shoptet.Image{URL: url})
		}
	}
	return images
}

func firstImageURL(images []shoptet.Image) string {
	if len(images) == 0 {
		return ""
	}
	return strings.TrimSpace(images[0].URL)
}

func truncateVariantValue(value string) string {
	value = strings.TrimSpace(value)
	runes := []rune(value)
	if len(runes) <= maxVariantValueRunes {
		return value
	}
	return string(runes[:maxVariantValueRunes])
}

func normalizeKey(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

type sourceItem struct {
	ProductName       string           `xml:"PRODUCTNAME"`
	Product           string           `xml:"PRODUCT"`
	Description       string           `xml:"DESCRIPTION"`
	ImageURL          string           `xml:"IMGURL"`
	AlternativeImages []string         `xml:"IMGURL_ALTERNATIVE"`
	PriceVAT          string           `xml:"PRICE_VAT"`
	Dues              string           `xml:"DUES"`
	DeliveryDate      string           `xml:"DELIVERY_DATE"`
	CategoryText      string           `xml:"CATEGORYTEXT"`
	Manufacturer      string           `xml:"MANUFACTURER"`
	Code              string           `xml:"CODE"`
	URL               string           `xml:"URL"`
	ItemID            string           `xml:"ITEM_ID"`
	ItemGroupID       string           `xml:"ITEMGROUP_ID"`
	EAN               string           `xml:"EAN"`
	Parameters        sourceParameters `xml:"PARAM"`
}

type sourceParameters []sourceParameter

func (parameters sourceParameters) Value(name string) string {
	for _, parameter := range parameters {
		if strings.EqualFold(strings.TrimSpace(parameter.Name), name) {
			return normalizeText(parameter.Value)
		}
	}
	return ""
}

type sourceParameter struct {
	Name  string `xml:"PARAM_NAME"`
	Value string `xml:"VAL"`
}

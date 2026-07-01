package autronic

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

const ProductsURL = "https://autronic.cz/feeds/product-feed.xml"

const (
	supplierName     = "Autronic"
	variantParameter = "Barva"
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
		return nil, fmt.Errorf("create Autronic request: %w", err)
	}

	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("download Autronic feed: %w", err)
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		_ = response.Body.Close()
		return nil, fmt.Errorf("download Autronic feed: unexpected HTTP status %s", response.Status)
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
			return shoptet.Feed{}, stats, fmt.Errorf("parse Autronic products XML: %w", err)
		}

		start, ok := token.(xml.StartElement)
		if !ok || start.Name.Local != "Product" {
			continue
		}

		var source sourceProduct
		if err := decoder.DecodeElement(&source, &start); err != nil {
			return shoptet.Feed{}, stats, fmt.Errorf("decode Autronic Product: %w", err)
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

func emitProducts(entries []productEntry, options ProductsOptions, stats *ProductsStats) shoptet.Feed {
	var result shoptet.Feed
	if len(entries) == 0 {
		return result
	}

	colorHints := variantColorHints(entries)
	groups := variantGroups(entries)
	emittedGroups := make(map[string]struct{}, len(groups.Items))
	for _, entry := range entries {
		root := groups.Find(entry.Item.Code)
		if _, ok := emittedGroups[root]; ok {
			continue
		}
		emittedGroups[root] = struct{}{}

		group := groups.Items[root]
		item := entry.Item
		if len(group) > 1 {
			item = transformVariantGroup(group, colorHints)
			stats.ItemsWithVariants++
			stats.VariantsEmitted += len(item.Variants)
		}

		stats.ProductsEmitted++
		result.Items = append(result.Items, item)
		if options.MaxProducts > 0 && stats.ProductsEmitted >= options.MaxProducts {
			return result
		}
	}
	return result
}

type productEntry struct {
	Source sourceProduct
	Item   shoptet.Item
}

func transformProduct(source sourceProduct) (shoptet.Item, bool) {
	code := strings.TrimSpace(source.Code)
	name := strings.TrimSpace(source.Name)
	categoryName := strings.TrimSpace(source.Category.Name)
	categoryShortName := strings.TrimSpace(source.Category.ShortName)

	if code == "" || name == "" {
		slog.Warn("skipping Autronic product without required identity", "code", code, "name", name)
		return shoptet.Item{}, false
	}
	if !isAllowedCategory(categoryShortName) {
		return shoptet.Item{}, false
	}

	categories, defaultCategory := transformCategory(categoryShortName, categoryName)
	stock, warehouses := transformAvailability(source.Availability)

	return shoptet.Item{
		Code:                  code,
		Name:                  name,
		Description:           transformDescription(source.Descriptions),
		Manufacturer:          transformManufacturer(source),
		Supplier:              supplierName,
		PriceVAT:              transformPrice(source.Prices),
		Stock:                 stock,
		Warehouses:            warehouses,
		EAN:                   strings.TrimSpace(source.EAN),
		Categories:            categories,
		DefaultCategory:       defaultCategory,
		Images:                transformImages(source.Images),
		InformationParameters: transformParameters(source.Parameters),
	}, true
}

type variantGroupSet struct {
	Parent map[string]string
	Items  map[string][]productEntry
}

func (groups variantGroupSet) Find(code string) string {
	if code == "" {
		return ""
	}
	parent, ok := groups.Parent[code]
	if !ok || parent == code {
		return code
	}
	root := groups.Find(parent)
	groups.Parent[code] = root
	return root
}

func variantGroups(entries []productEntry) variantGroupSet {
	groups := variantGroupSet{
		Parent: make(map[string]string, len(entries)),
		Items:  make(map[string][]productEntry, len(entries)),
	}
	for _, entry := range entries {
		code := strings.TrimSpace(entry.Item.Code)
		if code != "" {
			groups.Parent[code] = code
		}
	}

	union := func(left, right string) {
		left = strings.TrimSpace(left)
		right = strings.TrimSpace(right)
		if left == "" || right == "" {
			return
		}
		if _, ok := groups.Parent[left]; !ok {
			return
		}
		if _, ok := groups.Parent[right]; !ok {
			return
		}
		leftRoot := groups.Find(left)
		rightRoot := groups.Find(right)
		if leftRoot == rightRoot {
			return
		}
		if rightRoot < leftRoot {
			leftRoot, rightRoot = rightRoot, leftRoot
		}
		groups.Parent[rightRoot] = leftRoot
	}

	for _, entry := range entries {
		for _, variant := range entry.Source.ColorVariants {
			union(entry.Item.Code, variant.Code)
		}
	}

	for _, entry := range entries {
		root := groups.Find(entry.Item.Code)
		if root == "" {
			root = entry.Item.Code
		}
		groups.Items[root] = append(groups.Items[root], entry)
	}
	return groups
}

func transformVariantGroup(entries []productEntry, colorHints map[string]string) shoptet.Item {
	first := entries[0].Item
	colors := make(map[string]string, len(entries))
	colorCounts := make(map[string]int, len(entries))
	for _, entry := range entries {
		color := variantColor(entry.Source, colorHints)
		colors[entry.Item.Code] = color
		colorCounts[normalizeKey(color)]++
	}

	variants := make([]shoptet.Variant, 0, len(entries))
	for _, entry := range entries {
		color := colors[entry.Item.Code]
		variants = append(variants, shoptet.Variant{
			Code:       entry.Item.Code,
			EAN:        entry.Item.EAN,
			PriceVAT:   entry.Item.PriceVAT,
			Stock:      entry.Item.Stock,
			Warehouses: entry.Item.Warehouses,
			ImageRef:   firstImageURL(entry.Item.Images),
			Parameters: []shoptet.Parameter{{
				Name:  variantParameter,
				Value: variantValue(color, entry.Item.Code, colorCounts),
			}},
		})
	}

	return shoptet.Item{
		Name:                  parentName(entries, colors),
		Description:           first.Description,
		Manufacturer:          first.Manufacturer,
		Supplier:              first.Supplier,
		Categories:            first.Categories,
		DefaultCategory:       first.DefaultCategory,
		Images:                mergeGroupImages(entries),
		InformationParameters: mergeInformationParameters(entries),
		Variants:              variants,
	}
}

func variantColorHints(entries []productEntry) map[string]string {
	result := make(map[string]string, len(entries))
	for _, entry := range entries {
		for _, variant := range entry.Source.ColorVariants {
			code := strings.TrimSpace(variant.Code)
			color := strings.TrimSpace(variant.Color)
			if code == "" || color == "" {
				continue
			}
			if _, ok := result[code]; !ok {
				result[code] = color
			}
		}
	}
	return result
}

func variantColor(source sourceProduct, hints map[string]string) string {
	if color := sourceParameterValue(source.Parameters, variantParameter); color != "" {
		return color
	}
	if color := strings.TrimSpace(source.Color); color != "" {
		return color
	}
	if color := strings.TrimSpace(hints[strings.TrimSpace(source.Code)]); color != "" {
		return color
	}
	return strings.TrimSpace(source.Code)
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

func parentName(entries []productEntry, colors map[string]string) string {
	names := make([][]string, 0, len(entries))
	for _, entry := range entries {
		name := productBaseName(entry.Item.Name, entry.Item.Code, colors[entry.Item.Code])
		parts := splitNameParts(name)
		if len(parts) > 0 {
			names = append(names, parts)
		}
	}
	if len(names) == 0 {
		return entries[0].Item.Name
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
		return withGroupCode(strings.Join(common, ", "), variantGroupCode(entries))
	}
	return withGroupCode(strings.Join(names[0], ", "), variantGroupCode(entries))
}

func productBaseName(name, code, color string) string {
	result := trimTrailingCode(strings.TrimSpace(name), code)
	parts := splitNameParts(result)
	filtered := parts[:0]
	colorKey := normalizeKey(color)
	for _, part := range parts {
		if colorKey != "" && normalizeKey(part) == colorKey {
			continue
		}
		filtered = append(filtered, part)
	}
	if len(filtered) == 0 {
		return result
	}
	return strings.Join(filtered, ", ")
}

func withGroupCode(name, code string) string {
	name = strings.TrimSpace(name)
	code = strings.TrimSpace(code)
	if name == "" || code == "" || textContainsCode(name, code) {
		return name
	}
	return name + ", " + code
}

func textContainsCode(text, code string) bool {
	text = strings.ToLower(text)
	code = strings.ToLower(code)
	start := 0
	for {
		index := strings.Index(text[start:], code)
		if index < 0 {
			return false
		}
		index += start
		end := index + len(code)
		if textBoundary(text, index-1) && textBoundary(text, end) {
			return true
		}
		start = index + 1
	}
}

func textBoundary(value string, index int) bool {
	if index < 0 || index >= len(value) {
		return true
	}
	return !isASCIIAlnum(value[index])
}

func isASCIIAlnum(value byte) bool {
	return value >= 'a' && value <= 'z' || value >= '0' && value <= '9'
}

func variantGroupCode(entries []productEntry) string {
	codes := make([]string, 0, len(entries))
	for _, entry := range entries {
		if code := strings.TrimSpace(entry.Item.Code); code != "" {
			codes = append(codes, code)
		}
	}
	if len(codes) < 2 {
		return ""
	}

	prefix := codes[0]
	for _, code := range codes[1:] {
		prefix = commonCodePrefix(prefix, code)
		if prefix == "" {
			return ""
		}
	}

	prefix = trimCodeSeparators(prefix)
	for prefix != "" && !codePrefixBoundary(codes, prefix) {
		prefix = previousCodePrefix(prefix)
	}
	return prefix
}

func commonCodePrefix(left, right string) string {
	leftLower := strings.ToLower(left)
	rightLower := strings.ToLower(right)
	limit := min(len(leftLower), len(rightLower))
	index := 0
	for index < limit && leftLower[index] == rightLower[index] {
		index++
	}
	return left[:index]
}

func codePrefixBoundary(codes []string, prefix string) bool {
	for _, code := range codes {
		if len(code) == len(prefix) {
			continue
		}
		if len(code) < len(prefix) || !strings.EqualFold(code[:len(prefix)], prefix) || !isCodeSeparator(code[len(prefix)]) {
			return false
		}
	}
	return true
}

func previousCodePrefix(prefix string) string {
	prefix = trimCodeSeparators(prefix)
	for index := len(prefix) - 1; index >= 0; index-- {
		if isCodeSeparator(prefix[index]) {
			return trimCodeSeparators(prefix[:index])
		}
	}
	return ""
}

func trimCodeSeparators(value string) string {
	return strings.TrimRight(strings.TrimSpace(value), " -_/")
}

func isCodeSeparator(value byte) bool {
	return value == ' ' || value == '-' || value == '_' || value == '/'
}

func trimTrailingCode(name, code string) string {
	code = strings.TrimSpace(code)
	if name == "" || code == "" || !strings.HasSuffix(strings.ToLower(name), strings.ToLower(code)) {
		return name
	}
	return strings.TrimRight(strings.TrimSpace(name[:len(name)-len(code)]), " ,-")
}

func splitNameParts(name string) []string {
	rawParts := strings.Split(name, ",")
	parts := make([]string, 0, len(rawParts))
	for _, part := range rawParts {
		part = strings.TrimSpace(part)
		if part != "" {
			parts = append(parts, part)
		}
	}
	return parts
}

func mergeGroupImages(entries []productEntry) []shoptet.Image {
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
			if strings.EqualFold(strings.TrimSpace(parameter.Name), variantParameter) {
				continue
			}
			key := normalizeKey(parameter.Name) + "\x00" + normalizeKey(parameter.Value)
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			result = append(result, parameter)
		}
	}
	return result
}

func firstImageURL(images []shoptet.Image) string {
	if len(images) == 0 {
		return ""
	}
	return strings.TrimSpace(images[0].URL)
}

func sourceParameterValue(parameters []sourceParameter, name string) string {
	for _, parameter := range parameters {
		if strings.EqualFold(strings.TrimSpace(parameter.Name), name) {
			return transformParameterValue(parameter)
		}
	}
	return ""
}

func truncateVariantValue(value string) string {
	value = strings.TrimSpace(value)
	if len([]rune(value)) <= 128 {
		return value
	}
	return string([]rune(value)[:128])
}

func normalizeKey(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func isAllowedCategory(shortName string) bool {
	shortName = strings.ToUpper(strings.TrimSpace(shortName))
	if strings.HasPrefix(shortName, "NA-") {
		return true
	}

	switch shortName {
	case "BD-BO",
		"BD-STKV",
		"BD-NS",
		"BD-ODK",
		"BD-ORG",
		"BD-PAR",
		"BD-PO",
		"BD-REG",
		"BD-REG-KOV",
		"BD-REG-MAS",
		"BD-ST-SAT",
		"BD-TAB",
		"BD-TAB-UL",
		"BD-VES-KOV",
		"BD-VES-MAS":
		return true
	default:
		return false
	}
}

func transformPrice(prices sourcePrices) string {
	if value := strings.TrimSpace(prices.RetailPromotionalPriceIncludingVAT.Value); value != "" {
		return value
	}
	return strings.TrimSpace(prices.RetailPriceIncludingVAT.Value)
}

func transformAvailability(availability sourceAvailability) (string, []shoptet.Warehouse) {
	stock := normalizeNumber(availability.Total.Quantity)
	warehouses := make([]shoptet.Warehouse, 0, len(availability.Stocks))
	for _, sourceStock := range availability.Stocks {
		name := strings.TrimSpace(sourceStock.Name)
		value := normalizeNumber(sourceStock.Quantity)
		if name == "" && value == "" {
			continue
		}
		warehouses = append(warehouses, shoptet.Warehouse{Name: name, Value: value})
	}
	return stock, warehouses
}

func transformDescription(descriptions []sourceDescription) string {
	var fallback string
	for _, description := range descriptions {
		format := strings.ToLower(strings.TrimSpace(description.Format))
		value := strings.TrimSpace(description.Value)
		if value == "" {
			continue
		}
		if format == "html" {
			return value
		}
		if fallback == "" {
			fallback = value
		}
	}
	return fallback
}

func transformManufacturer(source sourceProduct) string {
	if brand := strings.TrimSpace(source.Brand); brand != "" {
		return brand
	}
	if manufacturer := strings.TrimSpace(source.Manufacturer); manufacturer != "" {
		return manufacturer
	}
	return supplierName
}

func transformImages(images []sourceImage) []shoptet.Image {
	result := make([]shoptet.Image, 0, len(images))
	seen := make(map[string]struct{}, len(images))
	for _, image := range images {
		url := strings.TrimSpace(image.LargeURL)
		if url == "" {
			url = strings.TrimSpace(image.MediumURL)
		}
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
		value := transformParameterValue(parameter)
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

func transformParameterValue(parameter sourceParameter) string {
	value := strings.TrimSpace(parameter.TextValue)
	if value == "" {
		value = normalizeNumber(parameter.NumericValue)
	}
	unit := strings.TrimSpace(parameter.Unit)
	if value == "" || unit == "" || parameterNameContainsUnit(parameter.Name, unit) {
		return value
	}
	return value + " " + unit
}

func parameterNameContainsUnit(name, unit string) bool {
	name = strings.ToLower(strings.TrimSpace(name))
	unit = strings.ToLower(strings.TrimSpace(unit))
	if name == "" || unit == "" {
		return false
	}
	return strings.Contains(name, "("+unit+")") || strings.Contains(name, " "+unit)
}

func transformCategory(categoryShortName, categoryName string) ([]shoptet.Category, *shoptet.Category) {
	if category, ok := exactCategory(strings.ToUpper(strings.TrimSpace(categoryShortName))); ok {
		return []shoptet.Category{category}, &category
	}

	name := strings.ToLower(strings.TrimSpace(categoryName))
	var category shoptet.Category

	switch {
	case strings.Contains(name, "síťovan") && strings.Contains(name, "kancelářsk"):
		category = shoptet.Category{ID: "1284", Path: "KANCELÁŘSKÉ ŽIDLE A KŘESLA > SÍŤOVANÉ KANCELÁŘSKÉ ŽIDLE"}
	case strings.Contains(name, "kancelářsk") && strings.Contains(name, "křes"):
		category = shoptet.Category{ID: "1275", Path: "KANCELÁŘSKÉ ŽIDLE A KŘESLA > ČALOUNĚNÁ KANCELÁŘSKÁ KŘESLA"}
	case strings.Contains(name, "kancelářsk") && strings.Contains(name, "židl"):
		category = shoptet.Category{ID: "881", Path: "KANCELÁŘSKÉ ŽIDLE A KŘESLA"}
	case strings.Contains(name, "barov") && strings.Contains(name, "židl"):
		category = shoptet.Category{ID: "1143", Path: "ŽIDLE > BAROVÉ ŽIDLE"}
	case strings.Contains(name, "čalouněn") && strings.Contains(name, "židl"):
		category = shoptet.Category{ID: "1134", Path: "ŽIDLE > ČALOUNĚNÉ ŽIDLE"}
	case strings.Contains(name, "zahradní") && strings.Contains(name, "židl"):
		category = shoptet.Category{ID: "1215", Path: "ZAHRADNÍ NÁBYTEK > ZAHRADNÍ ŽIDLE"}
	case strings.Contains(name, "židl"):
		category = shoptet.Category{ID: "902", Path: "ŽIDLE"}
	case strings.Contains(name, "jídelní") && strings.Contains(name, "stol"):
		category = shoptet.Category{ID: "974", Path: "STOLY > JÍDELNÍ STOLY"}
	case strings.Contains(name, "konferenční") && strings.Contains(name, "stol"):
		category = shoptet.Category{ID: "1263", Path: "STOLY > KONFERENČNÍ STOLY"}
	case strings.Contains(name, "odkládací") || strings.Contains(name, "přístav"):
		category = shoptet.Category{ID: "1269", Path: "STOLY > ODKLÁDACÍ A PŘÍSTAVNÉ STOLKY"}
	case strings.Contains(name, "zahradní") && strings.Contains(name, "stol"):
		category = shoptet.Category{ID: "1218", Path: "ZAHRADNÍ NÁBYTEK > ZAHRADNÍ STOLY"}
	case strings.Contains(name, "stol"):
		category = shoptet.Category{ID: "971", Path: "STOLY"}
	default:
		return nil, nil
	}

	return []shoptet.Category{category}, &category
}

func exactCategory(shortName string) (shoptet.Category, bool) {
	switch shortName {
	case "BD-BO":
		return shoptet.Category{ID: "1200", Path: "BYTOVÉ DOPLŇKY > BOTNÍKY"}, true
	case "BD-NS", "BD-ST-SAT":
		return shoptet.Category{ID: "1194", Path: "BYTOVÉ DOPLŇKY > NĚMÍ SLUHOVÉ"}, true
	case "BD-ODK":
		return shoptet.Category{ID: "1212", Path: "BYTOVÉ DOPLŇKY > ODPADKOVÉ KOŠE"}, true
	case "BD-ORG", "BD-PO":
		return shoptet.Category{ID: "1206", Path: "BYTOVÉ DOPLŇKY > POLIČKY"}, true
	case "BD-PAR":
		return shoptet.Category{ID: "1209", Path: "BYTOVÉ DOPLŇKY > PARAVÁNY"}, true
	case "BD-REG", "BD-REG-KOV", "BD-REG-MAS":
		return shoptet.Category{ID: "1203", Path: "BYTOVÉ DOPLŇKY > REGALY"}, true
	case "BD-STKV":
		return shoptet.Category{ID: "1227", Path: "ZAHRADNÍ NÁBYTEK > ZAHRADNÍ DOPLŃKY"}, true
	case "BD-TAB", "BD-TAB-UL":
		return shoptet.Category{ID: "1155", Path: "ŽIDLE > TABURETY"}, true
	case "BD-VES-KOV", "BD-VES-MAS":
		return shoptet.Category{ID: "1191", Path: "BYTOVÉ DOPLŇKY > VĚŠÁKY"}, true
	case "NA-KOM":
		return shoptet.Category{ID: "1197", Path: "BYTOVÉ DOPLŇKY > KOMODY"}, true
	case "NA-KRE-EL", "NA-KRE-HO", "NA-KRE-KT", "NA-KRE-POL":
		return shoptet.Category{ID: "944", Path: "SEDACÍ SOUPRAVY > KŘESLA"}, true
	case "NA-POH-PEV", "NA-POH-ROZ":
		return shoptet.Category{ID: "941", Path: "SEDACÍ SOUPRAVY > POHOVKY"}, true
	case "NA-POS-CAL":
		return shoptet.Category{ID: "1185", Path: "LOŽNICE > POSTELE"}, true
	case "NA-POS-KOV":
		return shoptet.Category{ID: "1185", Path: "LOŽNICE > POSTELE"}, true
	case "NA-SED-POL":
		return shoptet.Category{ID: "938", Path: "SEDACÍ SOUPRAVY > SEDACÍ SOUPRAVY"}, true
	case "NA-SET-2", "NA-SET-4":
		return shoptet.Category{ID: "1272", Path: "STOLY > JÍDELNÍ SETY"}, true
	case "NA-SKF-DYH", "NA-SKF-MAS", "NA-SKF-MO":
		return shoptet.Category{ID: "1263", Path: "STOLY > KONFERENČNÍ STOLY"}, true
	case "NA-SKF-OP":
		return shoptet.Category{ID: "1269", Path: "STOLY > ODKLÁDACÍ A PŘÍSTAVNÉ STOLKY"}, true
	case "NA-STO-MAS":
		return shoptet.Category{ID: "1245", Path: "STOLY > STOLY MASIV"}, true
	case "NA-STO-PRAC":
		return shoptet.Category{ID: "1257", Path: "STOLY > PRACOVNÍ STOLY"}, true
	case "NA-STO-DYH", "NA-STO-JID", "NA-STO-ROZ":
		return shoptet.Category{ID: "974", Path: "STOLY > JÍDELNÍ STOLY"}, true
	case "NA-ZAH-BAL", "NA-ZAH-JIDSET", "NA-ZAH-RELSET":
		return shoptet.Category{ID: "1221", Path: "ZAHRADNÍ NÁBYTEK > ZAHRADNÍ SESTAVY"}, true
	case "NA-ZAH-LEH":
		return shoptet.Category{ID: "1224", Path: "ZAHRADNÍ NÁBYTEK > ZAHRADNÍ LEHÁTKA"}, true
	case "NA-ZAH-STO":
		return shoptet.Category{ID: "1218", Path: "ZAHRADNÍ NÁBYTEK > ZAHRADNÍ STOLY"}, true
	case "NA-ZAH-ZID", "NA-ZAH-ZK":
		return shoptet.Category{ID: "1215", Path: "ZAHRADNÍ NÁBYTEK > ZAHRADNÍ ŽIDLE"}, true
	case "NA-ZAH", "NA-ZAH-BOX", "NA-ZAH-ST":
		return shoptet.Category{ID: "1227", Path: "ZAHRADNÍ NÁBYTEK > ZAHRADNÍ DOPLŃKY"}, true
	case "NA-ZID-BAR":
		return shoptet.Category{ID: "1143", Path: "ŽIDLE > BAROVÉ ŽIDLE"}, true
	case "NA-ZID-CAL":
		return shoptet.Category{ID: "1134", Path: "ŽIDLE > ČALOUNĚNÉ ŽIDLE"}, true
	case "NA-ZID-DR":
		return shoptet.Category{ID: "905", Path: "ŽIDLE > DŘEVĚNÉ ŽIDLE"}, true
	case "NA-ZID-PLAST":
		return shoptet.Category{ID: "911", Path: "ŽIDLE > PLASTOVÉ ŽIDLE"}, true
	case "NA-ZKA-CAL":
		return shoptet.Category{ID: "1275", Path: "KANCELÁŘSKÉ ŽIDLE A KŘESLA > ČALOUNĚNÁ KANCELÁŘSKÁ KŘESLA"}, true
	case "NA-ZKA-HER":
		return shoptet.Category{ID: "899", Path: "KANCELÁŘSKÉ ŽIDLE A KŘESLA > HERNÍ KŘESLA"}, true
	case "NA-ZKA-SIT":
		return shoptet.Category{ID: "1284", Path: "KANCELÁŘSKÉ ŽIDLE A KŘESLA > SÍŤOVANÉ KANCELÁŘSKÉ ŽIDLE"}, true
	default:
		return shoptet.Category{}, false
	}
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

type sourceProduct struct {
	Code          string               `xml:"ProductCode"`
	Name          string               `xml:"ProductName"`
	Brand         string               `xml:"Brand"`
	Manufacturer  string               `xml:"Manufacturer"`
	Color         string               `xml:"Color"`
	Category      sourceCategory       `xml:"ProductCategory"`
	EAN           string               `xml:"Ean"`
	Prices        sourcePrices         `xml:"Prices"`
	Availability  sourceAvailability   `xml:"Availability"`
	Descriptions  []sourceDescription  `xml:"Descriptions>Description"`
	Images        []sourceImage        `xml:"Images>Image"`
	Parameters    []sourceParameter    `xml:"Parameters>Parameter"`
	ColorVariants []sourceColorVariant `xml:"ColorVariants>Product"`
}

type sourceCategory struct {
	Name      string `xml:"CategoryName"`
	ShortName string `xml:"CategoryShortName"`
}

type sourcePrices struct {
	RetailPriceIncludingVAT            sourcePrice `xml:"RetailPriceIncludingVat"`
	RetailPromotionalPriceIncludingVAT sourcePrice `xml:"RetailPromotionalPriceIncludingVat"`
}

type sourcePrice struct {
	Currency string `xml:"currency,attr"`
	Value    string `xml:"value,attr"`
}

type sourceAvailability struct {
	Status string        `xml:"AvailabilityStatus"`
	Total  sourceTotal   `xml:"StockAvailabilityTotal"`
	Stocks []sourceStock `xml:"StockAvailability>Stock"`
}

type sourceTotal struct {
	Quantity string `xml:"Quantity,attr"`
}

type sourceStock struct {
	Name     string `xml:"Name,attr"`
	Quantity string `xml:"Quantity,attr"`
}

type sourceDescription struct {
	Format string `xml:"format,attr"`
	Value  string `xml:",chardata"`
}

type sourceImage struct {
	MediumURL string `xml:"mediumSizeUrl,attr"`
	LargeURL  string `xml:"largeSizeUrl,attr"`
}

type sourceParameter struct {
	Name         string `xml:"Name"`
	TextValue    string `xml:"TextValue"`
	NumericValue string `xml:"NumericValue"`
	Unit         string `xml:"Unit"`
}

type sourceColorVariant struct {
	Code  string `xml:"ProductCode"`
	Color string `xml:"Color"`
}

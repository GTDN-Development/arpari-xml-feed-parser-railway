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

const supplierName = "Autronic"

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
	ProductsRead    int
	ProductsEmitted int
	ProductsSkipped int
}

func ParseProducts(ctx context.Context, r io.Reader, options ProductsOptions) (shoptet.Feed, ProductsStats, error) {
	decoder := xml.NewDecoder(r)
	var result shoptet.Feed
	var stats ProductsStats

	for {
		if err := ctx.Err(); err != nil {
			return shoptet.Feed{}, stats, err
		}

		token, err := decoder.Token()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return result, stats, nil
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

		stats.ProductsEmitted++
		result.Items = append(result.Items, item)
		if options.MaxProducts > 0 && stats.ProductsEmitted >= options.MaxProducts {
			return result, stats, nil
		}
	}
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

func isAllowedCategory(shortName string) bool {
	shortName = strings.ToUpper(strings.TrimSpace(shortName))
	if strings.HasPrefix(shortName, "NA-") {
		return true
	}

	switch shortName {
	case "BD-BO",
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
		return shoptet.Category{ID: "1224", Path: "ZAHRADNÍ NÁBYTEK > ZAHRADNÍ LEHÁDKA"}, true
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
	Code         string              `xml:"ProductCode"`
	Name         string              `xml:"ProductName"`
	Category     sourceCategory      `xml:"ProductCategory"`
	EAN          string              `xml:"Ean"`
	Prices       sourcePrices        `xml:"Prices"`
	Availability sourceAvailability  `xml:"Availability"`
	Descriptions []sourceDescription `xml:"Descriptions>Description"`
	Images       []sourceImage       `xml:"Images>Image"`
	Parameters   []sourceParameter   `xml:"Parameters>Parameter"`
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

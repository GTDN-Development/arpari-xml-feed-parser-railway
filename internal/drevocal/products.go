package drevocal

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/fanda/arpari-xml-feed-parser-railway/internal/shoptet"
)

const ProductsURL = "https://www.matrace-drevocal.cz/feed-b2b.xml"

const (
	supplierName = "DŘEVOČAL"
	defaultVAT   = "21"
)

var (
	mattressCategory     = shoptet.Category{ID: "1188", Path: "LOŽNICE > MATRACE"}
	slattedFrameCategory = shoptet.Category{ID: "1281", Path: "LOŽNICE > ROŠTY"}
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
		return nil, fmt.Errorf("create Dřevočal request: %w", err)
	}

	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("download Dřevočal feed: %w", err)
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		_ = response.Body.Close()
		return nil, fmt.Errorf("download Dřevočal feed: unexpected HTTP status %s", response.Status)
	}

	return response.Body, nil
}

type ProductsOptions struct {
	MaxProducts           int
	MaxVariantsPerProduct int
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
	groups := make(map[string][]variantEntry)
	var groupOrder []string

	for {
		if err := ctx.Err(); err != nil {
			return shoptet.Feed{}, stats, err
		}

		token, err := decoder.Token()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return emitProducts(groups, groupOrder, options.MaxProducts, maxVariants, &stats), stats, nil
			}
			return shoptet.Feed{}, stats, fmt.Errorf("parse Dřevočal products XML: %w", err)
		}

		start, ok := token.(xml.StartElement)
		if !ok || start.Name.Local != "SHOPITEM" {
			continue
		}

		var source sourceItem
		if err := decoder.DecodeElement(&source, &start); err != nil {
			return shoptet.Feed{}, stats, fmt.Errorf("decode Dřevočal SHOPITEM: %w", err)
		}

		stats.ProductsRead++
		entry, ok := transformVariant(source)
		if !ok {
			stats.ProductsSkipped++
			stats.VariantsSkipped++
			continue
		}

		if _, exists := groups[entry.GroupID]; !exists {
			groupOrder = append(groupOrder, entry.GroupID)
		}
		groups[entry.GroupID] = append(groups[entry.GroupID], entry)
	}
}

type variantEntry struct {
	GroupID      string
	SourceName   string
	Category     shoptet.Category
	Manufacturer string
	Gift         string
	ImageURL     string
	Parameters   map[string]string
	Variant      shoptet.Variant
}

type productMapping struct {
	Category           shoptet.Category
	RequiredParameters []string
}

func transformVariant(source sourceItem) (variantEntry, bool) {
	code := strings.TrimSpace(source.ItemID)
	groupID := strings.TrimSpace(source.ItemGroupID)
	name := strings.TrimSpace(source.ProductName)
	if code == "" || groupID == "" || name == "" {
		slog.Warn("skipping Dřevočal variant without required identity", "code", code, "groupID", groupID, "name", name)
		return variantEntry{}, false
	}

	mapping, ok := productMappingForCategory(source.CategoryText)
	if !ok {
		return variantEntry{}, false
	}

	parameters := source.Parameters.Map()
	var variantParameters []shoptet.Parameter
	for _, parameterName := range mapping.RequiredParameters {
		value := strings.TrimSpace(parameters[parameterName])
		if value == "" {
			slog.Warn("skipping Dřevočal variant without required parameter", "code", code, "parameter", parameterName)
			return variantEntry{}, false
		}
		variantParameters = append(variantParameters, shoptet.Parameter{Name: parameterName, Value: value})
	}

	currency := strings.TrimSpace(source.Currency)
	priceVAT := strings.TrimSpace(source.PriceVAT)
	vat := ""
	if priceVAT != "" {
		vat = defaultVAT
		if currency == "" {
			currency = "CZK"
		}
	}

	return variantEntry{
		GroupID:      groupID,
		SourceName:   name,
		Category:     mapping.Category,
		Manufacturer: transformManufacturer(source.Manufacturer),
		Gift:         strings.TrimSpace(source.Gift),
		ImageURL:     strings.TrimSpace(source.ImageURL),
		Parameters:   parameters,
		Variant: shoptet.Variant{
			Code:         code,
			EAN:          strings.TrimSpace(source.EAN),
			PriceVAT:     priceVAT,
			VAT:          vat,
			Currency:     currency,
			Availability: strings.TrimSpace(source.Availability),
			ImageRef:     strings.TrimSpace(source.ImageURL),
			Parameters:   variantParameters,
		},
	}, true
}

func emitProducts(groups map[string][]variantEntry, groupOrder []string, maxProducts, maxVariants int, stats *ProductsStats) shoptet.Feed {
	var result shoptet.Feed
	for _, groupID := range groupOrder {
		entries := groups[groupID]
		if len(entries) == 0 {
			continue
		}

		variants := make([]shoptet.Variant, 0, min(len(entries), maxVariants))
		for _, entry := range entries {
			if len(variants) >= maxVariants {
				stats.VariantsTrimmed++
				continue
			}
			variants = append(variants, entry.Variant)
		}
		if len(variants) == 0 {
			stats.ProductsSkipped++
			continue
		}
		if len(entries) > len(variants) {
			stats.ProductsTrimmed++
			slog.Warn(
				"trimmed Dřevočal variants to Shoptet limit",
				"groupID", groupID,
				"kept", len(variants),
				"trimmed", len(entries)-len(variants),
			)
		}

		first := entries[0]
		category := first.Category
		item := shoptet.Item{
			Code:                  "DREVOCAL-" + groupID,
			Name:                  parentName(first),
			Manufacturer:          first.Manufacturer,
			Supplier:              supplierName,
			Categories:            []shoptet.Category{category},
			DefaultCategory:       &category,
			Images:                groupImages(entries),
			InformationParameters: giftInformationParameters(entries),
			Variants:              variants,
		}

		result.Items = append(result.Items, item)
		stats.ProductsEmitted++
		stats.ItemsWithVariants++
		stats.VariantsEmitted += len(variants)
		if maxProducts > 0 && stats.ProductsEmitted >= maxProducts {
			return result
		}
	}
	return result
}

func productMappingForCategory(categoryText string) (productMapping, bool) {
	switch strings.TrimSpace(categoryText) {
	case "Lamelové rošty":
		return productMapping{
			Category:           slattedFrameCategory,
			RequiredParameters: []string{"Rozměr"},
		}, true
	case "", "Matrace", "Doplňky":
		// Keep Doplňky on the legacy mattress-compatible path so adding rošty does not change existing output.
		return productMapping{
			Category:           mattressCategory,
			RequiredParameters: []string{"Rozměr", "Výška", "Potah"},
		}, true
	default:
		return productMapping{}, false
	}
}

func transformManufacturer(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return supplierName
	}
	return value
}

func parentName(entry variantEntry) string {
	name := strings.TrimSpace(entry.SourceName)
	if name == "" {
		return entry.GroupID
	}

	dimension := strings.TrimSpace(entry.Parameters["Rozměr"])
	height := normalizeHeightForName(entry.Parameters["Výška"])
	cover := strings.TrimSpace(entry.Parameters["Potah"])

	result := trimTrailingValue(name, cover)
	if dimension != "" && height != "" {
		result = trimTrailingValue(result, dimension+"x"+height)
		result = trimTrailingValue(result, dimension+"×"+height)
	}
	if height != "" {
		result = trimTrailingValue(result, height+" cm")
		result = trimTrailingValue(result, height)
	}
	result = trimTrailingValue(result, dimension)

	result = strings.Join(strings.Fields(result), " ")
	if result == "" {
		return name
	}
	return result
}

func normalizeHeightForName(value string) string {
	value = strings.TrimSpace(value)
	value = strings.TrimSuffix(value, "cm")
	value = strings.TrimSuffix(value, "CM")
	return strings.TrimSpace(value)
}

func trimTrailingValue(value, suffix string) string {
	value = strings.TrimSpace(value)
	suffix = strings.TrimSpace(suffix)
	if value == "" || suffix == "" {
		return value
	}
	if !strings.HasSuffix(strings.ToLower(value), strings.ToLower(suffix)) {
		return value
	}
	trimmed := strings.TrimSpace(value[:len(value)-len(suffix)])
	return strings.TrimRight(trimmed, " ,-–")
}

func groupImages(entries []variantEntry) []shoptet.Image {
	images := make([]shoptet.Image, 0, 1)
	seen := make(map[string]struct{}, len(entries))
	for _, entry := range entries {
		url := strings.TrimSpace(entry.ImageURL)
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

func giftInformationParameters(entries []variantEntry) []shoptet.Parameter {
	seen := make(map[string]struct{}, len(entries))
	var result []shoptet.Parameter
	for _, entry := range entries {
		gift := strings.TrimSpace(entry.Gift)
		if gift == "" {
			continue
		}
		if _, ok := seen[gift]; ok {
			continue
		}
		seen[gift] = struct{}{}
		result = append(result, shoptet.Parameter{Name: "Dárek", Value: gift})
	}
	return result
}

type sourceItem struct {
	ItemID       string           `xml:"ITEM_ID"`
	ItemGroupID  string           `xml:"ITEMGROUP_ID"`
	ProductName  string           `xml:"PRODUCTNAME"`
	Manufacturer string           `xml:"MANUFACTURER"`
	PriceVAT     string           `xml:"PRICE_VAT"`
	Currency     string           `xml:"CURRENCY"`
	EAN          string           `xml:"EAN"`
	CategoryText string           `xml:"CATEGORYTEXT"`
	URL          string           `xml:"URL"`
	ImageURL     string           `xml:"IMGURL"`
	Availability string           `xml:"AVAILABILITY"`
	Gift         string           `xml:"GIFT"`
	Parameters   sourceParameters `xml:"PARAM"`
}

type sourceParameters []sourceParameter

func (parameters sourceParameters) Map() map[string]string {
	result := make(map[string]string, len(parameters))
	for _, parameter := range parameters {
		name := strings.TrimSpace(parameter.Name)
		value := strings.TrimSpace(parameter.Value)
		if name == "" || value == "" {
			continue
		}
		if _, exists := result[name]; !exists {
			result[name] = value
		}
	}
	return result
}

type sourceParameter struct {
	Name  string `xml:"PARAM_NAME"`
	Value string `xml:"VAL"`
}

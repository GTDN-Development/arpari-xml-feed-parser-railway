package stima

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

const (
	ProductsURL   = "https://www.stima.cz/userfiles/xml/ITTC_SHT_products.xml"
	StockURL      = "https://www.stima.cz/userfiles/xml/ITTC_SHT_stock.xml"
	StockPriceURL = "https://www.stima.cz/userfiles/xml/ITTC_SHT_stock_price.xml"
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
		return nil, fmt.Errorf("create STIMA request: %w", err)
	}

	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("download STIMA feed: %w", err)
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		_ = response.Body.Close()
		return nil, fmt.Errorf("download STIMA feed: unexpected HTTP status %s", response.Status)
	}

	return response.Body, nil
}

type ProductsOptions struct {
	MaxVariantsPerProduct int
	MaxProducts           int
}

type ProductsStats struct {
	ProductsRead      int
	ProductsEmitted   int
	ProductsSkipped   int
	ProductsTrimmed   int
	VariantsEmitted   int
	VariantsSkipped   int
	VariantsTrimmed   int
	ItemsWithVariants int
}

func ParseProducts(ctx context.Context, r io.Reader, options ProductsOptions) (shoptet.Feed, ProductsStats, error) {
	maxVariants := options.MaxVariantsPerProduct
	if maxVariants <= 0 {
		maxVariants = shoptet.DefaultMaxVariantsPerItem
	}

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
			return shoptet.Feed{}, stats, fmt.Errorf("parse STIMA products XML: %w", err)
		}

		start, ok := token.(xml.StartElement)
		if !ok || start.Name.Local != "SHOPITEM" {
			continue
		}

		var source sourceShopItem
		if err := decoder.DecodeElement(&source, &start); err != nil {
			return shoptet.Feed{}, stats, fmt.Errorf("decode STIMA SHOPITEM: %w", err)
		}

		stats.ProductsRead++
		item, itemStats, ok := transformProduct(source, maxVariants)
		stats.ProductsSkipped += itemStats.ProductsSkipped
		stats.ProductsTrimmed += itemStats.ProductsTrimmed
		stats.VariantsEmitted += itemStats.VariantsEmitted
		stats.VariantsSkipped += itemStats.VariantsSkipped
		stats.VariantsTrimmed += itemStats.VariantsTrimmed
		if !ok {
			continue
		}

		if len(item.Variants) > 0 {
			stats.ItemsWithVariants++
		}
		stats.ProductsEmitted++
		result.Items = append(result.Items, item)
		if options.MaxProducts > 0 && stats.ProductsEmitted >= options.MaxProducts {
			return result, stats, nil
		}
	}
}

type productTransformStats struct {
	ProductsSkipped int
	ProductsTrimmed int
	VariantsEmitted int
	VariantsSkipped int
	VariantsTrimmed int
}

func transformProduct(source sourceShopItem, maxVariants int) (shoptet.Item, productTransformStats, bool) {
	var stats productTransformStats
	name := strings.TrimSpace(source.Name)

	if len(source.Variants) == 0 {
		code := strings.TrimSpace(source.Code)
		if code == "" {
			stats.ProductsSkipped = 1
			slog.Warn("skipping STIMA simple product without CODE", "name", name)
			return shoptet.Item{}, stats, false
		}

		stock, warehouses := transformStock(source.Stock)
		categories, defaultCategory := transformCategories(source.Categories, name)
		if len(categories) == 0 {
			stats.ProductsSkipped = 1
			slog.Warn("skipping STIMA simple product without mapped category", "name", name, "code", code)
			return shoptet.Item{}, stats, false
		}
		return shoptet.Item{
			Code:            code,
			Name:            name,
			EAN:             strings.TrimSpace(source.EAN),
			PriceVAT:        strings.TrimSpace(source.PriceVAT),
			Stock:           stock,
			Warehouses:      warehouses,
			Categories:      categories,
			DefaultCategory: defaultCategory,
			Images:          transformImages(source.Images),
		}, stats, true
	}

	variants := make([]shoptet.Variant, 0, min(len(source.Variants), maxVariants))
	var firstValidVariantCode string
	for variantIndex, sourceVariant := range source.Variants {
		code := strings.TrimSpace(sourceVariant.Code)
		if code == "" {
			stats.VariantsSkipped++
			slog.Warn("skipping STIMA variant without CODE", "name", name, "variantIndex", variantIndex)
			continue
		}
		if firstValidVariantCode == "" {
			firstValidVariantCode = code
		}
		if len(variants) >= maxVariants {
			stats.VariantsTrimmed++
			continue
		}

		stock, warehouses := transformStock(sourceVariant.Stock)
		variants = append(variants, shoptet.Variant{
			Code:       code,
			EAN:        strings.TrimSpace(sourceVariant.EAN),
			PriceVAT:   strings.TrimSpace(sourceVariant.PriceVAT),
			Stock:      stock,
			Warehouses: warehouses,
			Parameters: transformParameters(sourceVariant.Parameters),
		})
	}

	if len(variants) == 0 {
		stats.ProductsSkipped = 1
		slog.Warn("skipping STIMA variant product without usable variants", "name", name)
		return shoptet.Item{}, stats, false
	}
	if stats.VariantsTrimmed > 0 {
		stats.ProductsTrimmed = 1
		slog.Warn(
			"trimmed STIMA product variants to Shoptet limit",
			"name", name,
			"code", parentCode(source.Code, firstValidVariantCode),
			"kept", len(variants),
			"trimmed", stats.VariantsTrimmed,
		)
	}

	categories, defaultCategory := transformCategories(source.Categories, name)
	if len(categories) == 0 {
		stats.ProductsSkipped = 1
		slog.Warn("skipping STIMA variant product without mapped category", "name", name, "code", parentCode(source.Code, firstValidVariantCode))
		return shoptet.Item{}, stats, false
	}
	stats.VariantsEmitted = len(variants)
	return shoptet.Item{
		Code:            parentCode(source.Code, firstValidVariantCode),
		Name:            name,
		EAN:             strings.TrimSpace(source.EAN),
		PriceVAT:        strings.TrimSpace(source.PriceVAT),
		Categories:      categories,
		DefaultCategory: defaultCategory,
		Images:          transformImages(source.Images),
		Variants:        variants,
	}, stats, true
}

func parentCode(sourceCode, firstVariantCode string) string {
	if code := strings.TrimSpace(sourceCode); code != "" {
		return code
	}

	code := strings.TrimSpace(firstVariantCode)
	if before, _, ok := strings.Cut(code, "-"); ok && strings.TrimSpace(before) != "" {
		return strings.TrimSpace(before)
	}
	return code
}

func transformStock(stock sourceStock) (string, []shoptet.Warehouse) {
	stockValue := strings.TrimSpace(stock.Value)
	if len(stock.Warehouses) == 0 {
		return stockValue, nil
	}

	warehouses := make([]shoptet.Warehouse, 0, len(stock.Warehouses))
	for _, warehouse := range stock.Warehouses {
		name := strings.TrimSpace(warehouse.Name)
		value := strings.TrimSpace(warehouse.Value)
		if name == "" && value == "" {
			continue
		}
		warehouses = append(warehouses, shoptet.Warehouse{
			Name:  name,
			Value: value,
		})
	}
	return stockValue, warehouses
}

func transformParameters(parameters []sourceParameter) []shoptet.Parameter {
	if len(parameters) == 0 {
		return nil
	}

	result := make([]shoptet.Parameter, 0, len(parameters))
	for _, parameter := range parameters {
		name := strings.TrimSpace(parameter.Name)
		value := strings.TrimSpace(parameter.Value)
		if !isAllowedVariantParameter(name) || value == "" {
			continue
		}
		result = append(result, shoptet.Parameter{
			Name:  name,
			Value: value,
		})
	}
	return result
}

func transformImages(images []string) []shoptet.Image {
	if len(images) == 0 {
		return nil
	}

	result := make([]shoptet.Image, 0, len(images))
	seen := make(map[string]struct{}, len(images))
	for _, image := range images {
		url := strings.TrimSpace(image)
		if url == "" {
			continue
		}
		if _, exists := seen[url]; exists {
			continue
		}
		seen[url] = struct{}{}
		result = append(result, shoptet.Image{URL: url})
	}
	return result
}

func isAllowedVariantParameter(name string) bool {
	switch name {
	case "KOSTRA", "Sedák", "Délka stolu", "Rozklad":
		return true
	default:
		return false
	}
}

type sourceShopItem struct {
	Name       string           `xml:"NAME"`
	Code       string           `xml:"CODE"`
	EAN        string           `xml:"EAN"`
	PriceVAT   string           `xml:"PRICE_VAT"`
	Stock      sourceStock      `xml:"STOCK"`
	Categories sourceCategories `xml:"CATEGORIES"`
	Images     []string         `xml:"IMAGES>IMAGE"`
	Variants   []sourceVariant  `xml:"VARIANTS>VARIANT"`
}

type sourceVariant struct {
	Code       string            `xml:"CODE"`
	EAN        string            `xml:"EAN"`
	PriceVAT   string            `xml:"PRICE_VAT"`
	Stock      sourceStock       `xml:"STOCK"`
	Parameters []sourceParameter `xml:"PARAMETERS>PARAMETER"`
}

type sourceStock struct {
	Value      string            `xml:",chardata"`
	Warehouses []sourceWarehouse `xml:"WAREHOUSES>WAREHOUSE"`
}

type sourceWarehouse struct {
	Name  string `xml:"NAME"`
	Value string `xml:"VALUE"`
}

type sourceParameter struct {
	Name  string `xml:"NAME"`
	Value string `xml:"VALUE"`
}

type sourceCategories struct {
	Items   []string `xml:"CATEGORY"`
	Default string   `xml:"DEFAULT_CATEGORY"`
}

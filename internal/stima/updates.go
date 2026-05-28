package stima

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/fanda/arpari-xml-feed-parser-railway/internal/shoptet"
)

type UpdateOptions struct {
	MaxVariantsPerProduct int
}

type UpdateStats struct {
	ProductsRead      int
	ProductsEmitted   int
	ProductsSkipped   int
	ProductsTrimmed   int
	VariantsEmitted   int
	VariantsSkipped   int
	VariantsTrimmed   int
	ItemsWithVariants int
}

func ParseStock(ctx context.Context, r io.Reader, options UpdateOptions) (shoptet.Feed, UpdateStats, error) {
	return parseUpdate(ctx, r, options, updateMode{IncludePrice: false, Name: "stock"})
}

func ParseStockPrice(ctx context.Context, r io.Reader, options UpdateOptions) (shoptet.Feed, UpdateStats, error) {
	return parseUpdate(ctx, r, options, updateMode{IncludePrice: true, Name: "stock-price"})
}

type updateMode struct {
	IncludePrice bool
	Name         string
}

func parseUpdate(ctx context.Context, r io.Reader, options UpdateOptions, mode updateMode) (shoptet.Feed, UpdateStats, error) {
	maxVariants := options.MaxVariantsPerProduct
	if maxVariants <= 0 {
		maxVariants = shoptet.DefaultMaxVariantsPerItem
	}

	decoder := xml.NewDecoder(r)
	var result shoptet.Feed
	var stats UpdateStats

	for {
		if err := ctx.Err(); err != nil {
			return shoptet.Feed{}, stats, err
		}

		token, err := decoder.Token()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return result, stats, nil
			}
			return shoptet.Feed{}, stats, fmt.Errorf("parse STIMA %s XML: %w", mode.Name, err)
		}

		start, ok := token.(xml.StartElement)
		if !ok || start.Name.Local != "SHOPITEM" {
			continue
		}

		var source sourceShopItem
		if err := decoder.DecodeElement(&source, &start); err != nil {
			return shoptet.Feed{}, stats, fmt.Errorf("decode STIMA %s SHOPITEM: %w", mode.Name, err)
		}

		stats.ProductsRead++
		item, itemStats, ok := transformUpdate(source, maxVariants, mode)
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
	}
}

func transformUpdate(source sourceShopItem, maxVariants int, mode updateMode) (shoptet.Item, productTransformStats, bool) {
	var stats productTransformStats

	if len(source.Variants) == 0 {
		code := strings.TrimSpace(source.Code)
		if code == "" {
			stats.ProductsSkipped = 1
			slog.Warn("skipping STIMA update product without CODE", "feed", mode.Name)
			return shoptet.Item{}, stats, false
		}
		if !hasUpdatePayload(source.Stock, source.PriceVAT, mode) {
			stats.ProductsSkipped = 1
			slog.Warn("skipping STIMA update product without update payload", "feed", mode.Name, "code", code)
			return shoptet.Item{}, stats, false
		}

		stock, warehouses := transformStock(source.Stock)
		item := shoptet.Item{
			Code:       code,
			Stock:      stock,
			Warehouses: warehouses,
		}
		if mode.IncludePrice {
			item.PriceVAT = strings.TrimSpace(source.PriceVAT)
		}
		return item, stats, true
	}

	variants := make([]shoptet.Variant, 0, min(len(source.Variants), maxVariants))
	var firstValidVariantCode string
	for variantIndex, sourceVariant := range source.Variants {
		code := strings.TrimSpace(sourceVariant.Code)
		if code == "" {
			stats.VariantsSkipped++
			slog.Warn("skipping STIMA update variant without CODE", "feed", mode.Name, "variantIndex", variantIndex)
			continue
		}
		if !hasUpdatePayload(sourceVariant.Stock, sourceVariant.PriceVAT, mode) {
			stats.VariantsSkipped++
			slog.Warn("skipping STIMA update variant without update payload", "feed", mode.Name, "code", code)
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
		variant := shoptet.Variant{
			Code:       code,
			Stock:      stock,
			Warehouses: warehouses,
			Parameters: transformParameters(sourceVariant.Parameters),
		}
		if mode.IncludePrice {
			variant.PriceVAT = strings.TrimSpace(sourceVariant.PriceVAT)
		}
		variants = append(variants, variant)
	}

	if len(variants) == 0 {
		stats.ProductsSkipped = 1
		slog.Warn("skipping STIMA update product without usable variants", "feed", mode.Name)
		return shoptet.Item{}, stats, false
	}
	if stats.VariantsTrimmed > 0 {
		stats.ProductsTrimmed = 1
		slog.Warn(
			"trimmed STIMA update product variants to Shoptet limit",
			"feed", mode.Name,
			"code", parentCode(source.Code, firstValidVariantCode),
			"kept", len(variants),
			"trimmed", stats.VariantsTrimmed,
		)
	}

	stats.VariantsEmitted = len(variants)
	return shoptet.Item{
		Code:     parentCode(source.Code, firstValidVariantCode),
		Variants: variants,
	}, stats, true
}

func hasUpdatePayload(stock sourceStock, priceVAT string, mode updateMode) bool {
	if hasStockPayload(stock) {
		return true
	}
	return mode.IncludePrice && strings.TrimSpace(priceVAT) != ""
}

func hasStockPayload(stock sourceStock) bool {
	if strings.TrimSpace(stock.Value) != "" {
		return true
	}
	for _, warehouse := range stock.Warehouses {
		if strings.TrimSpace(warehouse.Name) != "" || strings.TrimSpace(warehouse.Value) != "" {
			return true
		}
	}
	return false
}

package autronic

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

const AvailabilityURL = "https://autronic.cz/feeds/availability-feed.xml"

type AvailabilityOptions struct{}

type AvailabilityStats struct {
	CatalogProductsRead     int
	CatalogProductsEmitted  int
	CatalogProductsSkipped  int
	ProductsRead            int
	ProductsEmitted         int
	ProductsSkipped         int
	ItemsWithVariants       int
	VariantsEmitted         int
	AvailabilityRecordsUsed int
}

type availabilityRecord struct {
	Stock      string
	Warehouses []shoptet.Warehouse
}

func ParseAvailability(ctx context.Context, catalogReader, availabilityReader io.Reader, options AvailabilityOptions) (shoptet.Feed, AvailabilityStats, error) {
	_ = options

	catalogFeed, catalogStats, err := ParseProducts(ctx, catalogReader, ProductsOptions{})
	if err != nil {
		return shoptet.Feed{}, AvailabilityStats{}, err
	}

	availabilityByCode, stats, err := parseAvailabilityRecords(ctx, availabilityReader)
	if err != nil {
		return shoptet.Feed{}, stats, err
	}
	stats.CatalogProductsRead = catalogStats.ProductsRead
	stats.CatalogProductsEmitted = catalogStats.ProductsEmitted
	stats.CatalogProductsSkipped = catalogStats.ProductsSkipped

	result, transformStats := transformAvailabilityUpdate(catalogFeed, availabilityByCode)
	stats.ProductsEmitted = transformStats.ProductsEmitted
	stats.ItemsWithVariants = transformStats.ItemsWithVariants
	stats.VariantsEmitted = transformStats.VariantsEmitted
	stats.AvailabilityRecordsUsed = transformStats.AvailabilityRecordsUsed
	stats.ProductsSkipped += len(availabilityByCode) - transformStats.AvailabilityRecordsUsed

	return result, stats, nil
}

func parseAvailabilityRecords(ctx context.Context, r io.Reader) (map[string]availabilityRecord, AvailabilityStats, error) {
	decoder := xml.NewDecoder(r)
	result := make(map[string]availabilityRecord)
	var stats AvailabilityStats

	for {
		if err := ctx.Err(); err != nil {
			return nil, stats, err
		}

		token, err := decoder.Token()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return result, stats, nil
			}
			return nil, stats, fmt.Errorf("parse Autronic availability XML: %w", err)
		}

		start, ok := token.(xml.StartElement)
		if !ok || start.Name.Local != "Product" {
			continue
		}

		var source sourceAvailabilityProduct
		if err := decoder.DecodeElement(&source, &start); err != nil {
			return nil, stats, fmt.Errorf("decode Autronic availability Product: %w", err)
		}

		stats.ProductsRead++
		code := strings.TrimSpace(source.Code)
		if code == "" {
			stats.ProductsSkipped++
			slog.Warn("skipping Autronic availability product without ProductCode")
			continue
		}

		stock, warehouses := transformAvailability(source.Availability)
		if !hasAvailabilityPayload(stock, warehouses) {
			stats.ProductsSkipped++
			slog.Warn("skipping Autronic availability product without stock payload", "code", code)
			continue
		}

		result[code] = availabilityRecord{Stock: stock, Warehouses: warehouses}
	}
}

type availabilityTransformStats struct {
	ProductsEmitted         int
	ItemsWithVariants       int
	VariantsEmitted         int
	AvailabilityRecordsUsed int
}

func transformAvailabilityUpdate(catalog shoptet.Feed, availabilityByCode map[string]availabilityRecord) (shoptet.Feed, availabilityTransformStats) {
	var result shoptet.Feed
	var stats availabilityTransformStats
	usedCodes := make(map[string]struct{}, len(availabilityByCode))

	for _, catalogItem := range catalog.Items {
		if len(catalogItem.Variants) > 0 {
			variants := make([]shoptet.Variant, 0, len(catalogItem.Variants))
			for _, catalogVariant := range catalogItem.Variants {
				code := strings.TrimSpace(catalogVariant.Code)
				availability, ok := availabilityByCode[code]
				if !ok {
					continue
				}
				usedCodes[code] = struct{}{}
				variants = append(variants, shoptet.Variant{
					Code:       code,
					PriceVAT:   catalogVariant.PriceVAT,
					Stock:      availability.Stock,
					Warehouses: availability.Warehouses,
					Parameters: catalogVariant.Parameters,
				})
			}
			if len(variants) == 0 {
				continue
			}

			result.Items = append(result.Items, shoptet.Item{Variants: variants})
			stats.ProductsEmitted++
			stats.ItemsWithVariants++
			stats.VariantsEmitted += len(variants)
			continue
		}

		code := strings.TrimSpace(catalogItem.Code)
		availability, ok := availabilityByCode[code]
		if !ok {
			continue
		}
		usedCodes[code] = struct{}{}
		result.Items = append(result.Items, shoptet.Item{
			Code:       code,
			PriceVAT:   catalogItem.PriceVAT,
			Stock:      availability.Stock,
			Warehouses: availability.Warehouses,
		})
		stats.ProductsEmitted++
	}

	stats.AvailabilityRecordsUsed = len(usedCodes)
	return result, stats
}

func hasAvailabilityPayload(stock string, warehouses []shoptet.Warehouse) bool {
	if strings.TrimSpace(stock) != "" {
		return true
	}
	for _, warehouse := range warehouses {
		if strings.TrimSpace(warehouse.Name) != "" || strings.TrimSpace(warehouse.Value) != "" {
			return true
		}
	}
	return false
}

type sourceAvailabilityProduct struct {
	Code         string             `xml:"ProductCode"`
	Availability sourceAvailability `xml:"Availability"`
}

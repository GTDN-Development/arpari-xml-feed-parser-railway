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

		stats.ProductsEmitted++
		result.Items = append(result.Items, item)
		if options.MaxProducts > 0 && stats.ProductsEmitted >= options.MaxProducts {
			return result, stats, nil
		}
	}
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
		Code:            code,
		Name:            name,
		Description:     strings.TrimSpace(source.Description),
		PriceVAT:        strings.TrimSpace(source.PriceVAT),
		Stock:           normalizeNumber(source.Stock),
		Availability:    strings.TrimSpace(source.Availability),
		Categories:      categories,
		DefaultCategory: defaultCategory,
		Images:          transformImages(source.Images),
	}, true
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
	ID           string   `xml:"ID"`
	MainCategory string   `xml:"MAIN_CATEGORY"`
	Product      string   `xml:"PRODUCT"`
	PriceVAT     string   `xml:"PRICE_VAT"`
	Availability string   `xml:"DOSTUPNOST"`
	Stock        string   `xml:"STOCK"`
	PartNumber   string   `xml:"PART_NUMBER"`
	Description  string   `xml:"DESCRIPTION"`
	Images       []string `xml:"IMGURL>IMGURL"`
}

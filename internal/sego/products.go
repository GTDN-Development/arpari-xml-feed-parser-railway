package sego

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

const ProductsURL = "https://segocz.cz/src/Frontend/Files/Feeds/Catalog/zbozi_123456.xml"

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
		return nil, fmt.Errorf("create SEGO request: %w", err)
	}

	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("download SEGO feed: %w", err)
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		_ = response.Body.Close()
		return nil, fmt.Errorf("download SEGO feed: unexpected HTTP status %s", response.Status)
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
			return shoptet.Feed{}, stats, fmt.Errorf("parse SEGO products XML: %w", err)
		}

		start, ok := token.(xml.StartElement)
		if !ok || start.Name.Local != "SHOPITEM" {
			continue
		}

		var source sourceItem
		if err := decoder.DecodeElement(&source, &start); err != nil {
			return shoptet.Feed{}, stats, fmt.Errorf("decode SEGO SHOPITEM: %w", err)
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
	code := strings.TrimSpace(source.ItemID)
	name := strings.TrimSpace(source.ProductName)
	if code == "" || name == "" {
		slog.Warn("skipping SEGO product without required identity", "code", code, "name", name)
		return shoptet.Item{}, false
	}

	category := targetCategory(source.ProductName)
	return shoptet.Item{
		Code:            code,
		Name:            name,
		Description:     normalizeDescription(source.Description),
		PriceVAT:        strings.TrimSpace(source.PriceVAT),
		Availability:    transformDeliveryDate(source.DeliveryDate),
		EAN:             strings.TrimSpace(source.EAN),
		Categories:      []shoptet.Category{category},
		DefaultCategory: &category,
		Images:          transformImages(source),
	}, true
}

func normalizeDescription(value string) string {
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

func targetCategory(productName string) shoptet.Category {
	name := strings.ToLower(strings.TrimSpace(productName))
	if strings.Contains(name, "jednací") || strings.Contains(name, "konferenční") {
		return shoptet.Category{ID: "1146", Path: "ŽIDLE > KONFERENČNÍ ŽIDLE"}
	}
	return shoptet.Category{ID: "881", Path: "KANCELÁŘSKÉ ŽIDLE A KŘESLA"}
}

type sourceItem struct {
	ItemID            string   `xml:"ITEM_ID"`
	ProductName       string   `xml:"PRODUCTNAME"`
	Description       string   `xml:"DESCRIPTION"`
	EAN               string   `xml:"EAN"`
	ImageURL          string   `xml:"IMGURL"`
	AlternativeImages []string `xml:"IMGURL_ALTERNATIVE"`
	PriceVAT          string   `xml:"PRICE_VAT"`
	DeliveryDate      string   `xml:"DELIVERY_DATE"`
}

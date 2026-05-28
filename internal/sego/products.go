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
	"net/url"
	"strings"
	"time"

	"github.com/fanda/arpari-xml-feed-parser-railway/internal/shoptet"
)

const ProductsURL = "https://segocz.cz/src/Frontend/Files/Feeds/Catalog/zbozi_123456.xml"

const maxImagesPerProduct = 20

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
	MaxProducts        int
	PreferVariantItems bool
}

type ProductsStats struct {
	ProductsRead      int
	ProductsEmitted   int
	ProductsSkipped   int
	VariantsEmitted   int
	ItemsWithVariants int
}

func ParseProducts(ctx context.Context, r io.Reader, options ProductsOptions) (shoptet.Feed, ProductsStats, error) {
	decoder := xml.NewDecoder(r)
	var stats ProductsStats
	var entries []productEntry
	groups := make(map[string]*variantGroup)

	for {
		if err := ctx.Err(); err != nil {
			return shoptet.Feed{}, stats, err
		}

		token, err := decoder.Token()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
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
		item, ok := transformSimpleProduct(source)
		if !ok {
			stats.ProductsSkipped++
			continue
		}

		if info, ok := flatVariantInfo(source); ok {
			group := groups[info.Key]
			if group == nil {
				group = newVariantGroup(info, item)
				groups[info.Key] = group
				entries = append(entries, productEntry{Group: group})
			}
			group.Add(item, info)
			continue
		}

		entries = append(entries, productEntry{Item: item, HasItem: true})
	}

	result := emitProducts(entries, options, &stats)
	return result, stats, nil
}

func transformSimpleProduct(source sourceItem) (shoptet.Item, bool) {
	code := strings.TrimSpace(source.ItemID)
	name := strings.TrimSpace(source.ProductName)
	if code == "" || name == "" {
		slog.Warn("skipping SEGO product without required identity", "code", code, "name", name)
		return shoptet.Item{}, false
	}

	price := strings.TrimSpace(source.PriceVAT)
	category := targetCategory(source)
	currency := ""
	if price != "" {
		currency = "CZK"
	}
	return shoptet.Item{
		Code:            code,
		Name:            name,
		Description:     normalizeDescription(source.Description),
		Price:           price,
		Currency:        currency,
		Availability:    transformDeliveryDate(source.DeliveryDate),
		EAN:             strings.TrimSpace(source.EAN),
		Categories:      []shoptet.Category{category},
		DefaultCategory: &category,
		Images:          transformImages(source),
	}, true
}

func emitProducts(entries []productEntry, options ProductsOptions, stats *ProductsStats) shoptet.Feed {
	var result shoptet.Feed
	for _, item := range orderedItems(entries, options) {
		if len(item.Variants) > 0 {
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

func orderedItems(entries []productEntry, options ProductsOptions) []shoptet.Item {
	items := make([]shoptet.Item, 0, len(entries))
	for _, entry := range entries {
		if entry.Group != nil {
			items = append(items, entry.Group.Item())
		} else if entry.HasItem {
			items = append(items, entry.Item)
		}
	}

	if !options.PreferVariantItems {
		return items
	}

	variantsFirst := make([]shoptet.Item, 0, len(items))
	simpleItems := make([]shoptet.Item, 0, len(items))
	for _, item := range items {
		if len(item.Variants) > 0 {
			variantsFirst = append(variantsFirst, item)
		} else {
			simpleItems = append(simpleItems, item)
		}
	}
	return append(variantsFirst, simpleItems...)
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
		if isBrokenVariantPreviewURL(url) {
			continue
		}
		if _, ok := seen[url]; ok {
			continue
		}
		seen[url] = struct{}{}
		images = append(images, shoptet.Image{URL: url})
		if len(images) >= maxImagesPerProduct {
			break
		}
	}
	return images
}

func isBrokenVariantPreviewURL(rawURL string) bool {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return false
	}
	return strings.Contains(parsed.Path, "/Catalog/VariantImages/")
}

type productEntry struct {
	Item    shoptet.Item
	HasItem bool
	Group   *variantGroup
}

type variantInfo struct {
	Key        string
	ParentCode string
	BaseName   string
	Value      string
}

type variantGroup struct {
	parentCode      string
	baseName        string
	firstItem       shoptet.Item
	variants        []shoptet.Variant
	images          []shoptet.Image
	seenImageURLs   map[string]struct{}
	simpleFallbacks []shoptet.Item
}

func newVariantGroup(info variantInfo, firstItem shoptet.Item) *variantGroup {
	return &variantGroup{
		parentCode:    info.ParentCode,
		baseName:      info.BaseName,
		firstItem:     firstItem,
		seenImageURLs: make(map[string]struct{}),
	}
}

func (group *variantGroup) Add(item shoptet.Item, info variantInfo) {
	group.simpleFallbacks = append(group.simpleFallbacks, item)
	for _, image := range item.Images {
		url := strings.TrimSpace(image.URL)
		if url == "" {
			continue
		}
		if _, ok := group.seenImageURLs[url]; ok {
			continue
		}
		group.seenImageURLs[url] = struct{}{}
		group.images = append(group.images, image)
	}

	parameters := []shoptet.Parameter{{Name: "Barva", Value: info.Value}}
	group.variants = append(group.variants, shoptet.Variant{
		Code:         item.Code,
		EAN:          item.EAN,
		Price:        item.Price,
		Currency:     item.Currency,
		Availability: item.Availability,
		ImageRef:     variantImageRef(item.Images),
		Parameters:   parameters,
	})
}

func variantImageRef(images []shoptet.Image) string {
	if len(images) == 0 {
		return ""
	}
	return strings.TrimSpace(images[0].URL)
}

func (group *variantGroup) Item() shoptet.Item {
	if len(group.variants) < 2 {
		return group.simpleFallbacks[0]
	}

	return shoptet.Item{
		Code:            "",
		Name:            group.baseName,
		Description:     group.firstItem.Description,
		Price:           group.firstItem.Price,
		Currency:        group.firstItem.Currency,
		Availability:    group.firstItem.Availability,
		Categories:      group.firstItem.Categories,
		DefaultCategory: group.firstItem.DefaultCategory,
		Images:          group.images,
		Variants:        group.variants,
	}
}

func flatVariantInfo(source sourceItem) (variantInfo, bool) {
	baseName, nameVariant, ok := strings.Cut(strings.TrimSpace(source.ProductName), "|")
	if !ok {
		return variantInfo{}, false
	}
	baseName = strings.TrimSpace(baseName)
	nameVariant = normalizeParameterValue(nameVariant)
	color := source.ParameterValue("Barva")
	if baseName == "" || color == "" || nameVariant == "" || !strings.EqualFold(color, nameVariant) {
		return variantInfo{}, false
	}

	slug := productSlug(source.URL)
	if slug == "" {
		return variantInfo{}, false
	}

	return variantInfo{
		Key:        slug + "|" + strings.ToLower(baseName),
		ParentCode: "sego-" + slug,
		BaseName:   baseName,
		Value:      color,
	}, true
}

func productSlug(rawURL string) string {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return ""
	}
	path := strings.Trim(parsed.Path, "/")
	const prefix = "produkty/detail/"
	if !strings.HasPrefix(path, prefix) {
		return ""
	}
	slug := strings.Trim(strings.TrimPrefix(path, prefix), "/")
	if strings.Contains(slug, "/") {
		return ""
	}
	return slug
}

func normalizeParameterValue(value string) string {
	value = html.UnescapeString(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, "\u00a0", " ")
	return strings.Join(strings.Fields(value), " ")
}

type sourceItem struct {
	ItemID            string            `xml:"ITEM_ID"`
	ProductName       string            `xml:"PRODUCTNAME"`
	Description       string            `xml:"DESCRIPTION"`
	URL               string            `xml:"URL"`
	EAN               string            `xml:"EAN"`
	ImageURL          string            `xml:"IMGURL"`
	AlternativeImages []string          `xml:"IMGURL_ALTERNATIVE"`
	PriceVAT          string            `xml:"PRICE_VAT"`
	DeliveryDate      string            `xml:"DELIVERY_DATE"`
	Parameters        []sourceParameter `xml:"PARAM"`
}

func (item sourceItem) ParameterValue(name string) string {
	for _, parameter := range item.Parameters {
		if strings.TrimSpace(parameter.Name) == name {
			return normalizeParameterValue(parameter.Value)
		}
	}
	return ""
}

type sourceParameter struct {
	Name  string `xml:"PARAM_NAME"`
	Value string `xml:"VAL"`
}

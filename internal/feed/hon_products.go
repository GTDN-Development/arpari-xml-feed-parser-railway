package feed

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/fanda/arpari-xml-feed-parser-railway/internal/hon"
	"github.com/fanda/arpari-xml-feed-parser-railway/internal/shoptet"
)

type Hon struct {
	Downloader hon.Downloader
	SourceURL  string
}

type HonTest struct {
	Downloader  hon.Downloader
	SourceURL   string
	MaxProducts int
}

func (Hon) Name() string {
	return "hon"
}

func (Hon) Filename() string {
	return "hon.xml"
}

func (generator Hon) Generate(ctx context.Context, w io.Writer) (Result, error) {
	return generateHonProducts(ctx, w, generator.Name(), generator.Downloader, generator.SourceURL, 0, false)
}

func (HonTest) Name() string {
	return "hon-test"
}

func (HonTest) Filename() string {
	return "hon-test.xml"
}

func (generator HonTest) Generate(ctx context.Context, w io.Writer) (Result, error) {
	maxProducts := generator.MaxProducts
	if maxProducts <= 0 {
		maxProducts = 5
	}
	return generateHonProducts(ctx, w, generator.Name(), generator.Downloader, generator.SourceURL, maxProducts, false)
}

func generateHonProducts(ctx context.Context, w io.Writer, supplier string, configuredDownloader hon.Downloader, configuredSourceURL string, maxProducts int, includeCategories bool) (Result, error) {
	downloader := configuredDownloader
	if downloader == nil {
		downloader = hon.HTTPDownloader{}
	}

	sourceURL := configuredSourceURL
	if sourceURL == "" {
		sourceURL = hon.ProductsURL
	}

	body, err := downloader.Download(ctx, sourceURL)
	if err != nil {
		return Result{}, err
	}
	defer body.Close()

	feed, stats, err := hon.ParseProducts(ctx, body, hon.ProductsOptions{MaxProducts: maxProducts})
	if err != nil {
		return Result{}, err
	}
	if len(feed.Items) == 0 {
		return Result{ItemsSkipped: stats.ProductsSkipped}, fmt.Errorf("HON output is empty after transformation")
	}
	if !includeCategories {
		stripCategories(feed)
	}

	slog.Info(
		"HON products transformed",
		"supplier", supplier,
		"productsRead", stats.ProductsRead,
		"productsEmitted", stats.ProductsEmitted,
		"productsSkipped", stats.ProductsSkipped,
		"itemsWithVariants", stats.ItemsWithVariants,
		"variantsEmitted", stats.VariantsEmitted,
	)

	if err := shoptet.Write(w, feed); err != nil {
		return Result{
			ItemsProcessed: stats.ProductsEmitted,
			ItemsSkipped:   stats.ProductsSkipped,
		}, err
	}

	return Result{
		ItemsProcessed: stats.ProductsEmitted,
		ItemsSkipped:   stats.ProductsSkipped,
	}, nil
}

func stripCategories(feed shoptet.Feed) {
	for index := range feed.Items {
		feed.Items[index].Categories = nil
		feed.Items[index].DefaultCategory = nil
	}
}

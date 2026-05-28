package feed

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/fanda/arpari-xml-feed-parser-railway/internal/shoptet"
	"github.com/fanda/arpari-xml-feed-parser-railway/internal/stima"
)

type StimaProducts struct {
	Downloader stima.Downloader
	SourceURL  string
}

type StimaProductsTest struct {
	Downloader  stima.Downloader
	SourceURL   string
	MaxProducts int
}

func (StimaProducts) Name() string {
	return "stima-products"
}

func (StimaProducts) Filename() string {
	return "stima-products.xml"
}

func (generator StimaProducts) Generate(ctx context.Context, w io.Writer) (Result, error) {
	return generateStimaProducts(ctx, w, generator.Name(), generator.Downloader, generator.SourceURL, 0)
}

func (StimaProductsTest) Name() string {
	return "stima-products-test"
}

func (StimaProductsTest) Filename() string {
	return "stima-products-test.xml"
}

func (generator StimaProductsTest) Generate(ctx context.Context, w io.Writer) (Result, error) {
	maxProducts := generator.MaxProducts
	if maxProducts <= 0 {
		maxProducts = 5
	}
	return generateStimaProducts(ctx, w, generator.Name(), generator.Downloader, generator.SourceURL, maxProducts)
}

func generateStimaProducts(ctx context.Context, w io.Writer, supplier string, configuredDownloader stima.Downloader, configuredSourceURL string, maxProducts int) (Result, error) {
	downloader := configuredDownloader
	if downloader == nil {
		downloader = stima.HTTPDownloader{}
	}

	sourceURL := configuredSourceURL
	if sourceURL == "" {
		sourceURL = stima.ProductsURL
	}

	body, err := downloader.Download(ctx, sourceURL)
	if err != nil {
		return Result{}, err
	}
	defer body.Close()

	feed, stats, err := stima.ParseProducts(ctx, body, stima.ProductsOptions{
		MaxVariantsPerProduct: shoptet.DefaultMaxVariantsPerItem,
		MaxProducts:           maxProducts,
	})
	if err != nil {
		return Result{}, err
	}
	if len(feed.Items) == 0 {
		return Result{ItemsSkipped: stats.ProductsSkipped}, fmt.Errorf("STIMA products output is empty after transformation")
	}

	slog.Info(
		"STIMA products transformed",
		"supplier", supplier,
		"productsRead", stats.ProductsRead,
		"productsEmitted", stats.ProductsEmitted,
		"productsSkipped", stats.ProductsSkipped,
		"productsTrimmed", stats.ProductsTrimmed,
		"itemsWithVariants", stats.ItemsWithVariants,
		"variantsEmitted", stats.VariantsEmitted,
		"variantsSkipped", stats.VariantsSkipped,
		"variantsTrimmed", stats.VariantsTrimmed,
	)

	if err := shoptet.WriteWithLimits(w, feed, shoptet.Limits{
		MaxVariantsPerItem: shoptet.DefaultMaxVariantsPerItem,
	}); err != nil {
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

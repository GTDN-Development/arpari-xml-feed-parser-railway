package feed

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/fanda/arpari-xml-feed-parser-railway/internal/drevocal"
	"github.com/fanda/arpari-xml-feed-parser-railway/internal/shoptet"
)

type Drevocal struct {
	Downloader drevocal.Downloader
	SourceURL  string
}

type DrevocalTest struct {
	Downloader  drevocal.Downloader
	SourceURL   string
	MaxProducts int
}

func (Drevocal) Name() string {
	return "drevocal"
}

func (Drevocal) Filename() string {
	return "drevocal.xml"
}

func (generator Drevocal) Generate(ctx context.Context, w io.Writer) (Result, error) {
	return generateDrevocalProducts(ctx, w, generator.Name(), generator.Downloader, generator.SourceURL, 0)
}

func (DrevocalTest) Name() string {
	return "drevocal-test"
}

func (DrevocalTest) Filename() string {
	return "drevocal-test.xml"
}

func (generator DrevocalTest) Generate(ctx context.Context, w io.Writer) (Result, error) {
	maxProducts := generator.MaxProducts
	if maxProducts <= 0 {
		maxProducts = 5
	}
	return generateDrevocalProducts(ctx, w, generator.Name(), generator.Downloader, generator.SourceURL, maxProducts)
}

func generateDrevocalProducts(ctx context.Context, w io.Writer, supplier string, configuredDownloader drevocal.Downloader, configuredSourceURL string, maxProducts int) (Result, error) {
	downloader := configuredDownloader
	if downloader == nil {
		downloader = drevocal.HTTPDownloader{}
	}

	sourceURL := configuredSourceURL
	if sourceURL == "" {
		sourceURL = drevocal.ProductsURL
	}

	body, err := downloader.Download(ctx, sourceURL)
	if err != nil {
		return Result{}, err
	}
	defer body.Close()

	feed, stats, err := drevocal.ParseProducts(ctx, body, drevocal.ProductsOptions{
		MaxProducts:           maxProducts,
		MaxVariantsPerProduct: shoptet.DefaultMaxVariantsPerItem,
	})
	if err != nil {
		return Result{}, err
	}
	if len(feed.Items) == 0 {
		return Result{ItemsSkipped: stats.ProductsSkipped}, fmt.Errorf("Dřevočal output is empty after transformation")
	}

	slog.Info(
		"Dřevočal products transformed",
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

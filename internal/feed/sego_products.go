package feed

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/fanda/arpari-xml-feed-parser-railway/internal/sego"
	"github.com/fanda/arpari-xml-feed-parser-railway/internal/shoptet"
)

type Sego struct {
	Downloader sego.Downloader
	SourceURL  string
}

type SegoTest struct {
	Downloader  sego.Downloader
	SourceURL   string
	MaxProducts int
}

func (Sego) Name() string {
	return "sego"
}

func (Sego) Filename() string {
	return "sego.xml"
}

func (generator Sego) Generate(ctx context.Context, w io.Writer) (Result, error) {
	return generateSegoProducts(ctx, w, generator.Name(), generator.Downloader, generator.SourceURL, 0)
}

func (SegoTest) Name() string {
	return "sego-test"
}

func (SegoTest) Filename() string {
	return "sego-test.xml"
}

func (generator SegoTest) Generate(ctx context.Context, w io.Writer) (Result, error) {
	maxProducts := generator.MaxProducts
	if maxProducts <= 0 {
		maxProducts = 5
	}
	return generateSegoProductsWithOptions(ctx, w, generator.Name(), generator.Downloader, generator.SourceURL, sego.ProductsOptions{
		MaxProducts:        maxProducts,
		PreferVariantItems: true,
	})
}

func generateSegoProducts(ctx context.Context, w io.Writer, supplier string, configuredDownloader sego.Downloader, configuredSourceURL string, maxProducts int) (Result, error) {
	return generateSegoProductsWithOptions(ctx, w, supplier, configuredDownloader, configuredSourceURL, sego.ProductsOptions{MaxProducts: maxProducts})
}

func generateSegoProductsWithOptions(ctx context.Context, w io.Writer, supplier string, configuredDownloader sego.Downloader, configuredSourceURL string, options sego.ProductsOptions) (Result, error) {
	downloader := configuredDownloader
	if downloader == nil {
		downloader = sego.HTTPDownloader{}
	}

	sourceURL := configuredSourceURL
	if sourceURL == "" {
		sourceURL = sego.ProductsURL
	}

	body, err := downloader.Download(ctx, sourceURL)
	if err != nil {
		return Result{}, err
	}
	defer body.Close()

	feed, stats, err := sego.ParseProducts(ctx, body, options)
	if err != nil {
		return Result{}, err
	}
	if len(feed.Items) == 0 {
		return Result{ItemsSkipped: stats.ProductsSkipped}, fmt.Errorf("SEGO output is empty after transformation")
	}

	slog.Info(
		"SEGO products transformed",
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

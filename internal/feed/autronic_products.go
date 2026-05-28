package feed

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/fanda/arpari-xml-feed-parser-railway/internal/autronic"
	"github.com/fanda/arpari-xml-feed-parser-railway/internal/shoptet"
)

type AutronicProducts struct {
	Downloader autronic.Downloader
	SourceURL  string
}

type AutronicProductsTest struct {
	Downloader  autronic.Downloader
	SourceURL   string
	MaxProducts int
}

func (AutronicProducts) Name() string {
	return "autronic-products"
}

func (AutronicProducts) Filename() string {
	return "autronic-products.xml"
}

func (generator AutronicProducts) Generate(ctx context.Context, w io.Writer) (Result, error) {
	return generateAutronicProducts(ctx, w, generator.Name(), generator.Downloader, generator.SourceURL, 0)
}

func (AutronicProductsTest) Name() string {
	return "autronic-products-test"
}

func (AutronicProductsTest) Filename() string {
	return "autronic-products-test.xml"
}

func (generator AutronicProductsTest) Generate(ctx context.Context, w io.Writer) (Result, error) {
	maxProducts := generator.MaxProducts
	if maxProducts <= 0 {
		maxProducts = 2
	}
	return generateAutronicProducts(ctx, w, generator.Name(), generator.Downloader, generator.SourceURL, maxProducts)
}

func generateAutronicProducts(ctx context.Context, w io.Writer, supplier string, configuredDownloader autronic.Downloader, configuredSourceURL string, maxProducts int) (Result, error) {
	downloader := configuredDownloader
	if downloader == nil {
		downloader = autronic.HTTPDownloader{}
	}

	sourceURL := configuredSourceURL
	if sourceURL == "" {
		sourceURL = autronic.ProductsURL
	}

	body, err := downloader.Download(ctx, sourceURL)
	if err != nil {
		return Result{}, err
	}
	defer body.Close()

	feed, stats, err := autronic.ParseProducts(ctx, body, autronic.ProductsOptions{MaxProducts: maxProducts})
	if err != nil {
		return Result{}, err
	}
	if len(feed.Items) == 0 {
		return Result{ItemsSkipped: stats.ProductsSkipped}, fmt.Errorf("Autronic products output is empty after transformation")
	}

	slog.Info(
		"Autronic products transformed",
		"supplier", supplier,
		"productsRead", stats.ProductsRead,
		"productsEmitted", stats.ProductsEmitted,
		"productsSkipped", stats.ProductsSkipped,
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

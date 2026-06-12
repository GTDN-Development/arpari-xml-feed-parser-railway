package feed

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/fanda/arpari-xml-feed-parser-railway/internal/sakypaky"
	"github.com/fanda/arpari-xml-feed-parser-railway/internal/shoptet"
)

type Sakypaky struct {
	Downloader sakypaky.Downloader
	SourceURL  string
}

type SakypakyTest struct {
	Downloader  sakypaky.Downloader
	SourceURL   string
	MaxProducts int
}

func (Sakypaky) Name() string {
	return "sakypaky"
}

func (Sakypaky) Filename() string {
	return "sakypaky.xml"
}

func (generator Sakypaky) Generate(ctx context.Context, w io.Writer) (Result, error) {
	return generateSakypakyProducts(ctx, w, generator.Name(), generator.Downloader, generator.SourceURL, sakypaky.ProductsOptions{})
}

func (SakypakyTest) Name() string {
	return "sakypaky-test"
}

func (SakypakyTest) Filename() string {
	return "sakypaky-test.xml"
}

func (generator SakypakyTest) Generate(ctx context.Context, w io.Writer) (Result, error) {
	maxProducts := generator.MaxProducts
	if maxProducts <= 0 {
		maxProducts = 5
	}
	return generateSakypakyProducts(ctx, w, generator.Name(), generator.Downloader, generator.SourceURL, sakypaky.ProductsOptions{
		MaxProducts:        maxProducts,
		PreferVariantItems: true,
	})
}

func generateSakypakyProducts(ctx context.Context, w io.Writer, supplier string, configuredDownloader sakypaky.Downloader, configuredSourceURL string, options sakypaky.ProductsOptions) (Result, error) {
	downloader := configuredDownloader
	if downloader == nil {
		downloader = sakypaky.HTTPDownloader{}
	}

	sourceURL := configuredSourceURL
	if sourceURL == "" {
		sourceURL = sakypaky.ProductsURL
	}

	body, err := downloader.Download(ctx, sourceURL)
	if err != nil {
		return Result{}, err
	}
	defer body.Close()

	if options.MaxVariantsPerProduct <= 0 {
		options.MaxVariantsPerProduct = shoptet.DefaultMaxVariantsPerItem
	}
	feed, stats, err := sakypaky.ParseProducts(ctx, body, options)
	if err != nil {
		return Result{}, err
	}
	if len(feed.Items) == 0 {
		return Result{ItemsSkipped: stats.ProductsSkipped}, fmt.Errorf("Sakypaky output is empty after transformation")
	}

	slog.Info(
		"Sakypaky products transformed",
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

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

type StimaMissingVariants struct {
	Downloader stima.Downloader
	SourceURL  string
}

type StimaWhitelistedVariants struct {
	Downloader stima.Downloader
	SourceURL  string
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

func (StimaMissingVariants) Name() string {
	return "stima-missing-variants"
}

func (StimaMissingVariants) Filename() string {
	return "stima-missing-variants.xml"
}

func (generator StimaMissingVariants) Generate(ctx context.Context, w io.Writer) (Result, error) {
	return generateStimaMissingVariants(ctx, w, generator.Name(), generator.Downloader, generator.SourceURL)
}

func (StimaWhitelistedVariants) Name() string {
	return "stima-whitelisted-variants"
}

func (StimaWhitelistedVariants) Filename() string {
	return "stima-whitelisted-variants.xml"
}

func (generator StimaWhitelistedVariants) Generate(ctx context.Context, w io.Writer) (Result, error) {
	return generateStimaWhitelistedVariants(ctx, w, generator.Name(), generator.Downloader, generator.SourceURL)
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
		VariantWhitelist:      stima.DefaultFabricWhitelist(),
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

func generateStimaMissingVariants(ctx context.Context, w io.Writer, supplier string, configuredDownloader stima.Downloader, configuredSourceURL string) (Result, error) {
	return generateStimaMinimalVariants(ctx, w, supplier, configuredDownloader, configuredSourceURL, stima.ProductsOptions{
		MaxVariantsPerProduct: shoptet.DefaultMaxVariantsPerItem,
		VariantWhitelist:      stima.DefaultFabricWhitelist(),
		WhitelistedOnly:       true,
		MissingVariantsOnly:   true,
		MinimalVariantCatalog: true,
	}, "STIMA missing variants")
}

func generateStimaWhitelistedVariants(ctx context.Context, w io.Writer, supplier string, configuredDownloader stima.Downloader, configuredSourceURL string) (Result, error) {
	return generateStimaMinimalVariants(ctx, w, supplier, configuredDownloader, configuredSourceURL, stima.ProductsOptions{
		MaxVariantsPerProduct: shoptet.DefaultMaxVariantsPerItem,
		VariantWhitelist:      stima.DefaultFabricWhitelist(),
		WhitelistedOnly:       true,
		MinimalVariantCatalog: true,
	}, "STIMA whitelisted variants")
}

func generateStimaMinimalVariants(ctx context.Context, w io.Writer, supplier string, configuredDownloader stima.Downloader, configuredSourceURL string, options stima.ProductsOptions, logName string) (Result, error) {
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

	feed, stats, err := stima.ParseProducts(ctx, body, options)
	if err != nil {
		return Result{}, err
	}
	if len(feed.Items) == 0 {
		return Result{ItemsSkipped: stats.ProductsSkipped}, fmt.Errorf("%s output is empty after transformation", logName)
	}

	slog.Info(
		logName+" transformed",
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

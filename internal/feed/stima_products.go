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

func (StimaProducts) Name() string {
	return "stima-products"
}

func (StimaProducts) Filename() string {
	return "stima-products.xml"
}

func (generator StimaProducts) Generate(ctx context.Context, w io.Writer) (Result, error) {
	downloader := generator.Downloader
	if downloader == nil {
		downloader = stima.HTTPDownloader{}
	}

	sourceURL := generator.SourceURL
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
	})
	if err != nil {
		return Result{}, err
	}
	if len(feed.Items) == 0 {
		return Result{ItemsSkipped: stats.ProductsSkipped}, fmt.Errorf("STIMA products output is empty after transformation")
	}

	slog.Info(
		"STIMA products transformed",
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

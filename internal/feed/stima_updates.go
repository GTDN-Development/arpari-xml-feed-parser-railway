package feed

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/fanda/arpari-xml-feed-parser-railway/internal/shoptet"
	"github.com/fanda/arpari-xml-feed-parser-railway/internal/stima"
)

type StimaStock struct {
	Downloader stima.Downloader
	SourceURL  string
}

func (StimaStock) Name() string {
	return "stima-stock"
}

func (StimaStock) Filename() string {
	return "stima-stock.xml"
}

func (generator StimaStock) Generate(ctx context.Context, w io.Writer) (Result, error) {
	return generateStimaUpdate(ctx, w, generator.Downloader, generator.SourceURL, stima.StockURL, "STIMA stock", stima.ParseStock)
}

type StimaStockPrice struct {
	Downloader stima.Downloader
	SourceURL  string
}

func (StimaStockPrice) Name() string {
	return "stima-stock-price"
}

func (StimaStockPrice) Filename() string {
	return "stima-stock-price.xml"
}

func (generator StimaStockPrice) Generate(ctx context.Context, w io.Writer) (Result, error) {
	return generateStimaUpdate(ctx, w, generator.Downloader, generator.SourceURL, stima.StockPriceURL, "STIMA stock-price", stima.ParseStockPrice)
}

type stimaUpdateParser func(context.Context, io.Reader, stima.UpdateOptions) (shoptet.Feed, stima.UpdateStats, error)

func generateStimaUpdate(ctx context.Context, w io.Writer, downloader stima.Downloader, sourceURL, defaultSourceURL, logName string, parser stimaUpdateParser) (Result, error) {
	if downloader == nil {
		downloader = stima.HTTPDownloader{}
	}
	if sourceURL == "" {
		sourceURL = defaultSourceURL
	}

	body, err := downloader.Download(ctx, sourceURL)
	if err != nil {
		return Result{}, err
	}
	defer body.Close()

	feed, stats, err := parser(ctx, body, stima.UpdateOptions{
		MaxVariantsPerProduct: shoptet.DefaultMaxVariantsPerItem,
	})
	if err != nil {
		return Result{}, err
	}
	if len(feed.Items) == 0 {
		return Result{ItemsSkipped: stats.ProductsSkipped}, fmt.Errorf("%s output is empty after transformation", logName)
	}

	slog.Info(
		logName+" transformed",
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

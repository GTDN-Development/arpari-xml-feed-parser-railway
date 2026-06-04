package feed

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/fanda/arpari-xml-feed-parser-railway/internal/autronic"
	"github.com/fanda/arpari-xml-feed-parser-railway/internal/shoptet"
)

type AutronicAvailability struct {
	Downloader            autronic.Downloader
	ProductsSourceURL     string
	AvailabilitySourceURL string
}

func (AutronicAvailability) Name() string {
	return "autronic-availability"
}

func (AutronicAvailability) Filename() string {
	return "autronic-availability.xml"
}

func (generator AutronicAvailability) Generate(ctx context.Context, w io.Writer) (Result, error) {
	return generateAutronicAvailability(ctx, w, generator.Downloader, generator.ProductsSourceURL, generator.AvailabilitySourceURL)
}

func generateAutronicAvailability(ctx context.Context, w io.Writer, configuredDownloader autronic.Downloader, configuredProductsSourceURL, configuredAvailabilitySourceURL string) (Result, error) {
	downloader := configuredDownloader
	if downloader == nil {
		downloader = autronic.HTTPDownloader{}
	}

	productsSourceURL := configuredProductsSourceURL
	if productsSourceURL == "" {
		productsSourceURL = autronic.ProductsURL
	}
	availabilitySourceURL := configuredAvailabilitySourceURL
	if availabilitySourceURL == "" {
		availabilitySourceURL = autronic.AvailabilityURL
	}

	productsBody, err := downloader.Download(ctx, productsSourceURL)
	if err != nil {
		return Result{}, err
	}
	defer productsBody.Close()

	availabilityBody, err := downloader.Download(ctx, availabilitySourceURL)
	if err != nil {
		return Result{}, err
	}
	defer availabilityBody.Close()

	feed, stats, err := autronic.ParseAvailability(ctx, productsBody, availabilityBody, autronic.AvailabilityOptions{})
	if err != nil {
		return Result{}, err
	}
	if len(feed.Items) == 0 {
		return Result{ItemsSkipped: stats.ProductsSkipped}, fmt.Errorf("Autronic availability output is empty after transformation")
	}

	slog.Info(
		"Autronic availability transformed",
		"catalogProductsRead", stats.CatalogProductsRead,
		"catalogProductsEmitted", stats.CatalogProductsEmitted,
		"catalogProductsSkipped", stats.CatalogProductsSkipped,
		"availabilityProductsRead", stats.ProductsRead,
		"productsEmitted", stats.ProductsEmitted,
		"productsSkipped", stats.ProductsSkipped,
		"availabilityRecordsUsed", stats.AvailabilityRecordsUsed,
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

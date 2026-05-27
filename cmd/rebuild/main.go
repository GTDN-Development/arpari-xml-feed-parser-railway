package main

import (
	"context"
	"flag"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/fanda/arpari-xml-feed-parser-railway/internal/config"
	"github.com/fanda/arpari-xml-feed-parser-railway/internal/feed"
	runstatus "github.com/fanda/arpari-xml-feed-parser-railway/internal/status"
	"github.com/fanda/arpari-xml-feed-parser-railway/internal/storage"
)

func main() {
	supplier := flag.String("supplier", "", "supplier feed to rebuild")
	flag.Parse()

	if *supplier == "" {
		slog.Error("missing required --supplier")
		os.Exit(2)
	}

	generator, err := feed.Find(*supplier)
	if err != nil {
		slog.Error("find supplier", "error", err)
		os.Exit(1)
	}

	dataDir := config.DataDir()
	publisher := storage.NewPublisher(dataDir)
	statusStore := runstatus.NewStore(dataDir)

	var result feed.Result
	if err := publisher.Publish(generator.Filename(), func(w io.Writer) error {
		var generateErr error
		result, generateErr = generator.Generate(context.Background(), w)
		return generateErr
	}); err != nil {
		slog.Error("rebuild feed", "supplier", generator.Name(), "error", err)
		if statusErr := statusStore.Update(generator.Name(), runstatus.NewFeedStatus(generator.Filename(), runstatus.Failed, result.ItemsProcessed, result.ItemsSkipped, err.Error(), time.Now())); statusErr != nil {
			slog.Error("write failed feed status", "supplier", generator.Name(), "error", statusErr)
		}
		os.Exit(1)
	}

	if err := statusStore.Update(generator.Name(), runstatus.NewFeedStatus(generator.Filename(), runstatus.Success, result.ItemsProcessed, result.ItemsSkipped, "", time.Now())); err != nil {
		slog.Error("write feed status", "supplier", generator.Name(), "error", err)
		os.Exit(1)
	}

	slog.Info("feed rebuilt", "supplier", generator.Name(), "path", filepath.Join(dataDir, "feeds", generator.Filename()))
}

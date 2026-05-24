package main

import (
	"context"
	"flag"
	"io"
	"log/slog"
	"os"

	"github.com/fanda/arpari-xml-feed-parser-railway/internal/feed"
	"github.com/fanda/arpari-xml-feed-parser-railway/internal/storage"
)

const dataDir = "data"

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

	publisher := storage.NewPublisher(dataDir)
	if err := publisher.Publish(generator.Filename(), func(w io.Writer) error {
		return generator.Generate(context.Background(), w)
	}); err != nil {
		slog.Error("rebuild feed", "supplier", generator.Name(), "error", err)
		os.Exit(1)
	}

	slog.Info("feed rebuilt", "supplier", generator.Name(), "path", "data/feeds/"+generator.Filename())
}

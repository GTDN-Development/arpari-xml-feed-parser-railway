package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/fanda/arpari-xml-feed-parser-railway/internal/config"
	feedrebuild "github.com/fanda/arpari-xml-feed-parser-railway/internal/rebuild"
)

func main() {
	supplier := flag.String("supplier", "", "supplier feed to rebuild")
	flag.Parse()

	if *supplier == "" {
		slog.Error("missing required --supplier")
		os.Exit(2)
	}

	dataDir := config.DataDir()
	runner := feedrebuild.NewRunner(dataDir)
	result, err := runner.RunName(context.Background(), *supplier)
	if err != nil {
		slog.Error("rebuild feed", "supplier", *supplier, "error", err)
		os.Exit(1)
	}

	slog.Info("feed rebuilt", "supplier", result.Supplier, "path", filepath.Join(dataDir, "feeds", result.Filename))
}

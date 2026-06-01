package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {
	rebuildURL := os.Getenv("REBUILD_URL")
	rebuildToken := os.Getenv("REBUILD_TOKEN")
	if rebuildURL == "" {
		slog.Error("missing required REBUILD_URL")
		os.Exit(2)
	}
	if rebuildToken == "" {
		slog.Error("missing required REBUILD_TOKEN")
		os.Exit(2)
	}

	timeout := 30 * time.Minute
	if value := os.Getenv("REBUILD_TIMEOUT"); value != "" {
		parsed, err := time.ParseDuration(value)
		if err != nil {
			slog.Error("parse REBUILD_TIMEOUT", "value", value, "error", err)
			os.Exit(2)
		}
		timeout = parsed
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, rebuildURL, nil)
	if err != nil {
		slog.Error("create rebuild request", "error", err)
		os.Exit(2)
	}
	request.Header.Set("Authorization", "Bearer "+rebuildToken)

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		slog.Error("call rebuild endpoint", "url", rebuildURL, "error", err)
		os.Exit(1)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(io.LimitReader(response.Body, 64*1024))
	if err != nil {
		slog.Error("read rebuild response", "error", err)
		os.Exit(1)
	}

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		slog.Error("rebuild endpoint failed", "status", response.Status, "body", strings.TrimSpace(string(body)))
		os.Exit(1)
	}

	fmt.Print(string(body))
	slog.Info("rebuild endpoint completed", "status", response.Status)
}

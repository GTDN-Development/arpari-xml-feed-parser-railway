package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strconv"
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
	maxAttempts := 5
	if value := os.Getenv("REBUILD_MAX_ATTEMPTS"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed <= 0 {
			slog.Error("parse REBUILD_MAX_ATTEMPTS", "value", value, "error", err)
			os.Exit(2)
		}
		maxAttempts = parsed
	}
	retryDelay := 30 * time.Second
	if value := os.Getenv("REBUILD_RETRY_DELAY"); value != "" {
		parsed, err := time.ParseDuration(value)
		if err != nil {
			slog.Error("parse REBUILD_RETRY_DELAY", "value", value, "error", err)
			os.Exit(2)
		}
		retryDelay = parsed
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	body, status, err := callRebuildWithRetries(ctx, http.DefaultClient, rebuildURL, rebuildToken, maxAttempts, retryDelay)
	if err != nil {
		slog.Error("call rebuild endpoint", "url", rebuildURL, "status", status, "error", err)
		os.Exit(1)
	}

	fmt.Print(body)
	slog.Info("rebuild endpoint completed", "status", status)
}

type httpClient interface {
	Do(*http.Request) (*http.Response, error)
}

func callRebuildWithRetries(ctx context.Context, client httpClient, rebuildURL, rebuildToken string, maxAttempts int, retryDelay time.Duration) (string, string, error) {
	var lastStatus string
	var lastErr error

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		body, status, retryable, err := callRebuild(ctx, client, rebuildURL, rebuildToken)
		if err == nil {
			return body, status, nil
		}
		lastStatus = status
		lastErr = err
		if !retryable || attempt == maxAttempts {
			return body, status, err
		}

		slog.Warn(
			"rebuild endpoint attempt failed; retrying",
			"url", rebuildURL,
			"attempt", attempt,
			"maxAttempts", maxAttempts,
			"status", status,
			"error", err,
			"retryDelay", retryDelay.String(),
		)
		if err := sleepContext(ctx, retryDelay); err != nil {
			return body, status, err
		}
	}

	return "", lastStatus, lastErr
}

func callRebuild(ctx context.Context, client httpClient, rebuildURL, rebuildToken string) (body string, status string, retryable bool, err error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, rebuildURL, nil)
	if err != nil {
		return "", "", false, fmt.Errorf("create rebuild request: %w", err)
	}
	request.Header.Set("Authorization", "Bearer "+rebuildToken)

	response, err := client.Do(request)
	if err != nil {
		return "", "", true, err
	}
	defer response.Body.Close()

	bodyBytes, err := io.ReadAll(io.LimitReader(response.Body, 64*1024))
	if err != nil {
		return "", response.Status, true, fmt.Errorf("read rebuild response: %w", err)
	}
	body = string(bodyBytes)

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return body, response.Status, isRetryableStatus(response.StatusCode), fmt.Errorf("rebuild endpoint failed: %s: %s", response.Status, strings.TrimSpace(body))
	}

	return body, response.Status, false, nil
}

func isRetryableStatus(statusCode int) bool {
	switch statusCode {
	case http.StatusRequestTimeout,
		http.StatusTooEarly,
		http.StatusTooManyRequests,
		http.StatusInternalServerError,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout:
		return true
	default:
		return false
	}
}

func sleepContext(ctx context.Context, duration time.Duration) error {
	timer := time.NewTimer(duration)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

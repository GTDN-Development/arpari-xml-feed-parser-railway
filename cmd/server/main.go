package main

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/fanda/arpari-xml-feed-parser-railway/internal/config"
	feedrebuild "github.com/fanda/arpari-xml-feed-parser-railway/internal/rebuild"
	runstatus "github.com/fanda/arpari-xml-feed-parser-railway/internal/status"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	dataDir := config.DataDir()

	addr := "0.0.0.0:" + port
	server := &http.Server{
		Addr:    addr,
		Handler: newMux(dataDir),
	}

	slog.Info("starting server", "addr", addr, "dataDir", dataDir)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}

func newMux(dataDir string) http.Handler {
	return newMuxWithRebuildRunner(dataDir, os.Getenv("REBUILD_TOKEN"), feedrebuild.NewRunner(dataDir))
}

type rebuildRunner interface {
	RunName(context.Context, string) (feedrebuild.Result, error)
	RunScheduled(context.Context) []feedrebuild.Result
}

func newMuxWithRebuildRunner(dataDir string, rebuildToken string, runner rebuildRunner) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", helloHandler)
	mux.HandleFunc("GET /healthz", healthHandler)
	mux.HandleFunc("GET /status", statusHandler(dataDir))
	mux.HandleFunc("GET /feeds/{filename}", feedHandler(dataDir))
	mux.HandleFunc("POST /internal/rebuild/all", rebuildAllHandler(rebuildToken, runner))
	mux.HandleFunc("POST /internal/rebuild/{supplier}", rebuildSupplierHandler(rebuildToken, runner))
	return mux
}

func helloHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintln(w, "Hello world!")
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintln(w, "ok")
}

func statusHandler(dataDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		file, err := runstatus.NewStore(dataDir).Read()
		if err != nil {
			http.Error(w, "read status", http.StatusInternalServerError)
			slog.Error("read status", "error", err)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(file); err != nil {
			slog.Error("write status response", "error", err)
		}
	}
}

func feedHandler(dataDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		filename := r.PathValue("filename")
		if filename == "" || filepath.Base(filename) != filename {
			http.NotFound(w, r)
			return
		}

		path := filepath.Join(dataDir, "feeds", filename)
		w.Header().Set("Content-Type", "application/xml; charset=utf-8")
		http.ServeFile(w, r, path)
	}
}

func rebuildSupplierHandler(rebuildToken string, runner rebuildRunner) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizeRebuild(w, r, rebuildToken) {
			return
		}

		supplier := r.PathValue("supplier")
		result, err := runner.RunName(r.Context(), supplier)
		statusCode := http.StatusOK
		if err != nil {
			statusCode = http.StatusInternalServerError
			if errors.Is(err, feedrebuild.ErrUnknownSupplier) {
				statusCode = http.StatusNotFound
			} else if errors.Is(err, feedrebuild.ErrAlreadyRunning) {
				statusCode = http.StatusConflict
			}
			slog.Error("rebuild feed", "supplier", supplier, "error", err)
		}

		writeJSON(w, statusCode, rebuildResponse{Results: []feedrebuild.Result{result}})
	}
}

func rebuildAllHandler(rebuildToken string, runner rebuildRunner) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizeRebuild(w, r, rebuildToken) {
			return
		}

		results := runner.RunScheduled(r.Context())
		statusCode := http.StatusOK
		for _, result := range results {
			if result.Status != runstatus.Success {
				if result.Status == "running" && statusCode == http.StatusOK {
					statusCode = http.StatusConflict
				} else if result.Status != "running" {
					statusCode = http.StatusInternalServerError
				}
				slog.Error("rebuild feed", "supplier", result.Supplier, "status", result.Status, "error", result.Error)
			}
		}

		writeJSON(w, statusCode, rebuildResponse{Results: results})
	}
}

type rebuildResponse struct {
	Results []feedrebuild.Result `json:"results"`
}

func authorizeRebuild(w http.ResponseWriter, r *http.Request, rebuildToken string) bool {
	if rebuildToken == "" {
		http.Error(w, "REBUILD_TOKEN is not configured", http.StatusServiceUnavailable)
		return false
	}

	token := bearerToken(r.Header.Get("Authorization"))
	if token == "" {
		token = r.Header.Get("X-Rebuild-Token")
	}

	if subtle.ConstantTimeCompare([]byte(token), []byte(rebuildToken)) != 1 {
		w.Header().Set("WWW-Authenticate", "Bearer")
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return false
	}

	return true
}

func bearerToken(header string) string {
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(header, prefix))
}

func writeJSON(w http.ResponseWriter, statusCode int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(value); err != nil {
		slog.Error("write json response", "error", err)
	}
}

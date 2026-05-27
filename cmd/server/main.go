package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"github.com/fanda/arpari-xml-feed-parser-railway/internal/config"
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
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", helloHandler)
	mux.HandleFunc("GET /healthz", healthHandler)
	mux.HandleFunc("GET /status", statusHandler(dataDir))
	mux.HandleFunc("GET /feeds/{filename}", feedHandler(dataDir))
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

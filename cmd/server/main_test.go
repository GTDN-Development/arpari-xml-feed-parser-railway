package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	runstatus "github.com/fanda/arpari-xml-feed-parser-railway/internal/status"
)

func TestHelloHandler(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	recorder := httptest.NewRecorder()

	newMux().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}

	if body := recorder.Body.String(); body != "Hello world!\n" {
		t.Fatalf("expected hello response, got %q", body)
	}
}

func TestUnknownRouteReturnsNotFound(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/nope", nil)
	recorder := httptest.NewRecorder()

	newMux().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, recorder.Code)
	}
}

func TestStatusHandlerReturnsEmptyStatusWhenMissing(t *testing.T) {
	t.Chdir(t.TempDir())

	request := httptest.NewRequest(http.MethodGet, "/status", nil)
	recorder := httptest.NewRecorder()

	newMux().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
	if contentType := recorder.Header().Get("Content-Type"); contentType != "application/json; charset=utf-8" {
		t.Fatalf("expected JSON content type, got %q", contentType)
	}

	var file runstatus.File
	if err := json.Unmarshal(recorder.Body.Bytes(), &file); err != nil {
		t.Fatalf("decode status response: %v", err)
	}
	if len(file.Feeds) != 0 {
		t.Fatalf("expected no feed statuses, got %d", len(file.Feeds))
	}
}

func TestFeedHandlerServesExistingFeed(t *testing.T) {
	t.Chdir(t.TempDir())

	feedDir := filepath.Join("data", "feeds")
	if err := os.MkdirAll(feedDir, 0o755); err != nil {
		t.Fatalf("create feed dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(feedDir, "hello.xml"), []byte("<SHOP></SHOP>"), 0o644); err != nil {
		t.Fatalf("write feed: %v", err)
	}

	request := httptest.NewRequest(http.MethodGet, "/feeds/hello.xml", nil)
	recorder := httptest.NewRecorder()

	newMux().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
	if contentType := recorder.Header().Get("Content-Type"); contentType != "application/xml; charset=utf-8" {
		t.Fatalf("expected XML content type, got %q", contentType)
	}
	if body := recorder.Body.String(); body != "<SHOP></SHOP>" {
		t.Fatalf("expected feed response, got %q", body)
	}
}

package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
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

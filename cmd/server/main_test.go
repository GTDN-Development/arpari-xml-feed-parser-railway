package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	feedrebuild "github.com/fanda/arpari-xml-feed-parser-railway/internal/rebuild"
	runstatus "github.com/fanda/arpari-xml-feed-parser-railway/internal/status"
)

func TestHelloHandler(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	recorder := httptest.NewRecorder()

	newMux(t.TempDir()).ServeHTTP(recorder, request)

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

	newMux(t.TempDir()).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, recorder.Code)
	}
}

func TestStatusHandlerReturnsEmptyStatusWhenMissing(t *testing.T) {
	dataDir := t.TempDir()

	request := httptest.NewRequest(http.MethodGet, "/status", nil)
	recorder := httptest.NewRecorder()

	newMux(dataDir).ServeHTTP(recorder, request)

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

func TestStatusHandlerReadsStatusFromDataDir(t *testing.T) {
	dataDir := t.TempDir()
	statusStore := runstatus.NewStore(dataDir)
	if err := statusStore.Write(runstatus.File{Feeds: map[string]runstatus.FeedStatus{
		"hello": {
			Filename:       "hello.xml",
			LastRunAt:      "2026-05-27T08:00:00Z",
			Status:         runstatus.Success,
			ItemsProcessed: 1,
			ItemsSkipped:   0,
			Error:          "",
		},
	}}); err != nil {
		t.Fatalf("write status: %v", err)
	}

	request := httptest.NewRequest(http.MethodGet, "/status", nil)
	recorder := httptest.NewRecorder()

	newMux(dataDir).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}

	var file runstatus.File
	if err := json.Unmarshal(recorder.Body.Bytes(), &file); err != nil {
		t.Fatalf("decode status response: %v", err)
	}
	if file.Feeds["hello"].Filename != "hello.xml" {
		t.Fatalf("expected hello status from data dir, got %#v", file.Feeds["hello"])
	}
}

func TestFeedHandlerServesExistingFeed(t *testing.T) {
	dataDir := t.TempDir()

	feedDir := filepath.Join(dataDir, "feeds")
	if err := os.MkdirAll(feedDir, 0o755); err != nil {
		t.Fatalf("create feed dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(feedDir, "hello.xml"), []byte("<SHOP></SHOP>"), 0o644); err != nil {
		t.Fatalf("write feed: %v", err)
	}

	request := httptest.NewRequest(http.MethodGet, "/feeds/hello.xml", nil)
	recorder := httptest.NewRecorder()

	newMux(dataDir).ServeHTTP(recorder, request)

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

func TestRebuildSupplierRequiresConfiguredToken(t *testing.T) {
	runner := &fakeRebuildRunner{}
	request := httptest.NewRequest(http.MethodPost, "/internal/rebuild/stima-stock", nil)
	request.Header.Set("Authorization", "Bearer secret")
	recorder := httptest.NewRecorder()

	newMuxWithRebuildRunner(t.TempDir(), "", runner).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d", http.StatusServiceUnavailable, recorder.Code)
	}
	if runner.runNameCalls != 0 {
		t.Fatalf("expected rebuild not to run, got %d calls", runner.runNameCalls)
	}
}

func TestRebuildSupplierRejectsMissingToken(t *testing.T) {
	runner := &fakeRebuildRunner{}
	request := httptest.NewRequest(http.MethodPost, "/internal/rebuild/stima-stock", nil)
	recorder := httptest.NewRecorder()

	newMuxWithRebuildRunner(t.TempDir(), "secret", runner).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, recorder.Code)
	}
	if runner.runNameCalls != 0 {
		t.Fatalf("expected rebuild not to run, got %d calls", runner.runNameCalls)
	}
}

func TestRebuildSupplierRunsWithBearerToken(t *testing.T) {
	runner := &fakeRebuildRunner{
		runNameResult: feedrebuild.Result{
			Supplier:       "stima-stock",
			Filename:       "stima-stock.xml",
			Status:         runstatus.Success,
			ItemsProcessed: 2,
		},
	}
	request := httptest.NewRequest(http.MethodPost, "/internal/rebuild/stima-stock", nil)
	request.Header.Set("Authorization", "Bearer secret")
	recorder := httptest.NewRecorder()

	newMuxWithRebuildRunner(t.TempDir(), "secret", runner).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
	if runner.runNameCalls != 1 || runner.runNameSupplier != "stima-stock" {
		t.Fatalf("expected stima-stock rebuild, got supplier %q and %d calls", runner.runNameSupplier, runner.runNameCalls)
	}

	var response rebuildResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode rebuild response: %v", err)
	}
	if len(response.Results) != 1 || response.Results[0].Filename != "stima-stock.xml" {
		t.Fatalf("unexpected rebuild response: %#v", response)
	}
}

func TestRebuildSupplierReturnsNotFoundForUnknownSupplier(t *testing.T) {
	runner := &fakeRebuildRunner{
		runNameResult: feedrebuild.Result{
			Supplier: "missing",
			Status:   runstatus.Failed,
			Error:    "unknown supplier",
		},
		runNameErr: feedrebuild.ErrUnknownSupplier,
	}
	request := httptest.NewRequest(http.MethodPost, "/internal/rebuild/missing", nil)
	request.Header.Set("X-Rebuild-Token", "secret")
	recorder := httptest.NewRecorder()

	newMuxWithRebuildRunner(t.TempDir(), "secret", runner).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, recorder.Code)
	}
}

func TestRebuildAllReturnsServerErrorWhenAnyFeedFails(t *testing.T) {
	runner := &fakeRebuildRunner{
		scheduledResults: []feedrebuild.Result{
			{Supplier: "stima-stock", Filename: "stima-stock.xml", Status: runstatus.Success},
			{Supplier: "hon", Filename: "hon.xml", Status: runstatus.Failed, Error: "download failed"},
		},
	}
	request := httptest.NewRequest(http.MethodPost, "/internal/rebuild/all", nil)
	request.Header.Set("Authorization", "Bearer secret")
	recorder := httptest.NewRecorder()

	newMuxWithRebuildRunner(t.TempDir(), "secret", runner).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, recorder.Code)
	}
	if runner.runScheduledCalls != 1 {
		t.Fatalf("expected scheduled rebuild to run once, got %d calls", runner.runScheduledCalls)
	}
}

type fakeRebuildRunner struct {
	runNameCalls    int
	runNameSupplier string
	runNameResult   feedrebuild.Result
	runNameErr      error

	runScheduledCalls int
	scheduledResults  []feedrebuild.Result
}

func (runner *fakeRebuildRunner) RunName(_ context.Context, supplier string) (feedrebuild.Result, error) {
	runner.runNameCalls++
	runner.runNameSupplier = supplier
	return runner.runNameResult, runner.runNameErr
}

func (runner *fakeRebuildRunner) RunScheduled(_ context.Context) []feedrebuild.Result {
	runner.runScheduledCalls++
	return runner.scheduledResults
}

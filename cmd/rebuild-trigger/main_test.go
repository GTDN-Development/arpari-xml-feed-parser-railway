package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestCallRebuildWithRetriesRetriesNetworkErrors(t *testing.T) {
	client := &fakeHTTPClient{calls: []fakeHTTPCall{
		{err: errors.New("net/http: TLS handshake timeout")},
		{response: response(http.StatusOK, `{"results":[]}`)},
	}}

	body, status, err := callRebuildWithRetries(context.Background(), client, "https://example.test/internal/rebuild/all", "secret", 5, 0)
	if err != nil {
		t.Fatalf("call rebuild: %v", err)
	}
	if body != `{"results":[]}` {
		t.Fatalf("unexpected body %q", body)
	}
	if status != "200 OK" {
		t.Fatalf("unexpected status %q", status)
	}
	if len(client.requests) != 2 {
		t.Fatalf("expected 2 requests, got %d", len(client.requests))
	}
	if got := client.requests[1].Header.Get("Authorization"); got != "Bearer secret" {
		t.Fatalf("unexpected auth header %q", got)
	}
}

func TestCallRebuildWithRetriesRetriesTemporaryStatus(t *testing.T) {
	client := &fakeHTTPClient{calls: []fakeHTTPCall{
		{response: response(http.StatusServiceUnavailable, "service unavailable")},
		{response: response(http.StatusOK, `{"results":[]}`)},
	}}

	_, status, err := callRebuildWithRetries(context.Background(), client, "https://example.test/internal/rebuild/all", "secret", 5, 0)
	if err != nil {
		t.Fatalf("call rebuild: %v", err)
	}
	if status != "200 OK" {
		t.Fatalf("unexpected status %q", status)
	}
	if len(client.requests) != 2 {
		t.Fatalf("expected 2 requests, got %d", len(client.requests))
	}
}

func TestCallRebuildWithRetriesStopsOnNonRetryableStatus(t *testing.T) {
	client := &fakeHTTPClient{calls: []fakeHTTPCall{
		{response: response(http.StatusUnauthorized, "unauthorized")},
		{response: response(http.StatusOK, `{"results":[]}`)},
	}}

	_, status, err := callRebuildWithRetries(context.Background(), client, "https://example.test/internal/rebuild/all", "secret", 5, 0)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "401 Unauthorized") {
		t.Fatalf("unexpected error %v", err)
	}
	if status != "401 Unauthorized" {
		t.Fatalf("unexpected status %q", status)
	}
	if len(client.requests) != 1 {
		t.Fatalf("expected 1 request, got %d", len(client.requests))
	}
}

type fakeHTTPClient struct {
	calls    []fakeHTTPCall
	requests []*http.Request
}

type fakeHTTPCall struct {
	response *http.Response
	err      error
}

func (client *fakeHTTPClient) Do(request *http.Request) (*http.Response, error) {
	client.requests = append(client.requests, request)
	index := len(client.requests) - 1
	if index >= len(client.calls) {
		return nil, errors.New("unexpected request")
	}
	call := client.calls[index]
	if call.err != nil {
		return nil, call.err
	}
	return call.response, nil
}

func response(statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Status:     fmt.Sprintf("%d %s", statusCode, http.StatusText(statusCode)),
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

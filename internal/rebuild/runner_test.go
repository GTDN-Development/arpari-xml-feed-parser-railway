package rebuild

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/fanda/arpari-xml-feed-parser-railway/internal/feed"
	runstatus "github.com/fanda/arpari-xml-feed-parser-railway/internal/status"
)

func TestRunnerRunPublishesFeedAndWritesStatus(t *testing.T) {
	dataDir := t.TempDir()
	runner := NewRunner(dataDir)
	generator := fakeGenerator{
		name:     "demo",
		filename: "demo.xml",
		output:   `<?xml version="1.0" encoding="UTF-8"?><SHOP></SHOP>`,
		result:   feed.Result{ItemsProcessed: 3, ItemsSkipped: 1},
	}

	result, err := runner.Run(context.Background(), generator)
	if err != nil {
		t.Fatalf("run rebuild: %v", err)
	}

	if result.Status != runstatus.Success {
		t.Fatalf("expected success result, got %#v", result)
	}
	if result.ItemsProcessed != 3 || result.ItemsSkipped != 1 {
		t.Fatalf("expected generator counts in result, got %#v", result)
	}

	data, err := os.ReadFile(filepath.Join(dataDir, "feeds", "demo.xml"))
	if err != nil {
		t.Fatalf("read published feed: %v", err)
	}
	if string(data) != generator.output {
		t.Fatalf("unexpected feed content: %q", string(data))
	}

	statusFile, err := runstatus.NewStore(dataDir).Read()
	if err != nil {
		t.Fatalf("read status: %v", err)
	}
	if statusFile.Feeds["demo"].Status != runstatus.Success {
		t.Fatalf("expected success status, got %#v", statusFile.Feeds["demo"])
	}
}

func TestRunnerRunWritesFailureStatus(t *testing.T) {
	dataDir := t.TempDir()
	runner := NewRunner(dataDir)
	generator := fakeGenerator{
		name:     "demo",
		filename: "demo.xml",
		err:      errors.New("download failed"),
	}

	result, err := runner.Run(context.Background(), generator)
	if err == nil {
		t.Fatal("expected rebuild error")
	}

	if result.Status != runstatus.Failed || result.Error == "" {
		t.Fatalf("expected failed result, got %#v", result)
	}

	statusFile, err := runstatus.NewStore(dataDir).Read()
	if err != nil {
		t.Fatalf("read status: %v", err)
	}
	if statusFile.Feeds["demo"].Status != runstatus.Failed {
		t.Fatalf("expected failed status, got %#v", statusFile.Feeds["demo"])
	}
}

type fakeGenerator struct {
	name     string
	filename string
	output   string
	result   feed.Result
	err      error
}

func (generator fakeGenerator) Name() string {
	return generator.name
}

func (generator fakeGenerator) Filename() string {
	return generator.filename
}

func (generator fakeGenerator) Generate(_ context.Context, w io.Writer) (feed.Result, error) {
	if generator.err != nil {
		return generator.result, generator.err
	}
	if _, err := io.WriteString(w, generator.output); err != nil {
		return feed.Result{}, err
	}
	return generator.result, nil
}

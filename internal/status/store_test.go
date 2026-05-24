package status

import (
	"testing"
	"time"
)

func TestStoreReadReturnsEmptyStatusWhenMissing(t *testing.T) {
	store := NewStore(t.TempDir())

	file, err := store.Read()
	if err != nil {
		t.Fatalf("read missing status: %v", err)
	}

	if len(file.Feeds) != 0 {
		t.Fatalf("expected no feed statuses, got %d", len(file.Feeds))
	}
}

func TestStoreUpdateWritesAndReadsSuccessStatus(t *testing.T) {
	store := NewStore(t.TempDir())
	now := time.Date(2026, 5, 24, 12, 34, 56, 0, time.UTC)

	err := store.Update("hello", NewFeedStatus("hello.xml", Success, 1, 0, "", now))
	if err != nil {
		t.Fatalf("update status: %v", err)
	}

	file, err := store.Read()
	if err != nil {
		t.Fatalf("read status: %v", err)
	}

	got := file.Feeds["hello"]
	if got.Status != Success {
		t.Fatalf("expected status %q, got %q", Success, got.Status)
	}
	if got.LastRunAt != "2026-05-24T12:34:56Z" {
		t.Fatalf("unexpected lastRunAt: %q", got.LastRunAt)
	}
	if got.ItemsProcessed != 1 {
		t.Fatalf("expected 1 processed item, got %d", got.ItemsProcessed)
	}
}

func TestStoreUpdatePreservesOtherFeeds(t *testing.T) {
	store := NewStore(t.TempDir())
	now := time.Date(2026, 5, 24, 12, 34, 56, 0, time.UTC)

	if err := store.Update("hello", NewFeedStatus("hello.xml", Success, 1, 0, "", now)); err != nil {
		t.Fatalf("seed status: %v", err)
	}
	if err := store.Update("broken", NewFeedStatus("broken.xml", Failed, 0, 0, "boom", now)); err != nil {
		t.Fatalf("update failed status: %v", err)
	}

	file, err := store.Read()
	if err != nil {
		t.Fatalf("read status: %v", err)
	}

	if file.Feeds["hello"].Status != Success {
		t.Fatalf("hello status was not preserved: %q", file.Feeds["hello"].Status)
	}
	if file.Feeds["broken"].Error != "boom" {
		t.Fatalf("expected broken feed error, got %q", file.Feeds["broken"].Error)
	}
}

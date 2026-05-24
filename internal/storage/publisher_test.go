package storage

import (
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestPublisherPublishesValidXML(t *testing.T) {
	publisher := NewPublisher(t.TempDir())

	err := publisher.Publish("hello.xml", func(w io.Writer) error {
		_, err := io.WriteString(w, `<?xml version="1.0" encoding="UTF-8"?><SHOP></SHOP>`)
		return err
	})
	if err != nil {
		t.Fatalf("publish valid XML: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(publisher.FeedDir, "hello.xml"))
	if err != nil {
		t.Fatalf("read published feed: %v", err)
	}

	if string(content) != `<?xml version="1.0" encoding="UTF-8"?><SHOP></SHOP>` {
		t.Fatalf("unexpected published content: %q", string(content))
	}
}

func TestPublisherDoesNotOverwriteExistingFeedWithInvalidXML(t *testing.T) {
	publisher := NewPublisher(t.TempDir())
	if err := os.MkdirAll(publisher.FeedDir, 0o755); err != nil {
		t.Fatalf("create feed dir: %v", err)
	}

	target := filepath.Join(publisher.FeedDir, "hello.xml")
	if err := os.WriteFile(target, []byte("<SHOP></SHOP>"), 0o644); err != nil {
		t.Fatalf("seed feed: %v", err)
	}

	err := publisher.Publish("hello.xml", func(w io.Writer) error {
		_, err := io.WriteString(w, `<SHOP>`)
		return err
	})
	if err == nil {
		t.Fatal("expected invalid XML error")
	}

	content, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read existing feed: %v", err)
	}

	if string(content) != "<SHOP></SHOP>" {
		t.Fatalf("existing feed was overwritten: %q", string(content))
	}
}

package feed

import (
	"bytes"
	"context"
	"encoding/xml"
	"testing"
)

func TestHelloGenerateReturnsWellFormedXML(t *testing.T) {
	var output bytes.Buffer

	if err := (Hello{}).Generate(context.Background(), &output); err != nil {
		t.Fatalf("generate hello feed: %v", err)
	}

	var parsed helloShop
	if err := xml.Unmarshal(output.Bytes(), &parsed); err != nil {
		t.Fatalf("hello feed is not well-formed XML: %v", err)
	}

	if parsed.Item.Code != "HELLO-001" {
		t.Fatalf("expected hello item code, got %q", parsed.Item.Code)
	}
}

package feed

import (
	"bytes"
	"context"
	"encoding/xml"
	"testing"
)

type helloShop struct {
	Item helloShopItem `xml:"SHOPITEM"`
}

type helloShopItem struct {
	Code     string `xml:"CODE"`
	Name     string `xml:"NAME"`
	PriceVAT string `xml:"PRICE_VAT"`
	Stock    struct {
		Amount string `xml:"AMOUNT"`
	} `xml:"STOCK"`
}

func TestHelloGenerateReturnsWellFormedXML(t *testing.T) {
	var output bytes.Buffer

	result, err := (Hello{}).Generate(context.Background(), &output)
	if err != nil {
		t.Fatalf("generate hello feed: %v", err)
	}
	if result.ItemsProcessed != 1 {
		t.Fatalf("expected 1 processed item, got %d", result.ItemsProcessed)
	}

	var parsed helloShop
	if err := xml.Unmarshal(output.Bytes(), &parsed); err != nil {
		t.Fatalf("hello feed is not well-formed XML: %v", err)
	}

	if parsed.Item.Code != "HELLO-001" {
		t.Fatalf("expected hello item code, got %q", parsed.Item.Code)
	}
	if parsed.Item.Name != "Hello world product" {
		t.Fatalf("expected hello item name, got %q", parsed.Item.Name)
	}
	if parsed.Item.PriceVAT != "123" {
		t.Fatalf("expected hello item price, got %q", parsed.Item.PriceVAT)
	}
	if parsed.Item.Stock.Amount != "7" {
		t.Fatalf("expected hello item stock, got %#v", parsed.Item.Stock)
	}
}

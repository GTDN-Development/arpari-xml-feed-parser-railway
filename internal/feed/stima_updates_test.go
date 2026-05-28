package feed

import (
	"bytes"
	"context"
	"encoding/xml"
	"testing"
)

func TestStimaStockGenerateUsesFixtureBackedDownloader(t *testing.T) {
	downloader := &fakeStimaDownloader{
		body: `<SHOP>
  <SHOPITEM>
    <VARIANTS>
      <VARIANT>
        <CODE>ART13627-k001</CODE>
        <STOCK>
          <WAREHOUSES>
            <WAREHOUSE><NAME>HLAVNÍ SKLAD</NAME><VALUE>7.00</VALUE></WAREHOUSE>
          </WAREHOUSES>
        </STOCK>
      </VARIANT>
    </VARIANTS>
  </SHOPITEM>
</SHOP>`,
	}
	generator := StimaStock{
		Downloader: downloader,
		SourceURL:  "https://example.test/stima-stock.xml",
	}

	var output bytes.Buffer
	result, err := generator.Generate(context.Background(), &output)
	if err != nil {
		t.Fatalf("generate STIMA stock: %v", err)
	}
	if downloader.lastURL != "https://example.test/stima-stock.xml" {
		t.Fatalf("expected custom source URL, got %q", downloader.lastURL)
	}
	if result.ItemsProcessed != 1 || result.ItemsSkipped != 0 {
		t.Fatalf("unexpected result: %#v", result)
	}

	parsed := parseGeneratedUpdate(t, output.Bytes())
	if len(parsed.Items) != 1 || parsed.Items[0].Code != "ART13627" {
		t.Fatalf("unexpected generated items: %#v", parsed.Items)
	}
	variant := parsed.Items[0].Variants[0]
	if variant.Code != "ART13627-k001" || variant.PriceVAT != "" || len(variant.Stock.Warehouses) != 1 || variant.Stock.Warehouses[0].Value != "7.00" {
		t.Fatalf("unexpected generated stock variant: %#v", variant)
	}
}

func TestStimaStockPriceGenerateUsesFixtureBackedDownloader(t *testing.T) {
	downloader := &fakeStimaDownloader{
		body: `<SHOP>
  <SHOPITEM>
    <CODE>ART-SIMPLE</CODE>
    <PRICE_VAT>12900.00</PRICE_VAT>
    <STOCK>
      <WAREHOUSES>
        <WAREHOUSE><NAME>HLAVNÍ SKLAD</NAME><VALUE>3.000</VALUE></WAREHOUSE>
      </WAREHOUSES>
    </STOCK>
  </SHOPITEM>
</SHOP>`,
	}
	generator := StimaStockPrice{
		Downloader: downloader,
		SourceURL:  "https://example.test/stima-stock-price.xml",
	}

	var output bytes.Buffer
	result, err := generator.Generate(context.Background(), &output)
	if err != nil {
		t.Fatalf("generate STIMA stock-price: %v", err)
	}
	if downloader.lastURL != "https://example.test/stima-stock-price.xml" {
		t.Fatalf("expected custom source URL, got %q", downloader.lastURL)
	}
	if result.ItemsProcessed != 1 || result.ItemsSkipped != 0 {
		t.Fatalf("unexpected result: %#v", result)
	}

	parsed := parseGeneratedUpdate(t, output.Bytes())
	if len(parsed.Items) != 1 || parsed.Items[0].Code != "ART-SIMPLE" || parsed.Items[0].PriceVAT != "12900.00" {
		t.Fatalf("unexpected generated item: %#v", parsed.Items)
	}
	if len(parsed.Items[0].Stock.Warehouses) != 1 || parsed.Items[0].Stock.Warehouses[0].Value != "3.000" {
		t.Fatalf("unexpected generated stock: %#v", parsed.Items[0].Stock)
	}
}

func parseGeneratedUpdate(t *testing.T, data []byte) generatedUpdateShop {
	t.Helper()
	var parsed generatedUpdateShop
	if err := xml.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("generated output is not XML: %v", err)
	}
	return parsed
}

type generatedUpdateShop struct {
	Items []generatedUpdateItem `xml:"SHOPITEM"`
}

type generatedUpdateItem struct {
	Code     string                   `xml:"CODE"`
	PriceVAT string                   `xml:"PRICE_VAT"`
	Stock    generatedUpdateStock     `xml:"STOCK"`
	Variants []generatedUpdateVariant `xml:"VARIANTS>VARIANT"`
}

type generatedUpdateVariant struct {
	Code     string               `xml:"CODE"`
	PriceVAT string               `xml:"PRICE_VAT"`
	Stock    generatedUpdateStock `xml:"STOCK"`
}

type generatedUpdateStock struct {
	Value      string                     `xml:",chardata"`
	Warehouses []generatedUpdateWarehouse `xml:"WAREHOUSES>WAREHOUSE"`
}

type generatedUpdateWarehouse struct {
	Name  string `xml:"NAME"`
	Value string `xml:"VALUE"`
}

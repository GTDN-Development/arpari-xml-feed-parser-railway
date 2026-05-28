package feed

import (
	"bytes"
	"context"
	"encoding/xml"
	"testing"
)

func TestAutronicProductsTestGenerateLimitsOutputItems(t *testing.T) {
	downloader := &fakeStimaDownloader{
		body: `<ProductFeed><Products>
  <Product><ProductCode>NA-1</ProductCode><ProductName>Židle 1</ProductName><ProductCategory><CategoryName>Čalouněné židle</CategoryName><CategoryShortName>NA-ZID-CAL</CategoryShortName></ProductCategory></Product>
  <Product><ProductCode>NA-2</ProductCode><ProductName>Židle 2</ProductName><ProductCategory><CategoryName>Čalouněné židle</CategoryName><CategoryShortName>NA-ZID-CAL</CategoryShortName></ProductCategory></Product>
</Products></ProductFeed>`,
	}
	generator := AutronicProductsTest{
		Downloader:  downloader,
		SourceURL:   "https://example.test/autronic-products.xml",
		MaxProducts: 1,
	}

	output := generateForTest(t, generator)
	items := parseGeneratedCodes(t, output)
	if downloader.lastURL != "https://example.test/autronic-products.xml" {
		t.Fatalf("expected custom source URL, got %q", downloader.lastURL)
	}
	if len(items) != 1 || items[0].Code != "NA-1" {
		t.Fatalf("unexpected generated items: %#v", items)
	}
}

func TestSegoTestGenerateLimitsOutputItems(t *testing.T) {
	downloader := &fakeStimaDownloader{
		body: `<SHOP>
  <SHOPITEM><ITEM_ID>1</ITEM_ID><PRODUCTNAME>Židle 1</PRODUCTNAME></SHOPITEM>
  <SHOPITEM><ITEM_ID>2</ITEM_ID><PRODUCTNAME>Židle 2</PRODUCTNAME></SHOPITEM>
</SHOP>`,
	}
	generator := SegoTest{
		Downloader:  downloader,
		SourceURL:   "https://example.test/sego.xml",
		MaxProducts: 1,
	}

	output := generateForTest(t, generator)
	items := parseGeneratedCodes(t, output)
	if downloader.lastURL != "https://example.test/sego.xml" {
		t.Fatalf("expected custom source URL, got %q", downloader.lastURL)
	}
	if len(items) != 1 || items[0].Code != "1" {
		t.Fatalf("unexpected generated items: %#v", items)
	}
}

func TestHonTestGenerateLimitsOutputItems(t *testing.T) {
	downloader := &fakeStimaDownloader{
		body: `<SHOP>
  <SHOPITEM><ID>1</ID><PRODUCT>Produkt 1</PRODUCT></SHOPITEM>
  <SHOPITEM><ID>2</ID><PRODUCT>Produkt 2</PRODUCT></SHOPITEM>
</SHOP>`,
	}
	generator := HonTest{
		Downloader:  downloader,
		SourceURL:   "https://example.test/hon.xml",
		MaxProducts: 1,
	}

	output := generateForTest(t, generator)
	items := parseGeneratedCodes(t, output)
	if downloader.lastURL != "https://example.test/hon.xml" {
		t.Fatalf("expected custom source URL, got %q", downloader.lastURL)
	}
	if len(items) != 1 || items[0].Code != "1" {
		t.Fatalf("unexpected generated items: %#v", items)
	}
}

func generateForTest(t *testing.T, generator Generator) []byte {
	t.Helper()
	var output bytes.Buffer
	result, err := generator.Generate(context.Background(), &output)
	if err != nil {
		t.Fatalf("generate %s: %v", generator.Name(), err)
	}
	if result.ItemsProcessed != 1 || result.ItemsSkipped != 0 {
		t.Fatalf("unexpected result for %s: %#v", generator.Name(), result)
	}
	return output.Bytes()
}

func parseGeneratedCodes(t *testing.T, data []byte) []struct {
	Code string `xml:"CODE"`
} {
	t.Helper()
	var parsed struct {
		Items []struct {
			Code string `xml:"CODE"`
		} `xml:"SHOPITEM"`
	}
	if err := xml.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("generated output is not XML: %v", err)
	}
	return parsed.Items
}

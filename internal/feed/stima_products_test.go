package feed

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
	"testing"
)

type fakeStimaDownloader struct {
	body    string
	lastURL string
}

func TestStimaProductsTestGenerateLimitsOutputItems(t *testing.T) {
	var source strings.Builder
	source.WriteString("<SHOP>")
	for index := 0; index < 25; index++ {
		fmt.Fprintf(&source, `<SHOPITEM><NAME>Židle Fixture %02d</NAME><CODE>ART-TEST-%02d</CODE><PRICE_VAT>1000.00</PRICE_VAT></SHOPITEM>`, index, index)
	}
	source.WriteString("</SHOP>")

	downloader := &fakeStimaDownloader{
		body: source.String(),
	}
	generator := StimaProductsTest{
		Downloader:  downloader,
		SourceURL:   "https://example.test/stima-products.xml",
		MaxProducts: 3,
	}

	var output bytes.Buffer
	result, err := generator.Generate(context.Background(), &output)
	if err != nil {
		t.Fatalf("generate limited STIMA products: %v", err)
	}
	if result.ItemsProcessed != 3 || result.ItemsSkipped != 0 {
		t.Fatalf("unexpected result: %#v", result)
	}

	var parsed struct {
		Items []struct {
			Code string `xml:"CODE"`
		} `xml:"SHOPITEM"`
	}
	if err := xml.Unmarshal(output.Bytes(), &parsed); err != nil {
		t.Fatalf("generated output is not XML: %v", err)
	}
	if len(parsed.Items) != 3 {
		t.Fatalf("expected 3 SHOPITEM elements, got %d", len(parsed.Items))
	}
	if parsed.Items[2].Code != "ART-TEST-02" {
		t.Fatalf("expected source order to be preserved, got %#v", parsed.Items)
	}
}

func (downloader *fakeStimaDownloader) Download(_ context.Context, url string) (io.ReadCloser, error) {
	downloader.lastURL = url
	return io.NopCloser(strings.NewReader(downloader.body)), nil
}

func TestStimaProductsGenerateUsesFixtureBackedDownloader(t *testing.T) {
	downloader := &fakeStimaDownloader{
		body: `<SHOP>
  <SHOPITEM>
    <NAME>Židle Fixture</NAME>
    <VARIANTS>
      <VARIANT>
        <CODE>ART13627-k001</CODE>
        <PRICE_VAT>1000.00</PRICE_VAT>
        <PARAMETERS>
          <PARAMETER><NAME>KOSTRA</NAME><VALUE>dub</VALUE></PARAMETER>
        </PARAMETERS>
      </VARIANT>
    </VARIANTS>
  </SHOPITEM>
</SHOP>`,
	}
	generator := StimaProducts{
		Downloader: downloader,
		SourceURL:  "https://example.test/stima-products.xml",
	}

	var output bytes.Buffer
	result, err := generator.Generate(context.Background(), &output)
	if err != nil {
		t.Fatalf("generate STIMA products: %v", err)
	}
	if downloader.lastURL != "https://example.test/stima-products.xml" {
		t.Fatalf("expected custom source URL, got %q", downloader.lastURL)
	}
	if result.ItemsProcessed != 1 || result.ItemsSkipped != 0 {
		t.Fatalf("unexpected result: %#v", result)
	}

	var parsed struct {
		Items []struct {
			ExternalID string `xml:"EXTERNAL_ID"`
			Code       string `xml:"CODE"`
			Variants   []struct {
				Code       string `xml:"CODE"`
				Parameters []struct {
					Name  string `xml:"NAME"`
					Value string `xml:"VALUE"`
				} `xml:"PARAMETERS>PARAMETER"`
			} `xml:"VARIANTS>VARIANT"`
		} `xml:"SHOPITEM"`
	}
	if err := xml.Unmarshal(output.Bytes(), &parsed); err != nil {
		t.Fatalf("generated output is not XML: %v", err)
	}
	if len(parsed.Items) != 1 || parsed.Items[0].ExternalID != "ART13627" || parsed.Items[0].Code != "" {
		t.Fatalf("unexpected generated items: %#v", parsed.Items)
	}
	if len(parsed.Items[0].Variants) != 1 || parsed.Items[0].Variants[0].Code != "ART13627-k001" {
		t.Fatalf("unexpected generated variants: %#v", parsed.Items[0].Variants)
	}
	if len(parsed.Items[0].Variants[0].Parameters) != 1 || parsed.Items[0].Variants[0].Parameters[0].Name != "KOSTRA" {
		t.Fatalf("unexpected generated parameters: %#v", parsed.Items[0].Variants[0].Parameters)
	}
}

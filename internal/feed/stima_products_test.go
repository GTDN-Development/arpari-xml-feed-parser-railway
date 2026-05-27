package feed

import (
	"bytes"
	"context"
	"encoding/xml"
	"io"
	"strings"
	"testing"
)

type fakeStimaDownloader struct {
	body    string
	lastURL string
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
			Code     string `xml:"CODE"`
			Variants []struct {
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
	if len(parsed.Items) != 1 || parsed.Items[0].Code != "ART13627" {
		t.Fatalf("unexpected generated items: %#v", parsed.Items)
	}
	if len(parsed.Items[0].Variants) != 1 || parsed.Items[0].Variants[0].Code != "ART13627-k001" {
		t.Fatalf("unexpected generated variants: %#v", parsed.Items[0].Variants)
	}
	if len(parsed.Items[0].Variants[0].Parameters) != 1 || parsed.Items[0].Variants[0].Parameters[0].Name != "KOSTRA" {
		t.Fatalf("unexpected generated parameters: %#v", parsed.Items[0].Variants[0].Parameters)
	}
}

package feed

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
)

func TestHonGenerateOmitsCategoriesToPreserveManualAssignments(t *testing.T) {
	downloader := &fakeHonDownloader{
		body: `<SHOP>
  <SHOPITEM>
    <ID>812660</ID>
    <MAIN_CATEGORY>Židle jednací OfficePro</MAIN_CATEGORY>
    <PRODUCT>TRITON NET</PRODUCT>
    <PRICE_VAT>1606.00</PRICE_VAT>
    <DOSTUPNOST>Skladem</DOSTUPNOST>
    <STOCK>7.0</STOCK>
    <PART_NUMBER>DY30020001-017043</PART_NUMBER>
    <DESCRIPTION>židle jednací</DESCRIPTION>
  </SHOPITEM>
</SHOP>`,
	}
	generator := Hon{
		Downloader: downloader,
		SourceURL:  "https://example.test/hon.xml",
	}

	var output bytes.Buffer
	result, err := generator.Generate(context.Background(), &output)
	if err != nil {
		t.Fatalf("generate HON feed: %v", err)
	}
	if downloader.lastURL != "https://example.test/hon.xml" {
		t.Fatalf("expected custom source URL, got %q", downloader.lastURL)
	}
	if result.ItemsProcessed != 1 || result.ItemsSkipped != 0 {
		t.Fatalf("unexpected result: %#v", result)
	}

	xml := output.String()
	if strings.Contains(xml, "<CATEGORIES>") || strings.Contains(xml, "KONFERENČNÍ ŽIDLE") {
		t.Fatalf("expected generated HON feed to omit categories, got:\n%s", xml)
	}
	if !strings.Contains(xml, "<CODE>DY30020001-017043</CODE>") || !strings.Contains(xml, "<SUPPLIER>HON</SUPPLIER>") {
		t.Fatalf("expected product identity fields, got:\n%s", xml)
	}
}

type fakeHonDownloader struct {
	body    string
	lastURL string
}

func (downloader *fakeHonDownloader) Download(_ context.Context, url string) (io.ReadCloser, error) {
	downloader.lastURL = url
	return io.NopCloser(strings.NewReader(downloader.body)), nil
}

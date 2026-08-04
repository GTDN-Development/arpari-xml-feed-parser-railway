package feed

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"testing"
)

func TestAutronicAvailabilityGenerateUsesCatalogAndAvailabilitySources(t *testing.T) {
	productsURL := "https://example.test/autronic-products.xml"
	availabilityURL := "https://example.test/autronic-availability.xml"
	downloader := &fakeAutronicDownloader{
		bodies: map[string]string{
			productsURL: `<ProductFeed><Products>
  <Product>
    <ProductCode>NA-CHAIR-1</ProductCode>
    <ProductName>Kancelářská židle</ProductName>
    <ProductCategory><CategoryName>Síťované kancelářské židle</CategoryName><CategoryShortName>NA-ZKA-SIT</CategoryShortName></ProductCategory>
	<Prices><RetailPromotionalPriceIncludingVat value="1790.00" /></Prices>
  </Product>
</Products></ProductFeed>`,
			availabilityURL: `<ProductFeed><Products>
  <Product>
    <ProductCode>NA-CHAIR-1</ProductCode>
    <Availability>
      <StockAvailabilityTotal Quantity="7" />
      <StockAvailability><Stock Name="Semčice" Quantity="7" /></StockAvailability>
    </Availability>
  </Product>
</Products></ProductFeed>`,
		},
	}
	generator := AutronicAvailability{
		Downloader:            downloader,
		ProductsSourceURL:     productsURL,
		AvailabilitySourceURL: availabilityURL,
	}

	var output bytes.Buffer
	result, err := generator.Generate(context.Background(), &output)
	if err != nil {
		t.Fatalf("generate Autronic availability: %v", err)
	}
	if result.ItemsProcessed != 1 || result.ItemsSkipped != 0 {
		t.Fatalf("unexpected result: %#v", result)
	}
	if strings.Join(downloader.urls, "\n") != productsURL+"\n"+availabilityURL {
		t.Fatalf("unexpected downloaded URLs: %#v", downloader.urls)
	}

	parsed := parseGeneratedUpdate(t, output.Bytes())
	if len(parsed.Items) != 1 || parsed.Items[0].Code != "NA-CHAIR-1" {
		t.Fatalf("unexpected generated items: %#v", parsed.Items)
	}
	if parsed.Items[0].PriceVAT != "1790" {
		t.Fatalf("unexpected generated price: %#v", parsed.Items[0])
	}
	if len(parsed.Items[0].Stock.Warehouses) != 1 || parsed.Items[0].Stock.Warehouses[0].Name != "Semčice" || parsed.Items[0].Stock.Warehouses[0].Value != "7" {
		t.Fatalf("unexpected generated stock: %#v", parsed.Items[0].Stock)
	}
}

type fakeAutronicDownloader struct {
	bodies map[string]string
	urls   []string
}

func (downloader *fakeAutronicDownloader) Download(_ context.Context, url string) (io.ReadCloser, error) {
	downloader.urls = append(downloader.urls, url)
	body, ok := downloader.bodies[url]
	if !ok {
		return nil, fmt.Errorf("unexpected URL %s", url)
	}
	return io.NopCloser(strings.NewReader(body)), nil
}

package autronic

import (
	"context"
	"strings"
	"testing"

	"github.com/fanda/arpari-xml-feed-parser-railway/internal/shoptet"
)

func TestParseProductsFiltersFurnitureAndMapsFields(t *testing.T) {
	input := `<?xml version="1.0" encoding="utf-8"?>
<ProductFeed>
  <Products>
    <Product>
      <ProductCode>NA-CHAIR-1</ProductCode>
      <ProductName>Kancelářská židle</ProductName>
      <ProductCategory>
        <CategoryName>Síťované kancelářské židle</CategoryName>
        <CategoryShortName>NA-ZKA-SIT</CategoryShortName>
      </ProductCategory>
      <Ean>859000000001</Ean>
      <Prices>
        <RetailPriceIncludingVat currency="CZK" value="1990.00" />
        <RetailPromotionalPriceIncludingVat currency="CZK" value="1790.00" />
      </Prices>
      <Availability>
        <AvailabilityStatus>InStock</AvailabilityStatus>
        <StockAvailabilityTotal Quantity="3" />
        <StockAvailability>
          <Stock Name="Semčice" Quantity="3" />
        </StockAvailability>
      </Availability>
      <Descriptions>
        <Description format="html">&lt;p&gt;Popis židle&lt;/p&gt;</Description>
      </Descriptions>
      <Images>
        <Image mediumSizeUrl="https://example.test/small.jpg" largeSizeUrl="https://example.test/large.jpg" />
      </Images>
      <Parameters>
        <Parameter type="Text">
          <Name>Barva</Name>
          <TextValue>Bílá</TextValue>
        </Parameter>
        <Parameter type="Numeric">
          <Name>Výška (cm)</Name>
          <NumericValue>108.0000</NumericValue>
          <Unit>cm</Unit>
        </Parameter>
        <Parameter type="Numeric">
          <Name>Počet balení</Name>
          <NumericValue>1.0000</NumericValue>
          <Unit>ks</Unit>
        </Parameter>
      </Parameters>
    </Product>
    <Product>
      <ProductCode>DE-FLOWER-1</ProductCode>
      <ProductName>Dekorace</ProductName>
      <ProductCategory>
        <CategoryName>Řezané</CategoryName>
        <CategoryShortName>DE-UMKV-REZANE</CategoryShortName>
      </ProductCategory>
    </Product>
  </Products>
</ProductFeed>`

	feed, stats, err := ParseProducts(context.Background(), strings.NewReader(input), ProductsOptions{})
	if err != nil {
		t.Fatalf("parse Autronic products: %v", err)
	}
	if stats.ProductsRead != 2 || stats.ProductsEmitted != 1 || stats.ProductsSkipped != 1 {
		t.Fatalf("unexpected stats: %#v", stats)
	}
	if len(feed.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(feed.Items))
	}

	item := feed.Items[0]
	if item.Code != "NA-CHAIR-1" || item.PriceVAT != "1790.00" || item.EAN != "859000000001" {
		t.Fatalf("unexpected item fields: %#v", item)
	}
	if item.Stock != "3" || len(item.Warehouses) != 1 || item.Warehouses[0].Name != "Semčice" {
		t.Fatalf("unexpected stock: %#v", item)
	}
	if len(item.Images) != 1 || item.Images[0].URL != "https://example.test/large.jpg" {
		t.Fatalf("unexpected images: %#v", item.Images)
	}
	if item.DefaultCategory == nil || item.DefaultCategory.ID != "1284" {
		t.Fatalf("unexpected category: %#v", item.DefaultCategory)
	}
	expectedParameters := []shoptet.Parameter{
		{Name: "Barva", Value: "Bílá"},
		{Name: "Výška (cm)", Value: "108"},
		{Name: "Počet balení", Value: "1 ks"},
	}
	if len(item.InformationParameters) != len(expectedParameters) {
		t.Fatalf("expected information parameters, got %#v", item.InformationParameters)
	}
	for index, expected := range expectedParameters {
		if item.InformationParameters[index] != expected {
			t.Fatalf("unexpected information parameter %d: %#v", index, item.InformationParameters[index])
		}
	}
}

func TestParseProductsLimitsTestOutput(t *testing.T) {
	input := `<ProductFeed><Products>
  <Product><ProductCode>NA-1</ProductCode><ProductName>Židle 1</ProductName><ProductCategory><CategoryName>Čalouněné židle</CategoryName><CategoryShortName>NA-ZID-CAL</CategoryShortName></ProductCategory></Product>
  <Product><ProductCode>NA-2</ProductCode><ProductName>Židle 2</ProductName><ProductCategory><CategoryName>Čalouněné židle</CategoryName><CategoryShortName>NA-ZID-CAL</CategoryShortName></ProductCategory></Product>
</Products></ProductFeed>`

	feed, stats, err := ParseProducts(context.Background(), strings.NewReader(input), ProductsOptions{MaxProducts: 1})
	if err != nil {
		t.Fatalf("parse limited Autronic products: %v", err)
	}
	if stats.ProductsRead != 1 || stats.ProductsEmitted != 1 {
		t.Fatalf("unexpected stats: %#v", stats)
	}
	if len(feed.Items) != 1 || feed.Items[0].Code != "NA-1" {
		t.Fatalf("unexpected feed: %#v", feed)
	}
}

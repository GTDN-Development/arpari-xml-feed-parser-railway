package autronic

import (
	"context"
	"strings"
	"testing"

	"github.com/fanda/arpari-xml-feed-parser-railway/internal/shoptet"
)

func TestParseAvailabilityFiltersToCatalogAndMapsStock(t *testing.T) {
	catalog := `<ProductFeed><Products>
  <Product>
    <ProductCode>NA-CHAIR-1</ProductCode>
    <ProductName>Kancelářská židle</ProductName>
    <ProductCategory><CategoryName>Síťované kancelářské židle</CategoryName><CategoryShortName>NA-ZKA-SIT</CategoryShortName></ProductCategory>
	<Prices><RetailPriceIncludingVat value="1990.00" /><RetailPromotionalPriceIncludingVat value="1790.00" /></Prices>
  </Product>
  <Product>
    <ProductCode>NA-MISSING</ProductCode>
    <ProductName>Křeslo bez skladového záznamu</ProductName>
    <ProductCategory><CategoryName>Křesla</CategoryName><CategoryShortName>NA-KRE-KT</CategoryShortName></ProductCategory>
  </Product>
  <Product>
    <ProductCode>DE-FLOWER-1</ProductCode>
    <ProductName>Dekorace</ProductName>
    <ProductCategory><CategoryName>Dekorace</CategoryName><CategoryShortName>DE-UMKV-REZANE</CategoryShortName></ProductCategory>
  </Product>
</Products></ProductFeed>`
	availability := `<ProductFeed><Products>
  <Product>
    <ProductCode>NA-CHAIR-1</ProductCode>
    <Availability>
      <AvailabilityStatus>InStock</AvailabilityStatus>
      <StockAvailabilityTotal Quantity="3" />
      <StockAvailability>
        <Stock Name="Semčice" Quantity="2" />
        <Stock Name="Loděnice" Quantity="1" />
      </StockAvailability>
    </Availability>
  </Product>
  <Product>
    <ProductCode>DE-FLOWER-1</ProductCode>
    <Availability><StockAvailabilityTotal Quantity="9" /></Availability>
  </Product>
  <Product>
    <Availability><StockAvailabilityTotal Quantity="1" /></Availability>
  </Product>
</Products></ProductFeed>`

	feed, stats, err := ParseAvailability(context.Background(), strings.NewReader(catalog), strings.NewReader(availability), AvailabilityOptions{})
	if err != nil {
		t.Fatalf("parse Autronic availability: %v", err)
	}
	if stats.CatalogProductsRead != 3 || stats.CatalogProductsEmitted != 2 || stats.CatalogProductsSkipped != 1 {
		t.Fatalf("unexpected catalog stats: %#v", stats)
	}
	if stats.ProductsRead != 3 || stats.ProductsEmitted != 1 || stats.ProductsSkipped != 2 || stats.AvailabilityRecordsUsed != 1 {
		t.Fatalf("unexpected availability stats: %#v", stats)
	}
	if len(feed.Items) != 1 {
		t.Fatalf("expected 1 update item, got %#v", feed.Items)
	}

	item := feed.Items[0]
	if item.Code != "NA-CHAIR-1" || item.PriceVAT != "1790.00" || item.Stock != "3" {
		t.Fatalf("unexpected item stock update: %#v", item)
	}
	expectedWarehouses := []shoptet.Warehouse{
		{Name: "Semčice", Value: "2"},
		{Name: "Loděnice", Value: "1"},
	}
	if len(item.Warehouses) != len(expectedWarehouses) {
		t.Fatalf("unexpected warehouses: %#v", item.Warehouses)
	}
	for index, expected := range expectedWarehouses {
		if item.Warehouses[index] != expected {
			t.Fatalf("unexpected warehouse %d: %#v", index, item.Warehouses[index])
		}
	}
	if item.Name != "" || len(item.Images) != 0 || item.DefaultCategory != nil {
		t.Fatalf("stock-price update must not carry unrelated catalog fields: %#v", item)
	}
}

func TestParseAvailabilityGroupsVariantStockWithParameters(t *testing.T) {
	catalog := `<ProductFeed><Products>
  <Product>
    <ProductCode>CHAIR-BK</ProductCode>
    <ProductName>Jídelní židle, černá, CHAIR-BK</ProductName>
    <ProductCategory><CategoryName>Čalouněné židle</CategoryName><CategoryShortName>NA-ZID-CAL</CategoryShortName></ProductCategory>
	<Prices><RetailPromotionalPriceIncludingVat value="1490.00" /></Prices>
    <Availability><StockAvailabilityTotal Quantity="1" /></Availability>
    <Parameters><Parameter><Name>Barva</Name><TextValue>Černá</TextValue></Parameter></Parameters>
    <ColorVariants><Product><ProductCode>CHAIR-WT</ProductCode><Color>Bílá</Color></Product></ColorVariants>
  </Product>
  <Product>
    <ProductCode>CHAIR-WT</ProductCode>
    <ProductName>Jídelní židle, bílá, CHAIR-WT</ProductName>
    <ProductCategory><CategoryName>Čalouněné židle</CategoryName><CategoryShortName>NA-ZID-CAL</CategoryShortName></ProductCategory>
	<Prices><RetailPriceIncludingVat value="1590.00" /></Prices>
    <Availability><StockAvailabilityTotal Quantity="1" /></Availability>
    <Parameters><Parameter><Name>Barva</Name><TextValue>Bílá</TextValue></Parameter></Parameters>
    <ColorVariants><Product><ProductCode>CHAIR-BK</ProductCode><Color>Černá</Color></Product></ColorVariants>
  </Product>
</Products></ProductFeed>`
	availability := `<ProductFeed><Products>
  <Product><ProductCode>CHAIR-BK</ProductCode><Availability><StockAvailabilityTotal Quantity="4" /></Availability></Product>
  <Product><ProductCode>CHAIR-WT</ProductCode><Availability><StockAvailabilityTotal Quantity="0" /></Availability></Product>
</Products></ProductFeed>`

	feed, stats, err := ParseAvailability(context.Background(), strings.NewReader(catalog), strings.NewReader(availability), AvailabilityOptions{})
	if err != nil {
		t.Fatalf("parse Autronic availability: %v", err)
	}
	if stats.ProductsRead != 2 || stats.ProductsEmitted != 1 || stats.ItemsWithVariants != 1 || stats.VariantsEmitted != 2 {
		t.Fatalf("unexpected stats: %#v", stats)
	}
	if len(feed.Items) != 1 || feed.Items[0].Code != "" {
		t.Fatalf("expected one variant parent without top-level code, got %#v", feed.Items)
	}

	variants := feed.Items[0].Variants
	if len(variants) != 2 {
		t.Fatalf("expected 2 variants, got %#v", variants)
	}
	if variants[0].Code != "CHAIR-BK" || variants[0].PriceVAT != "1490.00" || variants[0].Stock != "4" || variants[1].Code != "CHAIR-WT" || variants[1].PriceVAT != "1590.00" || variants[1].Stock != "0" {
		t.Fatalf("unexpected variant stock: %#v", variants)
	}
	if len(variants[0].Parameters) != 1 || variants[0].Parameters[0] != (shoptet.Parameter{Name: "Barva", Value: "Černá"}) {
		t.Fatalf("variant parameters must be preserved for Shoptet schema: %#v", variants[0].Parameters)
	}
	if len(variants[1].Parameters) != 1 || variants[1].Parameters[0] != (shoptet.Parameter{Name: "Barva", Value: "Bílá"}) {
		t.Fatalf("variant parameters must be preserved for Shoptet schema: %#v", variants[1].Parameters)
	}
}

func TestParseAvailabilitySkipsProductsWithoutStockPayload(t *testing.T) {
	catalog := `<ProductFeed><Products>
  <Product>
    <ProductCode>NA-CHAIR-1</ProductCode>
    <ProductName>Kancelářská židle</ProductName>
    <ProductCategory><CategoryName>Síťované kancelářské židle</CategoryName><CategoryShortName>NA-ZKA-SIT</CategoryShortName></ProductCategory>
  </Product>
</Products></ProductFeed>`
	availability := `<ProductFeed><Products>
  <Product><ProductCode>NA-CHAIR-1</ProductCode><Availability /></Product>
</Products></ProductFeed>`

	feed, stats, err := ParseAvailability(context.Background(), strings.NewReader(catalog), strings.NewReader(availability), AvailabilityOptions{})
	if err != nil {
		t.Fatalf("parse Autronic availability: %v", err)
	}
	if len(feed.Items) != 0 || stats.ProductsRead != 1 || stats.ProductsSkipped != 1 {
		t.Fatalf("expected empty update with skipped payload, feed=%#v stats=%#v", feed, stats)
	}
}

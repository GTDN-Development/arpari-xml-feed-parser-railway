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
	if item.Supplier != "Autronic" {
		t.Fatalf("unexpected supplier: %q", item.Supplier)
	}
	if item.Manufacturer != "Autronic" {
		t.Fatalf("unexpected manufacturer fallback: %q", item.Manufacturer)
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
	if stats.ProductsRead != 2 || stats.ProductsEmitted != 1 {
		t.Fatalf("unexpected stats: %#v", stats)
	}
	if len(feed.Items) != 1 || feed.Items[0].Code != "NA-1" {
		t.Fatalf("unexpected feed: %#v", feed)
	}
}

func TestTransformImagesKeepsFullAutronicGallery(t *testing.T) {
	const totalImages = 12
	sourceImages := make([]sourceImage, 0, totalImages)
	for index := range totalImages {
		sourceImages = append(sourceImages, sourceImage{LargeURL: "https://example.test/image-" + string(rune('a'+index)) + ".jpg"})
	}

	images := transformImages(sourceImages)
	if len(images) != totalImages {
		t.Fatalf("expected %d images, got %#v", totalImages, images)
	}
	if images[0].URL != "https://example.test/image-a.jpg" || images[totalImages-1].URL != "https://example.test/image-l.jpg" {
		t.Fatalf("unexpected image order: %#v", images)
	}
}

func TestMergeGroupImagesKeepsFullAutronicGallery(t *testing.T) {
	entries := []productEntry{{
		Item: shoptet.Item{Images: []shoptet.Image{
			{URL: "https://example.test/image-01.jpg"},
			{URL: "https://example.test/image-02.jpg"},
			{URL: "https://example.test/image-03.jpg"},
			{URL: "https://example.test/image-04.jpg"},
			{URL: "https://example.test/image-05.jpg"},
			{URL: "https://example.test/image-06.jpg"},
		}},
	}, {
		Item: shoptet.Item{Images: []shoptet.Image{
			{URL: "https://example.test/image-07.jpg"},
			{URL: "https://example.test/image-08.jpg"},
			{URL: "https://example.test/image-09.jpg"},
			{URL: "https://example.test/image-10.jpg"},
			{URL: "https://example.test/image-11.jpg"},
			{URL: "https://example.test/image-12.jpg"},
		}},
	}}

	images := mergeGroupImages(entries)
	if len(images) != 12 {
		t.Fatalf("expected 12 merged images, got %#v", images)
	}
	if images[11].URL != "https://example.test/image-12.jpg" {
		t.Fatalf("unexpected merged image order: %#v", images)
	}
}

func TestParseProductsGroupsColorVariants(t *testing.T) {
	input := `<ProductFeed><Products>
  <Product>
    <ProductCode>CHAIR-BK</ProductCode>
    <ProductName>Jídelní židle, černá, CHAIR-BK</ProductName>
    <ProductCategory><CategoryName>Čalouněné židle</CategoryName><CategoryShortName>NA-ZID-CAL</CategoryShortName></ProductCategory>
    <Ean>859000000001</Ean>
    <Prices><RetailPriceIncludingVat value="1990.00" /></Prices>
    <Availability><StockAvailabilityTotal Quantity="3" /></Availability>
    <Images><Image largeSizeUrl="https://example.test/black.jpg" /></Images>
    <Parameters>
      <Parameter><Name>Barva</Name><TextValue>Černá</TextValue></Parameter>
      <Parameter><Name>Materiál</Name><TextValue>Látka</TextValue></Parameter>
    </Parameters>
    <ColorVariants>
      <Product><ProductCode>CHAIR-WT</ProductCode><Color>Bílá</Color></Product>
    </ColorVariants>
  </Product>
  <Product>
    <ProductCode>CHAIR-WT</ProductCode>
    <ProductName>Jídelní židle, bílá, CHAIR-WT</ProductName>
    <ProductCategory><CategoryName>Čalouněné židle</CategoryName><CategoryShortName>NA-ZID-CAL</CategoryShortName></ProductCategory>
    <Ean>859000000002</Ean>
    <Prices><RetailPriceIncludingVat value="2090.00" /></Prices>
    <Availability><StockAvailabilityTotal Quantity="4" /></Availability>
    <Images><Image largeSizeUrl="https://example.test/white.jpg" /></Images>
    <Parameters>
      <Parameter><Name>Barva</Name><TextValue>Bílá</TextValue></Parameter>
      <Parameter><Name>Materiál</Name><TextValue>Látka</TextValue></Parameter>
    </Parameters>
    <ColorVariants>
      <Product><ProductCode>CHAIR-BK</ProductCode><Color>Černá</Color></Product>
    </ColorVariants>
  </Product>
</Products></ProductFeed>`

	feed, stats, err := ParseProducts(context.Background(), strings.NewReader(input), ProductsOptions{})
	if err != nil {
		t.Fatalf("parse Autronic color variants: %v", err)
	}
	if stats.ProductsRead != 2 || stats.ProductsEmitted != 1 || stats.ItemsWithVariants != 1 || stats.VariantsEmitted != 2 {
		t.Fatalf("unexpected stats: %#v", stats)
	}
	if len(feed.Items) != 1 {
		t.Fatalf("expected 1 grouped item, got %#v", feed.Items)
	}

	item := feed.Items[0]
	if item.Code != "" || item.Name != "Jídelní židle, CHAIR" {
		t.Fatalf("unexpected parent item: %#v", item)
	}
	if item.Manufacturer != "Autronic" {
		t.Fatalf("unexpected parent manufacturer fallback: %q", item.Manufacturer)
	}
	if len(item.Images) != 2 {
		t.Fatalf("expected merged images, got %#v", item.Images)
	}
	if len(item.InformationParameters) != 1 || item.InformationParameters[0] != (shoptet.Parameter{Name: "Materiál", Value: "Látka"}) {
		t.Fatalf("expected parent information parameters without Barva, got %#v", item.InformationParameters)
	}
	if len(item.Variants) != 2 {
		t.Fatalf("expected 2 variants, got %#v", item.Variants)
	}
	if item.Variants[0].Code != "CHAIR-BK" || item.Variants[0].EAN != "859000000001" || item.Variants[0].PriceVAT != "1990.00" || item.Variants[0].Stock != "3" {
		t.Fatalf("unexpected first variant: %#v", item.Variants[0])
	}
	if len(item.Variants[0].Parameters) != 1 || item.Variants[0].Parameters[0] != (shoptet.Parameter{Name: "Barva", Value: "Černá"}) {
		t.Fatalf("unexpected first variant parameters: %#v", item.Variants[0].Parameters)
	}
	if item.Variants[0].ImageRef != "https://example.test/black.jpg" {
		t.Fatalf("unexpected first image ref: %q", item.Variants[0].ImageRef)
	}
}

func TestParseProductsKeepsVariantGroupCodeInParentName(t *testing.T) {
	input := `<ProductFeed><Products>
  <Product>
    <ProductCode>KA-B2361 BK</ProductCode>
    <ProductName>Kancelářská židle, synchronní mechanismus, černá síťovina, KA-B2361 BK</ProductName>
    <ProductCategory><CategoryName>Síťované kancelářské židle</CategoryName><CategoryShortName>NA-ZKA-SIT</CategoryShortName></ProductCategory>
    <Parameters><Parameter><Name>Barva</Name><TextValue>Černá</TextValue></Parameter></Parameters>
    <ColorVariants><Product><ProductCode>KA-B2361 GREY</ProductCode><Color>Šedá</Color></Product></ColorVariants>
  </Product>
  <Product>
    <ProductCode>KA-B2361 GREY</ProductCode>
    <ProductName>Kancelářská židle, synchronní mechanismus, šedá síťovina, KA-B2361 GREY</ProductName>
    <ProductCategory><CategoryName>Síťované kancelářské židle</CategoryName><CategoryShortName>NA-ZKA-SIT</CategoryShortName></ProductCategory>
    <Parameters><Parameter><Name>Barva</Name><TextValue>Šedá</TextValue></Parameter></Parameters>
    <ColorVariants><Product><ProductCode>KA-B2361 BK</ProductCode><Color>Černá</Color></Product></ColorVariants>
  </Product>
  <Product>
    <ProductCode>KA-B2363 BK</ProductCode>
    <ProductName>Kancelářská židle, synchronní mechanismus, černá síťovina, KA-B2363 BK</ProductName>
    <ProductCategory><CategoryName>Síťované kancelářské židle</CategoryName><CategoryShortName>NA-ZKA-SIT</CategoryShortName></ProductCategory>
    <Parameters><Parameter><Name>Barva</Name><TextValue>Černá</TextValue></Parameter></Parameters>
    <ColorVariants><Product><ProductCode>KA-B2363 GREY</ProductCode><Color>Šedá</Color></Product></ColorVariants>
  </Product>
  <Product>
    <ProductCode>KA-B2363 GREY</ProductCode>
    <ProductName>Kancelářská židle, synchronní mechanismus, šedá síťovina, KA-B2363 GREY</ProductName>
    <ProductCategory><CategoryName>Síťované kancelářské židle</CategoryName><CategoryShortName>NA-ZKA-SIT</CategoryShortName></ProductCategory>
    <Parameters><Parameter><Name>Barva</Name><TextValue>Šedá</TextValue></Parameter></Parameters>
    <ColorVariants><Product><ProductCode>KA-B2363 BK</ProductCode><Color>Černá</Color></Product></ColorVariants>
  </Product>
</Products></ProductFeed>`

	feed, _, err := ParseProducts(context.Background(), strings.NewReader(input), ProductsOptions{})
	if err != nil {
		t.Fatalf("parse Autronic variant group codes: %v", err)
	}
	if len(feed.Items) != 2 {
		t.Fatalf("expected 2 grouped items, got %#v", feed.Items)
	}
	if feed.Items[0].Name != "Kancelářská židle, synchronní mechanismus, KA-B2361" {
		t.Fatalf("unexpected first parent name: %q", feed.Items[0].Name)
	}
	if feed.Items[1].Name != "Kancelářská židle, synchronní mechanismus, KA-B2363" {
		t.Fatalf("unexpected second parent name: %q", feed.Items[1].Name)
	}
}

func TestParseProductsDisambiguatesDuplicateColorVariants(t *testing.T) {
	input := `<ProductFeed><Products>
  <Product>
    <ProductCode>CHAIR-BR1</ProductCode>
    <ProductName>Jídelní židle, hnědá, látka, CHAIR-BR1</ProductName>
    <ProductCategory><CategoryName>Čalouněné židle</CategoryName><CategoryShortName>NA-ZID-CAL</CategoryShortName></ProductCategory>
    <Parameters><Parameter><Name>Barva</Name><TextValue>Hnědá</TextValue></Parameter></Parameters>
    <ColorVariants><Product><ProductCode>CHAIR-BR2</ProductCode><Color>Hnědá</Color></Product></ColorVariants>
  </Product>
  <Product>
    <ProductCode>CHAIR-BR2</ProductCode>
    <ProductName>Jídelní židle, hnědá, samet, CHAIR-BR2</ProductName>
    <ProductCategory><CategoryName>Čalouněné židle</CategoryName><CategoryShortName>NA-ZID-CAL</CategoryShortName></ProductCategory>
    <Parameters><Parameter><Name>Barva</Name><TextValue>Hnědá</TextValue></Parameter></Parameters>
    <ColorVariants><Product><ProductCode>CHAIR-BR1</ProductCode><Color>Hnědá</Color></Product></ColorVariants>
  </Product>
</Products></ProductFeed>`

	feed, _, err := ParseProducts(context.Background(), strings.NewReader(input), ProductsOptions{})
	if err != nil {
		t.Fatalf("parse Autronic duplicate color variants: %v", err)
	}
	if len(feed.Items) != 1 || len(feed.Items[0].Variants) != 2 {
		t.Fatalf("expected grouped variants, got %#v", feed.Items)
	}
	if feed.Items[0].Variants[0].Parameters[0].Value != "Hnědá (CHAIR-BR1)" || feed.Items[0].Variants[1].Parameters[0].Value != "Hnědá (CHAIR-BR2)" {
		t.Fatalf("expected disambiguated colors, got %#v", feed.Items[0].Variants)
	}
}

func TestParseProductsAutronicCategorySelection(t *testing.T) {
	input := `<ProductFeed><Products>
  <Product><ProductCode>NA-KOM-1</ProductCode><ProductName>Komoda</ProductName><ProductCategory><CategoryName>Komody</CategoryName><CategoryShortName>NA-KOM</CategoryShortName></ProductCategory></Product>
  <Product><ProductCode>NA-POS-CAL-1</ProductCode><ProductName>Čalouněná postel</ProductName><ProductCategory><CategoryName>Čalouněné postele</CategoryName><CategoryShortName>NA-POS-CAL</CategoryShortName></ProductCategory></Product>
  <Product><ProductCode>BD-BO-1</ProductCode><ProductName>Botník</ProductName><ProductCategory><CategoryName>Botníky</CategoryName><CategoryShortName>BD-BO</CategoryShortName></ProductCategory></Product>
  <Product><ProductCode>BD-ORG-1</ProductCode><ProductName>Organizér</ProductName><ProductCategory><CategoryName>Organizéry</CategoryName><CategoryShortName>BD-ORG</CategoryShortName></ProductCategory></Product>
  <Product><ProductCode>BD-STKV-1</ProductCode><ProductName>Stojan na květiny</ProductName><ProductCategory><CategoryName>Stojany na květiny</CategoryName><CategoryShortName>BD-STKV</CategoryShortName></ProductCategory></Product>
  <Product><ProductCode>NA-ZAH-LEH-1</ProductCode><ProductName>Zahradní lehátko</ProductName><ProductCategory><CategoryName>Zahradní lehátka</CategoryName><CategoryShortName>NA-ZAH-LEH</CategoryShortName></ProductCategory></Product>
  <Product><ProductCode>BD-ZR-1</ProductCode><ProductName>Zrcadlo</ProductName><ProductCategory><CategoryName>Zrcadla</CategoryName><CategoryShortName>BD-ZR</CategoryShortName></ProductCategory></Product>
  <Product><ProductCode>DE-STOL-1</ProductCode><ProductName>Stolování</ProductName><ProductCategory><CategoryName>Stolování</CategoryName><CategoryShortName>DE-STOL-BAMB</CategoryShortName></ProductCategory></Product>
</Products></ProductFeed>`

	feed, stats, err := ParseProducts(context.Background(), strings.NewReader(input), ProductsOptions{})
	if err != nil {
		t.Fatalf("parse Autronic category selection: %v", err)
	}
	if stats.ProductsRead != 8 || stats.ProductsEmitted != 6 || stats.ProductsSkipped != 2 {
		t.Fatalf("unexpected stats: %#v", stats)
	}
	if len(feed.Items) != 6 {
		t.Fatalf("expected 6 emitted items, got %#v", feed.Items)
	}

	expected := []shoptet.Category{
		{ID: "1197", Path: "BYTOVÉ DOPLŇKY > KOMODY"},
		{ID: "1185", Path: "LOŽNICE > POSTELE"},
		{ID: "1200", Path: "BYTOVÉ DOPLŇKY > BOTNÍKY"},
		{ID: "1206", Path: "BYTOVÉ DOPLŇKY > POLIČKY"},
		{ID: "1227", Path: "ZAHRADNÍ NÁBYTEK > ZAHRADNÍ DOPLŃKY"},
		{ID: "1224", Path: "ZAHRADNÍ NÁBYTEK > ZAHRADNÍ LEHÁTKA"},
	}
	for index, category := range expected {
		if feed.Items[index].DefaultCategory == nil || *feed.Items[index].DefaultCategory != category {
			t.Fatalf("unexpected category for item %d: %#v", index, feed.Items[index].DefaultCategory)
		}
	}
}

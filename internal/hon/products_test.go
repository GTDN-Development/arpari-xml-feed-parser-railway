package hon

import (
	"context"
	"strings"
	"testing"

	"github.com/fanda/arpari-xml-feed-parser-railway/internal/shoptet"
)

func TestParseProductsMapsHONItems(t *testing.T) {
	input := `<?xml version="1.0" encoding="utf-8"?>
<SHOP>
  <SHOPITEM>
    <ID>812620</ID>
    <MAIN_CATEGORY>Židle kancelářské HonPlus</MAIN_CATEGORY>
    <PRODUCT>MERENS BP</PRODUCT>
    <PRICE_VAT>10756.90</PRICE_VAT>
    <DOSTUPNOST>Skladem</DOSTUPNOST>
    <IMGURL><IMGURL>https://example.test/merens.jpg</IMGURL></IMGURL>
    <STOCK>74.0</STOCK>
    <PART_NUMBER>DY10010001-010042</PART_NUMBER>
    <DESCRIPTION>černá BI 201, kanc. židle bez podhlavníku</DESCRIPTION>
    <PARAM><PARAM_NAME>Šířka</PARAM_NAME><VAL>20.0</VAL></PARAM>
    <PARAM><PARAM_NAME>Výška</PARAM_NAME><VAL>85.50</VAL></PARAM>
  </SHOPITEM>
  <SHOPITEM>
    <PRODUCT>Bez kódu</PRODUCT>
  </SHOPITEM>
</SHOP>`

	feed, stats, err := ParseProducts(context.Background(), strings.NewReader(input), ProductsOptions{})
	if err != nil {
		t.Fatalf("parse HON products: %v", err)
	}
	if stats.ProductsRead != 2 || stats.ProductsEmitted != 1 || stats.ProductsSkipped != 1 {
		t.Fatalf("unexpected stats: %#v", stats)
	}
	if len(feed.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(feed.Items))
	}

	item := feed.Items[0]
	if item.Code != "DY10010001-010042" || item.Stock != "74" || item.PriceVAT != "10756.90" {
		t.Fatalf("unexpected item fields: %#v", item)
	}
	if item.Supplier != "HON" {
		t.Fatalf("unexpected supplier: %q", item.Supplier)
	}
	if item.Manufacturer != "HON" {
		t.Fatalf("unexpected manufacturer fallback: %q", item.Manufacturer)
	}
	if item.Name != "MERENS BP - černá BI 201, kanc. židle bez podhlavníku" {
		t.Fatalf("unexpected name: %q", item.Name)
	}
	if len(item.Images) != 1 || item.Images[0].URL != "https://example.test/merens.jpg" {
		t.Fatalf("unexpected images: %#v", item.Images)
	}
	if item.DefaultCategory == nil || item.DefaultCategory.ID != "881" {
		t.Fatalf("unexpected category: %#v", item.DefaultCategory)
	}
	expectedParameters := []shoptet.Parameter{
		{Name: "Šířka", Value: "20"},
		{Name: "Výška", Value: "85.5"},
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
	input := `<SHOP>
  <SHOPITEM><ID>1</ID><PRODUCT>Produkt 1</PRODUCT></SHOPITEM>
  <SHOPITEM><ID>2</ID><PRODUCT>Produkt 2</PRODUCT></SHOPITEM>
</SHOP>`

	feed, stats, err := ParseProducts(context.Background(), strings.NewReader(input), ProductsOptions{MaxProducts: 1})
	if err != nil {
		t.Fatalf("parse limited HON products: %v", err)
	}
	if stats.ProductsRead != 2 || stats.ProductsEmitted != 1 {
		t.Fatalf("unexpected stats: %#v", stats)
	}
	if len(feed.Items) != 1 || feed.Items[0].Code != "1" {
		t.Fatalf("unexpected feed: %#v", feed)
	}
}

func TestTransformManufacturerMapsKnownHONCategoryMarkers(t *testing.T) {
	tests := []struct {
		name         string
		mainCategory string
		expected     string
	}{
		{
			name:         "officepro",
			mainCategory: "Židle kancelářské OfficePro",
			expected:     "Office Pro",
		},
		{
			name:         "loffler",
			mainCategory: "Židle LÖFFLER",
			expected:     "LÖFFLER",
		},
		{
			name:         "honplus fallback",
			mainCategory: "Židle kancelářské HonPlus",
			expected:     "HON",
		},
		{
			name:         "generic fallback",
			mainCategory: "Doplňky",
			expected:     "HON",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := transformManufacturer(test.mainCategory); got != test.expected {
				t.Fatalf("unexpected manufacturer: %q", got)
			}
		})
	}
}

func TestParseProductsMapsOfficeProManufacturerFromMainCategory(t *testing.T) {
	input := `<SHOP>
  <SHOPITEM>
    <ID>812692</ID>
    <MAIN_CATEGORY>Židle kancelářské OfficePro</MAIN_CATEGORY>
    <PRODUCT>CALYPSO</PRODUCT>
    <PART_NUMBER>DY40010001-073028</PART_NUMBER>
    <DESCRIPTION>antracit 1211,kanc.židle</DESCRIPTION>
  </SHOPITEM>
  <SHOPITEM>
    <ID>812693</ID>
    <MAIN_CATEGORY>Židle kancelářské OfficePro</MAIN_CATEGORY>
    <PRODUCT>CALYPSO</PRODUCT>
    <PART_NUMBER>DY40010002-073027</PART_NUMBER>
    <DESCRIPTION>sv.šedá12A11,kanc.židle</DESCRIPTION>
  </SHOPITEM>
</SHOP>`

	feed, stats, err := ParseProducts(context.Background(), strings.NewReader(input), ProductsOptions{})
	if err != nil {
		t.Fatalf("parse HON products: %v", err)
	}
	if stats.ProductsRead != 2 || stats.ProductsEmitted != 1 || stats.ItemsWithVariants != 1 {
		t.Fatalf("unexpected stats: %#v", stats)
	}
	if len(feed.Items) != 1 {
		t.Fatalf("expected grouped Office Pro item, got %#v", feed.Items)
	}
	item := feed.Items[0]
	if item.Supplier != "HON" {
		t.Fatalf("unexpected supplier: %q", item.Supplier)
	}
	if item.Manufacturer != "Office Pro" {
		t.Fatalf("unexpected manufacturer: %q", item.Manufacturer)
	}
}

func TestTransformCategoryMapsMeetingChairsToExistingConferenceCategory(t *testing.T) {
	categories, defaultCategory := transformCategory("Židle jednací OfficePro")
	expected := shoptet.Category{ID: "1146", Path: "ŽIDLE > KONFERENČNÍ ŽIDLE"}
	if len(categories) != 1 || categories[0] != expected {
		t.Fatalf("unexpected categories: %#v", categories)
	}
	if defaultCategory == nil || *defaultCategory != expected {
		t.Fatalf("unexpected default category: %#v", defaultCategory)
	}
}

func TestParseProductsGroupsHONVariantsByProductAndDescription(t *testing.T) {
	input := `<SHOP>
  <SHOPITEM>
    <ID>812728</ID>
    <MAIN_CATEGORY>Ostatní židle</MAIN_CATEGORY>
    <PRODUCT>DORA</PRODUCT>
    <PRICE_VAT>1379.40</PRICE_VAT>
    <DOSTUPNOST>Na dotaz</DOSTUPNOST>
    <STOCK>0.0</STOCK>
    <PART_NUMBER>DY60010001-041039</PART_NUMBER>
    <DESCRIPTION>dřevěná židle,TŘEŠEŇ/CHROM</DESCRIPTION>
    <IMGURL><IMGURL>https://example.test/dora-tresen-chrom.jpg</IMGURL></IMGURL>
    <PARAM><PARAM_NAME>Šířka</PARAM_NAME><VAL>870.0</VAL></PARAM>
  </SHOPITEM>
  <SHOPITEM>
    <ID>812729</ID>
    <MAIN_CATEGORY>Ostatní židle</MAIN_CATEGORY>
    <PRODUCT>DORA</PRODUCT>
    <PRICE_VAT>1379.40</PRICE_VAT>
    <DOSTUPNOST>Na dotaz</DOSTUPNOST>
    <STOCK>0.0</STOCK>
    <PART_NUMBER>DY60010002-041038</PART_NUMBER>
    <DESCRIPTION>dřevěná židle,BUK/CHROM</DESCRIPTION>
    <IMGURL><IMGURL>https://example.test/dora-buk-chrom.jpg</IMGURL></IMGURL>
    <PARAM><PARAM_NAME>Šířka</PARAM_NAME><VAL>870.0</VAL></PARAM>
  </SHOPITEM>
  <SHOPITEM>
    <ID>812730</ID>
    <MAIN_CATEGORY>Ostatní židle</MAIN_CATEGORY>
    <PRODUCT>DORA</PRODUCT>
    <PRICE_VAT>1379.40</PRICE_VAT>
    <DOSTUPNOST>Na dotaz</DOSTUPNOST>
    <STOCK>0.0</STOCK>
    <PART_NUMBER>DY60010003-040039</PART_NUMBER>
    <DESCRIPTION>dřevěná židle,TŘEŠEŇ/HLINÍK</DESCRIPTION>
    <IMGURL><IMGURL>https://example.test/dora-tresen-hlinik.jpg</IMGURL></IMGURL>
    <PARAM><PARAM_NAME>Šířka</PARAM_NAME><VAL>870.0</VAL></PARAM>
  </SHOPITEM>
  <SHOPITEM>
    <ID>812731</ID>
    <MAIN_CATEGORY>Ostatní židle</MAIN_CATEGORY>
    <PRODUCT>DORA</PRODUCT>
    <PRICE_VAT>1379.40</PRICE_VAT>
    <DOSTUPNOST>Na dotaz</DOSTUPNOST>
    <STOCK>0.0</STOCK>
    <PART_NUMBER>DY60010004-040038</PART_NUMBER>
    <DESCRIPTION>dřevěná židle,BUK/HLINÍK</DESCRIPTION>
    <IMGURL><IMGURL>https://example.test/dora-buk-hlinik.jpg</IMGURL></IMGURL>
    <PARAM><PARAM_NAME>Šířka</PARAM_NAME><VAL>870.0</VAL></PARAM>
  </SHOPITEM>
</SHOP>`

	feed, stats, err := ParseProducts(context.Background(), strings.NewReader(input), ProductsOptions{})
	if err != nil {
		t.Fatalf("parse HON products: %v", err)
	}
	if stats.ProductsRead != 4 || stats.ProductsEmitted != 1 || stats.ItemsWithVariants != 1 || stats.VariantsEmitted != 4 {
		t.Fatalf("unexpected stats: %#v", stats)
	}
	if len(feed.Items) != 1 {
		t.Fatalf("expected grouped HON product, got %#v", feed.Items)
	}

	item := feed.Items[0]
	if item.Code != "HON-DORA" || item.Name != "DORA" || item.Description != "dřevěná židle" {
		t.Fatalf("unexpected parent item: %#v", item)
	}
	if item.Manufacturer != "HON" {
		t.Fatalf("unexpected parent manufacturer fallback: %q", item.Manufacturer)
	}
	if len(item.Images) != 4 {
		t.Fatalf("expected merged images, got %#v", item.Images)
	}
	if len(item.InformationParameters) != 1 || item.InformationParameters[0] != (shoptet.Parameter{Name: "Šířka", Value: "870"}) {
		t.Fatalf("unexpected information parameters: %#v", item.InformationParameters)
	}
	if len(item.Variants) != 4 {
		t.Fatalf("expected 4 variants, got %#v", item.Variants)
	}
	expectedValues := []string{"TŘEŠEŇ/CHROM", "BUK/CHROM", "TŘEŠEŇ/HLINÍK", "BUK/HLINÍK"}
	for index, expectedValue := range expectedValues {
		variant := item.Variants[index]
		if len(variant.Parameters) != 1 || variant.Parameters[0] != (shoptet.Parameter{Name: "Provedení", Value: expectedValue}) {
			t.Fatalf("unexpected variant parameter %d: %#v", index, variant.Parameters)
		}
		if variant.ImageRef == "" {
			t.Fatalf("expected image ref on variant %d", index)
		}
	}
}

func TestParseProductsKeepsAmbiguousRepeatedProductsStandalone(t *testing.T) {
	input := `<SHOP>
  <SHOPITEM>
    <ID>1</ID>
    <MAIN_CATEGORY>Podložky pod židle</MAIN_CATEGORY>
    <PRODUCT>Podložka</PRODUCT>
    <PART_NUMBER>DYC0010001-000000</PART_NUMBER>
    <DESCRIPTION>Podložka pod židle univerzální OFFICE /120x98 cm/</DESCRIPTION>
  </SHOPITEM>
  <SHOPITEM>
    <ID>2</ID>
    <MAIN_CATEGORY>Podložky pod židle</MAIN_CATEGORY>
    <PRODUCT>Podložka</PRODUCT>
    <PART_NUMBER>DYC0020001-000000</PART_NUMBER>
    <DESCRIPTION>Podložka pod židle s hroty na koberec OFFICE H /120x100 cm/</DESCRIPTION>
  </SHOPITEM>
</SHOP>`

	feed, stats, err := ParseProducts(context.Background(), strings.NewReader(input), ProductsOptions{})
	if err != nil {
		t.Fatalf("parse HON products: %v", err)
	}
	if stats.ProductsRead != 2 || stats.ProductsEmitted != 2 || stats.ItemsWithVariants != 0 || stats.VariantsEmitted != 0 {
		t.Fatalf("unexpected stats: %#v", stats)
	}
	if len(feed.Items) != 2 {
		t.Fatalf("expected standalone products, got %#v", feed.Items)
	}
	for _, item := range feed.Items {
		if len(item.Variants) != 0 {
			t.Fatalf("expected simple product, got variants: %#v", item)
		}
	}
}

func TestParseProductsKeepsDuplicateVariantValuesStandalone(t *testing.T) {
	input := `<SHOP>
  <SHOPITEM>
    <ID>1</ID>
    <MAIN_CATEGORY>Židle</MAIN_CATEGORY>
    <PRODUCT>DUP</PRODUCT>
    <PART_NUMBER>DUP-1</PART_NUMBER>
    <DESCRIPTION>židle,černá</DESCRIPTION>
  </SHOPITEM>
  <SHOPITEM>
    <ID>2</ID>
    <MAIN_CATEGORY>Židle</MAIN_CATEGORY>
    <PRODUCT>DUP</PRODUCT>
    <PART_NUMBER>DUP-2</PART_NUMBER>
    <DESCRIPTION>židle,černá</DESCRIPTION>
  </SHOPITEM>
</SHOP>`

	feed, stats, err := ParseProducts(context.Background(), strings.NewReader(input), ProductsOptions{})
	if err != nil {
		t.Fatalf("parse HON products: %v", err)
	}
	if stats.ProductsEmitted != 2 || stats.ItemsWithVariants != 0 {
		t.Fatalf("unexpected stats: %#v", stats)
	}
	if len(feed.Items) != 2 {
		t.Fatalf("expected duplicate values to stay standalone, got %#v", feed.Items)
	}
}

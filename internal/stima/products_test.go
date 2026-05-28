package stima

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/fanda/arpari-xml-feed-parser-railway/internal/shoptet"
)

func TestParseProductsSimpleProduct(t *testing.T) {
	input := `<SHOP>
  <SHOPITEM>
    <NAME>Stůl SIMPLE</NAME>
    <SHORT_DESCRIPTION>Krátký popis</SHORT_DESCRIPTION>
    <DESCRIPTION>Dlouhý popis</DESCRIPTION>
    <CODE>ART-SIMPLE</CODE>
    <EAN>8590000000001</EAN>
    <PRICE_VAT>1234.00</PRICE_VAT>
    <CATEGORIES>
      <CATEGORY>Katalog &gt; Židle</CATEGORY>
      <CATEGORY>Katalog &gt; Židle &gt; Dřevěné židle</CATEGORY>
      <CATEGORY>Katalog &gt; Katalog 2026</CATEGORY>
    </CATEGORIES>
    <IMAGES>
      <IMAGE>https://www.stima.cz/userfiles/xml/pictures/simple.jpg</IMAGE>
      <IMAGE>https://www.stima.cz/userfiles/xml/pictures/simple.jpg</IMAGE>
      <IMAGE>  </IMAGE>
    </IMAGES>
    <INFORMATION_PARAMETERS>
      <INFORMATION_PARAMETER><NAME>Materiál</NAME><VALUE>Buk</VALUE></INFORMATION_PARAMETER>
      <INFORMATION_PARAMETER><NAME>Nosnost</NAME><VALUE>120 kg</VALUE></INFORMATION_PARAMETER>
    </INFORMATION_PARAMETERS>
    <STOCK>
      <WAREHOUSES>
        <WAREHOUSE>
          <NAME>HLAVNÍ SKLAD</NAME>
          <VALUE>5.000</VALUE>
        </WAREHOUSE>
      </WAREHOUSES>
    </STOCK>
  </SHOPITEM>
</SHOP>`

	feed, stats, err := ParseProducts(context.Background(), strings.NewReader(input), ProductsOptions{})
	if err != nil {
		t.Fatalf("parse products: %v", err)
	}
	if stats.ProductsRead != 1 || stats.ProductsEmitted != 1 || stats.ProductsSkipped != 0 {
		t.Fatalf("unexpected stats: %#v", stats)
	}
	if len(feed.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(feed.Items))
	}

	item := feed.Items[0]
	if item.Code != "ART-SIMPLE" || item.Name != "Stůl SIMPLE" || item.EAN != "8590000000001" || item.PriceVAT != "1234.00" {
		t.Fatalf("unexpected simple item: %#v", item)
	}
	if item.ShortDescription != "Krátký popis" || item.Description != "Dlouhý popis" {
		t.Fatalf("unexpected descriptions: %#v", item)
	}
	if len(item.Warehouses) != 1 || item.Warehouses[0] != (shoptet.Warehouse{Name: "HLAVNÍ SKLAD", Value: "5.000"}) {
		t.Fatalf("unexpected warehouses: %#v", item.Warehouses)
	}
	if len(item.Categories) != 2 {
		t.Fatalf("expected mapped categories, got %#v", item.Categories)
	}
	if item.Categories[1] != (shoptet.Category{ID: "905", Path: "ŽIDLE > DŘEVĚNÉ ŽIDLE"}) {
		t.Fatalf("unexpected mapped category: %#v", item.Categories[1])
	}
	if item.DefaultCategory == nil || *item.DefaultCategory != (shoptet.Category{ID: "905", Path: "ŽIDLE > DŘEVĚNÉ ŽIDLE"}) {
		t.Fatalf("unexpected default category: %#v", item.DefaultCategory)
	}
	if len(item.Images) != 1 || item.Images[0] != (shoptet.Image{URL: "https://www.stima.cz/userfiles/xml/pictures/simple.jpg"}) {
		t.Fatalf("unexpected images: %#v", item.Images)
	}
	expectedParameters := []shoptet.Parameter{
		{Name: "Materiál", Value: "Buk"},
		{Name: "Nosnost", Value: "120 kg"},
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

func TestParseProductsVariantProductDerivesParentCodeAndMapsParameters(t *testing.T) {
	input := `<SHOP>
  <SHOPITEM>
    <NAME>Židle NANCY KR58</NAME>
    <SHORT_DESCRIPTION>Krátký popis židle</SHORT_DESCRIPTION>
    <DESCRIPTION>Dlouhý popis židle</DESCRIPTION>
    <CATEGORIES>
      <CATEGORY>Katalog &gt; Restaurační židle</CATEGORY>
      <CATEGORY>Katalog &gt; Restaurační židle &gt; Stále skladem</CATEGORY>
      <CATEGORY>Katalog &gt; Židle</CATEGORY>
    </CATEGORIES>
    <IMAGES>
      <IMAGE>https://www.stima.cz/userfiles/xml/pictures/nancy.jpg</IMAGE>
    </IMAGES>
    <INFORMATION_PARAMETERS>
      <INFORMATION_PARAMETER><NAME>Výška</NAME><VALUE>88 cm</VALUE></INFORMATION_PARAMETER>
    </INFORMATION_PARAMETERS>
    <VARIANTS>
      <VARIANT>
        <CODE>ART13627-k002-l244</CODE>
        <EAN>8590000000002</EAN>
        <PRICE_VAT>2850.00</PRICE_VAT>
        <STOCK>4</STOCK>
        <PARAMETERS>
          <PARAMETER><NAME>KOSTRA</NAME><VALUE>olše</VALUE></PARAMETER>
          <PARAMETER><NAME>Sedák</NAME><VALUE>raven 15 šedá</VALUE></PARAMETER>
          <PARAMETER><NAME>Barva</NAME><VALUE>ignored</VALUE></PARAMETER>
        </PARAMETERS>
      </VARIANT>
    </VARIANTS>
  </SHOPITEM>
</SHOP>`

	feed, stats, err := ParseProducts(context.Background(), strings.NewReader(input), ProductsOptions{})
	if err != nil {
		t.Fatalf("parse products: %v", err)
	}
	if stats.ProductsRead != 1 || stats.ProductsEmitted != 1 || stats.ItemsWithVariants != 1 || stats.VariantsEmitted != 1 {
		t.Fatalf("unexpected stats: %#v", stats)
	}

	item := feed.Items[0]
	if item.Code != "ART13627" {
		t.Fatalf("expected derived parent code ART13627, got %q", item.Code)
	}
	if item.ShortDescription != "Krátký popis židle" || item.Description != "Dlouhý popis židle" {
		t.Fatalf("unexpected descriptions: %#v", item)
	}
	if item.DefaultCategory == nil || *item.DefaultCategory != (shoptet.Category{ID: "1128", Path: "ŽIDLE > RESTAURAČNÍ ŽIDLE"}) {
		t.Fatalf("unexpected default category: %#v", item.DefaultCategory)
	}
	if len(item.Images) != 1 || item.Images[0].URL != "https://www.stima.cz/userfiles/xml/pictures/nancy.jpg" {
		t.Fatalf("unexpected images: %#v", item.Images)
	}
	if len(item.InformationParameters) != 1 || item.InformationParameters[0] != (shoptet.Parameter{Name: "Výška", Value: "88 cm"}) {
		t.Fatalf("unexpected information parameters: %#v", item.InformationParameters)
	}
	if len(item.Variants) != 1 {
		t.Fatalf("expected 1 variant, got %d", len(item.Variants))
	}

	variant := item.Variants[0]
	if variant.Code != "ART13627-k002-l244" || variant.EAN != "8590000000002" || variant.PriceVAT != "2850.00" || variant.Stock != "4" {
		t.Fatalf("unexpected variant: %#v", variant)
	}
	if len(variant.Parameters) != 2 {
		t.Fatalf("expected only allowed parameters, got %#v", variant.Parameters)
	}
	if variant.Parameters[0] != (shoptet.Parameter{Name: "KOSTRA", Value: "olše"}) {
		t.Fatalf("unexpected first parameter: %#v", variant.Parameters[0])
	}
	if variant.Parameters[1] != (shoptet.Parameter{Name: "Sedák", Value: "raven 15 šedá"}) {
		t.Fatalf("unexpected second parameter: %#v", variant.Parameters[1])
	}
}

func TestParseProductsTrimsProductsAboveShoptetVariantLimit(t *testing.T) {
	var input bytes.Buffer
	input.WriteString(`<SHOP><SHOPITEM><NAME>Large Variant Product</NAME><CATEGORIES><CATEGORY>Katalog &gt; Židle</CATEGORY></CATEGORIES><VARIANTS>`)
	for i := 0; i < shoptet.DefaultMaxVariantsPerItem+2; i++ {
		fmt.Fprintf(&input, `<VARIANT><CODE>ART-LARGE-%03d</CODE><PRICE_VAT>100.00</PRICE_VAT></VARIANT>`, i)
	}
	input.WriteString(`</VARIANTS></SHOPITEM></SHOP>`)

	feed, stats, err := ParseProducts(context.Background(), &input, ProductsOptions{})
	if err != nil {
		t.Fatalf("parse products: %v", err)
	}
	if len(feed.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(feed.Items))
	}
	if len(feed.Items[0].Variants) != shoptet.DefaultMaxVariantsPerItem {
		t.Fatalf("expected %d variants, got %d", shoptet.DefaultMaxVariantsPerItem, len(feed.Items[0].Variants))
	}
	if stats.ProductsTrimmed != 1 || stats.VariantsTrimmed != 2 {
		t.Fatalf("unexpected trim stats: %#v", stats)
	}
	if feed.Items[0].Variants[shoptet.DefaultMaxVariantsPerItem-1].Code != "ART-LARGE-511" {
		t.Fatalf("expected source order to be preserved, got last code %q", feed.Items[0].Variants[shoptet.DefaultMaxVariantsPerItem-1].Code)
	}
}

func TestParseProductsSkipsVariantWithoutCode(t *testing.T) {
	input := `<SHOP>
  <SHOPITEM>
    <NAME>Židle Variant Product</NAME>
    <VARIANTS>
      <VARIANT><PRICE_VAT>100.00</PRICE_VAT></VARIANT>
      <VARIANT><CODE>ART-SAFE-k001</CODE><PRICE_VAT>120.00</PRICE_VAT></VARIANT>
    </VARIANTS>
  </SHOPITEM>
</SHOP>`

	feed, stats, err := ParseProducts(context.Background(), strings.NewReader(input), ProductsOptions{})
	if err != nil {
		t.Fatalf("parse products: %v", err)
	}
	if stats.VariantsSkipped != 1 || stats.VariantsEmitted != 1 || stats.ProductsSkipped != 0 {
		t.Fatalf("unexpected stats: %#v", stats)
	}
	if len(feed.Items) != 1 || len(feed.Items[0].Variants) != 1 {
		t.Fatalf("unexpected feed: %#v", feed)
	}
	if feed.Items[0].Code != "ART" {
		t.Fatalf("expected derived parent code ART, got %q", feed.Items[0].Code)
	}
}

func TestParseProductsSkipsSimpleProductWithoutCode(t *testing.T) {
	input := `<SHOP>
  <SHOPITEM>
    <NAME>Missing code</NAME>
    <PRICE_VAT>100.00</PRICE_VAT>
  </SHOPITEM>
</SHOP>`

	feed, stats, err := ParseProducts(context.Background(), strings.NewReader(input), ProductsOptions{})
	if err != nil {
		t.Fatalf("parse products: %v", err)
	}
	if len(feed.Items) != 0 {
		t.Fatalf("expected skipped product, got %#v", feed.Items)
	}
	if stats.ProductsRead != 1 || stats.ProductsSkipped != 1 {
		t.Fatalf("unexpected stats: %#v", stats)
	}
}

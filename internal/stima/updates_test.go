package stima

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/fanda/arpari-xml-feed-parser-railway/internal/shoptet"
)

func TestParseStockSimpleAndVariantItems(t *testing.T) {
	input := `<SHOP>
  <SHOPITEM>
    <CODE>ART-SIMPLE</CODE>
    <STOCK>
      <WAREHOUSES>
        <WAREHOUSE><NAME>HLAVNÍ SKLAD</NAME><VALUE>7.00</VALUE></WAREHOUSE>
      </WAREHOUSES>
    </STOCK>
  </SHOPITEM>
  <SHOPITEM>
    <VARIANTS>
      <VARIANT>
        <CODE>ART13627-k002-l244</CODE>
        <STOCK>
          <WAREHOUSES>
            <WAREHOUSE><NAME>HLAVNÍ SKLAD</NAME><VALUE>4.00</VALUE></WAREHOUSE>
          </WAREHOUSES>
        </STOCK>
        <PARAMETERS>
          <PARAMETER><NAME>KOSTRA</NAME><VALUE>olše</VALUE></PARAMETER>
          <PARAMETER><NAME>Sedák</NAME><VALUE>raven 15 šedá</VALUE></PARAMETER>
          <PARAMETER><NAME>Specifikace</NAME><VALUE>výška 70</VALUE></PARAMETER>
        </PARAMETERS>
      </VARIANT>
    </VARIANTS>
  </SHOPITEM>
</SHOP>`

	feed, stats, err := ParseStock(context.Background(), strings.NewReader(input), UpdateOptions{})
	if err != nil {
		t.Fatalf("parse stock: %v", err)
	}
	if stats.ProductsRead != 2 || stats.ProductsEmitted != 2 || stats.ItemsWithVariants != 1 || stats.VariantsEmitted != 1 {
		t.Fatalf("unexpected stats: %#v", stats)
	}
	if len(feed.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(feed.Items))
	}

	simple := feed.Items[0]
	if simple.Code != "ART-SIMPLE" || len(simple.Warehouses) != 1 || simple.Warehouses[0].Value != "7.00" {
		t.Fatalf("unexpected simple stock item: %#v", simple)
	}

	variantParent := feed.Items[1]
	if variantParent.Code != "ART13627" {
		t.Fatalf("expected derived parent code ART13627, got %q", variantParent.Code)
	}
	if len(variantParent.Variants) != 1 {
		t.Fatalf("expected 1 stock variant, got %d", len(variantParent.Variants))
	}
	variant := variantParent.Variants[0]
	if variant.Code != "ART13627-k002-l244" || variant.PriceVAT != "" || len(variant.Warehouses) != 1 || variant.Warehouses[0].Value != "4.00" {
		t.Fatalf("unexpected stock variant: %#v", variant)
	}
	if len(variant.Parameters) != 3 {
		t.Fatalf("expected parameters on stock variant, got %#v", variant.Parameters)
	}
	if variant.Parameters[2] != (shoptet.Parameter{Name: "Specifikace", Value: "výška 70"}) {
		t.Fatalf("expected specification parameter on stock variant, got %#v", variant.Parameters)
	}
}

func TestParseStockPriceKeepsPriceAndStock(t *testing.T) {
	input := `<SHOP>
  <SHOPITEM>
    <CODE>ART-SIMPLE</CODE>
    <PRICE_VAT>12900.00</PRICE_VAT>
    <STOCK>
      <WAREHOUSES>
        <WAREHOUSE><NAME>HLAVNÍ SKLAD</NAME><VALUE>0.000</VALUE></WAREHOUSE>
      </WAREHOUSES>
    </STOCK>
  </SHOPITEM>
  <SHOPITEM>
    <VARIANTS>
      <VARIANT>
        <CODE>ART13627-k002-l244</CODE>
        <PRICE_VAT>2850.00</PRICE_VAT>
        <STOCK>
          <WAREHOUSES>
            <WAREHOUSE><NAME>HLAVNÍ SKLAD</NAME><VALUE>4.000</VALUE></WAREHOUSE>
          </WAREHOUSES>
        </STOCK>
      </VARIANT>
    </VARIANTS>
  </SHOPITEM>
</SHOP>`

	feed, stats, err := ParseStockPrice(context.Background(), strings.NewReader(input), UpdateOptions{})
	if err != nil {
		t.Fatalf("parse stock-price: %v", err)
	}
	if stats.ProductsRead != 2 || stats.ProductsEmitted != 2 || stats.VariantsEmitted != 1 {
		t.Fatalf("unexpected stats: %#v", stats)
	}
	if feed.Items[0].Code != "ART-SIMPLE" || feed.Items[0].PriceVAT != "12900.00" {
		t.Fatalf("unexpected simple price item: %#v", feed.Items[0])
	}
	variant := feed.Items[1].Variants[0]
	if variant.Code != "ART13627-k002-l244" || variant.PriceVAT != "2850.00" || len(variant.Warehouses) != 1 {
		t.Fatalf("unexpected stock-price variant: %#v", variant)
	}
}

func TestParseStockPriceFiltersWhitelistedFabricPrefixes(t *testing.T) {
	input := `<SHOP>
  <SHOPITEM>
    <VARIANTS>
      <VARIANT>
        <CODE>ART13445-k001-l177</CODE>
        <PRICE_VAT>100</PRICE_VAT>
        <STOCK>77</STOCK>
        <PARAMETERS>
          <PARAMETER><NAME>Sedák</NAME><VALUE>lux 15 bordo</VALUE></PARAMETER>
        </PARAMETERS>
      </VARIANT>
      <VARIANT>
        <CODE>ART13445-k001-l200</CODE>
        <PRICE_VAT>120</PRICE_VAT>
        <STOCK>2</STOCK>
        <PARAMETERS>
          <PARAMETER><NAME>Sedák</NAME><VALUE>boss 20 šedá</VALUE></PARAMETER>
        </PARAMETERS>
      </VARIANT>
    </VARIANTS>
  </SHOPITEM>
</SHOP>`

	feed, stats, err := ParseStockPrice(context.Background(), strings.NewReader(input), UpdateOptions{
		VariantWhitelist: FabricWhitelist{
			"ART13445-K001-L177": {"lux"},
		},
	})
	if err != nil {
		t.Fatalf("parse stock-price: %v", err)
	}
	if len(feed.Items) != 1 || len(feed.Items[0].Variants) != 1 {
		t.Fatalf("expected one whitelisted variant, got %#v", feed.Items)
	}
	variant := feed.Items[0].Variants[0]
	if variant.Code != "ART13445-k001-l177" || variant.PriceVAT != "100" || variant.Stock != "77" {
		t.Fatalf("unexpected variant: %#v", variant)
	}
	if stats.VariantsSkipped != 1 || stats.VariantsEmitted != 1 {
		t.Fatalf("unexpected stats: %#v", stats)
	}
}

func TestParseStockSkipsItemsWithoutCodeOrPayload(t *testing.T) {
	input := `<SHOP>
  <SHOPITEM>
    <STOCK>1</STOCK>
  </SHOPITEM>
  <SHOPITEM>
    <CODE>ART-NO-PAYLOAD</CODE>
  </SHOPITEM>
  <SHOPITEM>
    <CODE>DOPRAVA</CODE>
    <STOCK>1</STOCK>
  </SHOPITEM>
  <SHOPITEM>
    <VARIANTS>
      <VARIANT><STOCK>1</STOCK></VARIANT>
      <VARIANT><CODE>ART-OK-k001</CODE></VARIANT>
      <VARIANT><CODE>ART-OK-k002</CODE><STOCK>2</STOCK></VARIANT>
    </VARIANTS>
  </SHOPITEM>
</SHOP>`

	feed, stats, err := ParseStock(context.Background(), strings.NewReader(input), UpdateOptions{})
	if err != nil {
		t.Fatalf("parse stock: %v", err)
	}
	if stats.ProductsRead != 4 || stats.ProductsSkipped != 3 || stats.VariantsSkipped != 2 || stats.VariantsEmitted != 1 {
		t.Fatalf("unexpected stats: %#v", stats)
	}
	if len(feed.Items) != 1 || feed.Items[0].Code != "ART" || len(feed.Items[0].Variants) != 1 {
		t.Fatalf("unexpected feed: %#v", feed)
	}
}

func TestParseStockTrimsProductsAboveShoptetVariantLimit(t *testing.T) {
	var input bytes.Buffer
	input.WriteString(`<SHOP><SHOPITEM><VARIANTS>`)
	for i := 0; i < shoptet.DefaultMaxVariantsPerItem+3; i++ {
		fmt.Fprintf(&input, `<VARIANT><CODE>ART-LARGE-%03d</CODE><STOCK>1</STOCK></VARIANT>`, i)
	}
	input.WriteString(`</VARIANTS></SHOPITEM></SHOP>`)

	feed, stats, err := ParseStock(context.Background(), &input, UpdateOptions{})
	if err != nil {
		t.Fatalf("parse stock: %v", err)
	}
	if len(feed.Items) != 1 || len(feed.Items[0].Variants) != shoptet.DefaultMaxVariantsPerItem {
		t.Fatalf("unexpected feed size: %#v", feed)
	}
	if stats.ProductsTrimmed != 1 || stats.VariantsTrimmed != 3 {
		t.Fatalf("unexpected trim stats: %#v", stats)
	}
}

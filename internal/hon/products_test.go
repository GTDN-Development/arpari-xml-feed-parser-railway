package hon

import (
	"context"
	"strings"
	"testing"
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
	if item.Name != "MERENS BP - černá BI 201, kanc. židle bez podhlavníku" {
		t.Fatalf("unexpected name: %q", item.Name)
	}
	if len(item.Images) != 1 || item.Images[0].URL != "https://example.test/merens.jpg" {
		t.Fatalf("unexpected images: %#v", item.Images)
	}
	if item.DefaultCategory == nil || item.DefaultCategory.ID != "881" {
		t.Fatalf("unexpected category: %#v", item.DefaultCategory)
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
	if stats.ProductsRead != 1 || stats.ProductsEmitted != 1 {
		t.Fatalf("unexpected stats: %#v", stats)
	}
	if len(feed.Items) != 1 || feed.Items[0].Code != "1" {
		t.Fatalf("unexpected feed: %#v", feed)
	}
}

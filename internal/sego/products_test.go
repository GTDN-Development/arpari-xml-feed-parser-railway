package sego

import (
	"context"
	"strings"
	"testing"
)

func TestParseProductsMapsZboziItems(t *testing.T) {
	input := `<?xml version="1.0" encoding="utf-8"?>
<SHOP xmlns="http://www.zbozi.cz/ns/offer/1.0">
  <SHOPITEM>
    <ITEM_ID>0745314610292</ITEM_ID>
    <PRODUCTNAME>AIR plus | Černá</PRODUCTNAME>
    <DESCRIPTION><![CDATA[Popis&nbsp;židle]]></DESCRIPTION>
    <BRAND>Pixel</BRAND>
    <EAN>0745314610292</EAN>
    <IMGURL>https://segocz.cz/main.jpg</IMGURL>
    <IMGURL_ALTERNATIVE>https://segocz.cz/alt.jpg</IMGURL_ALTERNATIVE>
    <PRICE_VAT>10951.00</PRICE_VAT>
    <DELIVERY_DATE>0</DELIVERY_DATE>
  </SHOPITEM>
  <SHOPITEM>
    <PRODUCTNAME>Bez kódu</PRODUCTNAME>
  </SHOPITEM>
</SHOP>`

	feed, stats, err := ParseProducts(context.Background(), strings.NewReader(input), ProductsOptions{})
	if err != nil {
		t.Fatalf("parse SEGO products: %v", err)
	}
	if stats.ProductsRead != 2 || stats.ProductsEmitted != 1 || stats.ProductsSkipped != 1 {
		t.Fatalf("unexpected stats: %#v", stats)
	}
	if len(feed.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(feed.Items))
	}

	item := feed.Items[0]
	if item.Code != "0745314610292" || item.PriceVAT != "10951.00" || item.Availability != "Skladem" {
		t.Fatalf("unexpected item fields: %#v", item)
	}
	if item.Description != "Popis židle" {
		t.Fatalf("unexpected description: %q", item.Description)
	}
	if len(item.Images) != 2 {
		t.Fatalf("expected 2 images, got %#v", item.Images)
	}
	if item.DefaultCategory == nil || item.DefaultCategory.ID != "881" {
		t.Fatalf("unexpected category: %#v", item.DefaultCategory)
	}
}

func TestParseProductsLimitsTestOutput(t *testing.T) {
	input := `<SHOP>
  <SHOPITEM><ITEM_ID>1</ITEM_ID><PRODUCTNAME>Produkt 1</PRODUCTNAME></SHOPITEM>
  <SHOPITEM><ITEM_ID>2</ITEM_ID><PRODUCTNAME>Produkt 2</PRODUCTNAME></SHOPITEM>
</SHOP>`

	feed, stats, err := ParseProducts(context.Background(), strings.NewReader(input), ProductsOptions{MaxProducts: 1})
	if err != nil {
		t.Fatalf("parse limited SEGO products: %v", err)
	}
	if stats.ProductsRead != 1 || stats.ProductsEmitted != 1 {
		t.Fatalf("unexpected stats: %#v", stats)
	}
	if len(feed.Items) != 1 || feed.Items[0].Code != "1" {
		t.Fatalf("unexpected feed: %#v", feed)
	}
}

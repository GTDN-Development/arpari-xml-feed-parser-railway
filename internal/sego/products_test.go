package sego

import (
	"context"
	"strings"
	"testing"

	"github.com/fanda/arpari-xml-feed-parser-railway/internal/shoptet"
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
    <PRICE_VAT>10951.49</PRICE_VAT>
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
	if item.Code != "0745314610292" || item.PriceVAT != "10951" || item.VAT != "21" || item.Currency != "CZK" || item.Availability != "Skladem" {
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

func TestParseProductsGroupsFlatColorVariants(t *testing.T) {
	input := `<?xml version="1.0" encoding="utf-8"?>
<SHOP xmlns="http://www.zbozi.cz/ns/offer/1.0">
  <SHOPITEM>
    <ITEM_ID>0745314610063</ITEM_ID>
    <PRODUCTNAME>Junior | Červená</PRODUCTNAME>
    <DESCRIPTION><![CDATA[Pevná dětská židle]]></DESCRIPTION>
    <URL>https://segocz.cz/produkty/detail/junior?color=cervena</URL>
    <EAN>0745314610063</EAN>
    <IMGURL>https://segocz.cz/src/Frontend/Files/Catalog/VariantImages/7/previewImg-7-0_cs.jpg</IMGURL>
    <IMGURL_ALTERNATIVE>https://segocz.cz/red.jpg</IMGURL_ALTERNATIVE>
    <PRICE_VAT>4623.00</PRICE_VAT>
    <DELIVERY_DATE>0</DELIVERY_DATE>
    <PARAM><PARAM_NAME><![CDATA[Barva]]></PARAM_NAME><VAL><![CDATA[Červená]]></VAL></PARAM>
  </SHOPITEM>
  <SHOPITEM>
    <ITEM_ID>0745314610056</ITEM_ID>
    <PRODUCTNAME>Junior | Zelená</PRODUCTNAME>
    <URL>https://segocz.cz/produkty/detail/junior?color=zelena</URL>
    <EAN>0745314610056</EAN>
    <IMGURL>https://segocz.cz/src/Frontend/Files/Catalog/VariantImages/7/previewImg-7-1_cs.jpg</IMGURL>
    <IMGURL_ALTERNATIVE>https://segocz.cz/green.jpg</IMGURL_ALTERNATIVE>
    <PRICE_VAT>4623.00</PRICE_VAT>
    <DELIVERY_DATE>0</DELIVERY_DATE>
    <PARAM><PARAM_NAME><![CDATA[Barva]]></PARAM_NAME><VAL><![CDATA[Zelená]]></VAL></PARAM>
  </SHOPITEM>
  <SHOPITEM>
    <ITEM_ID>0745314610070</ITEM_ID>
    <PRODUCTNAME>Junior | Modrá</PRODUCTNAME>
    <URL>https://segocz.cz/produkty/detail/junior?color=modra</URL>
    <EAN>0745314610070</EAN>
    <IMGURL>https://segocz.cz/src/Frontend/Files/Catalog/VariantImages/7/previewImg-7-2_cs.jpg</IMGURL>
    <IMGURL_ALTERNATIVE>https://segocz.cz/blue.jpg</IMGURL_ALTERNATIVE>
    <PRICE_VAT>4623.00</PRICE_VAT>
    <DELIVERY_DATE>0</DELIVERY_DATE>
    <PARAM><PARAM_NAME><![CDATA[Barva]]></PARAM_NAME><VAL><![CDATA[Modrá]]></VAL></PARAM>
  </SHOPITEM>
</SHOP>`

	feed, stats, err := ParseProducts(context.Background(), strings.NewReader(input), ProductsOptions{})
	if err != nil {
		t.Fatalf("parse SEGO products: %v", err)
	}
	if stats.ProductsRead != 3 || stats.ProductsEmitted != 1 || stats.ItemsWithVariants != 1 || stats.VariantsEmitted != 3 {
		t.Fatalf("unexpected stats: %#v", stats)
	}
	if len(feed.Items) != 1 {
		t.Fatalf("expected 1 grouped item, got %d", len(feed.Items))
	}

	item := feed.Items[0]
	if item.Code != "" || item.Name != "Junior" || item.EAN != "" {
		t.Fatalf("unexpected parent item: %#v", item)
	}
	if item.Description != "Pevná dětská židle" {
		t.Fatalf("unexpected parent description: %q", item.Description)
	}
	if item.DefaultCategory == nil || *item.DefaultCategory != (shoptet.Category{ID: "1125", Path: "KANCELÁŘSKÉ ŽIDLE A KŘESLA > DĚTSKÉ ŽIDLE"}) {
		t.Fatalf("unexpected parent category: %#v", item.DefaultCategory)
	}
	if len(item.Images) != 3 {
		t.Fatalf("expected merged variant images, got %#v", item.Images)
	}
	if len(item.Variants) != 3 {
		t.Fatalf("expected 3 variants, got %#v", item.Variants)
	}

	first := item.Variants[0]
	if first.Code != "0745314610063" || first.EAN != "0745314610063" || first.PriceVAT != "4623" || first.VAT != "21" || first.Currency != "CZK" || first.Availability != "Skladem" {
		t.Fatalf("unexpected first variant: %#v", first)
	}
	if len(first.Parameters) != 1 || first.Parameters[0] != (shoptet.Parameter{Name: "Barva", Value: "Červená"}) {
		t.Fatalf("unexpected first variant parameters: %#v", first.Parameters)
	}
	if first.ImageRef != "https://segocz.cz/red.jpg" {
		t.Fatalf("unexpected first variant image ref: %q", first.ImageRef)
	}
}

func TestParseProductsMapsSegoSubcategories(t *testing.T) {
	input := `<SHOP>
  <SHOPITEM>
    <ITEM_ID>1</ITEM_ID>
    <PRODUCTNAME>Kolečko univerzální | Černá</PRODUCTNAME>
    <DESCRIPTION>Náhradní díl k židli</DESCRIPTION>
  </SHOPITEM>
  <SHOPITEM>
    <ITEM_ID>2</ITEM_ID>
    <PRODUCTNAME>Medical | koženka MECAMED</PRODUCTNAME>
    <DESCRIPTION>Otočná židle vhodná nejen do zdravotnických zařízení.</DESCRIPTION>
  </SHOPITEM>
  <SHOPITEM>
    <ITEM_ID>3</ITEM_ID>
    <PRODUCTNAME>Stream | Černá</PRODUCTNAME>
    <DESCRIPTION>Moderní jednací stohovatelná židle do veřejných prostor.</DESCRIPTION>
  </SHOPITEM>
  <SHOPITEM>
    <ITEM_ID>4</ITEM_ID>
    <PRODUCTNAME>AIR plus | Černá</PRODUCTNAME>
    <PARAM><PARAM_NAME>Čalounění opěráku</PARAM_NAME><VAL>Síťovaný</VAL></PARAM>
  </SHOPITEM>
  <SHOPITEM>
    <ITEM_ID>5</ITEM_ID>
    <PRODUCTNAME>Sirio | Černá</PRODUCTNAME>
    <PARAM><PARAM_NAME>Čalounění opěráku</PARAM_NAME><VAL>Látkový</VAL></PARAM>
  </SHOPITEM>
</SHOP>`

	feed, stats, err := ParseProducts(context.Background(), strings.NewReader(input), ProductsOptions{})
	if err != nil {
		t.Fatalf("parse SEGO products: %v", err)
	}
	if stats.ProductsRead != 5 || stats.ProductsEmitted != 5 || stats.ProductsSkipped != 0 {
		t.Fatalf("unexpected stats: %#v", stats)
	}

	expected := []shoptet.Category{
		{ID: "896", Path: "KANCELÁŘSKÉ ŽIDLE A KŘESLA > NÁHRADNÍ DÍLY A PODLOŽKY"},
		{ID: "1230", Path: "LABORATORNÍ ŽIDLE A LAVICE > ZDRAVOTNÍ ŽIDLE"},
		{ID: "1146", Path: "ŽIDLE > KONFERENČNÍ ŽIDLE"},
		{ID: "1284", Path: "KANCELÁŘSKÉ ŽIDLE A KŘESLA > SÍŤOVANÉ KANCELÁŘSKÉ ŽIDLE"},
		{ID: "1275", Path: "KANCELÁŘSKÉ ŽIDLE A KŘESLA > ČALOUNĚNÁ KANCELÁŘSKÁ KŘESLA"},
	}
	for index, category := range expected {
		if feed.Items[index].DefaultCategory == nil || *feed.Items[index].DefaultCategory != category {
			t.Fatalf("item %d expected category %#v, got %#v", index, category, feed.Items[index].DefaultCategory)
		}
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
	if stats.ProductsRead != 2 || stats.ProductsEmitted != 1 {
		t.Fatalf("unexpected stats: %#v", stats)
	}
	if len(feed.Items) != 1 || feed.Items[0].Code != "1" {
		t.Fatalf("unexpected feed: %#v", feed)
	}
}

func TestParseProductsPreferVariantItemsKeepsCompleteVariantGroups(t *testing.T) {
	input := `<SHOP>
  <SHOPITEM><ITEM_ID>SIMPLE-1</ITEM_ID><PRODUCTNAME>Simple 1</PRODUCTNAME></SHOPITEM>
  <SHOPITEM><ITEM_ID>A-RED</ITEM_ID><PRODUCTNAME>Variant A | Červená</PRODUCTNAME><URL>https://segocz.cz/produkty/detail/variant-a?color=cervena</URL><PARAM><PARAM_NAME>Barva</PARAM_NAME><VAL>Červená</VAL></PARAM></SHOPITEM>
  <SHOPITEM><ITEM_ID>A-BLUE</ITEM_ID><PRODUCTNAME>Variant A | Modrá</PRODUCTNAME><URL>https://segocz.cz/produkty/detail/variant-a?color=modra</URL><PARAM><PARAM_NAME>Barva</PARAM_NAME><VAL>Modrá</VAL></PARAM></SHOPITEM>
  <SHOPITEM><ITEM_ID>SIMPLE-2</ITEM_ID><PRODUCTNAME>Simple 2</PRODUCTNAME></SHOPITEM>
  <SHOPITEM><ITEM_ID>B-RED</ITEM_ID><PRODUCTNAME>Variant B | Červená</PRODUCTNAME><URL>https://segocz.cz/produkty/detail/variant-b?color=cervena</URL><PARAM><PARAM_NAME>Barva</PARAM_NAME><VAL>Červená</VAL></PARAM></SHOPITEM>
  <SHOPITEM><ITEM_ID>B-GREEN</ITEM_ID><PRODUCTNAME>Variant B | Zelená</PRODUCTNAME><URL>https://segocz.cz/produkty/detail/variant-b?color=zelena</URL><PARAM><PARAM_NAME>Barva</PARAM_NAME><VAL>Zelená</VAL></PARAM></SHOPITEM>
</SHOP>`

	feed, stats, err := ParseProducts(context.Background(), strings.NewReader(input), ProductsOptions{
		MaxProducts:        2,
		PreferVariantItems: true,
	})
	if err != nil {
		t.Fatalf("parse SEGO products: %v", err)
	}
	if stats.ProductsRead != 6 || stats.ProductsEmitted != 2 || stats.ItemsWithVariants != 2 || stats.VariantsEmitted != 4 {
		t.Fatalf("unexpected stats: %#v", stats)
	}
	if len(feed.Items) != 2 {
		t.Fatalf("expected 2 items, got %#v", feed.Items)
	}
	if feed.Items[0].Name != "Variant A" || len(feed.Items[0].Variants) != 2 {
		t.Fatalf("expected complete Variant A group first, got %#v", feed.Items[0])
	}
	if feed.Items[1].Name != "Variant B" || len(feed.Items[1].Variants) != 2 {
		t.Fatalf("expected complete Variant B group second, got %#v", feed.Items[1])
	}
}

package drevocal

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/fanda/arpari-xml-feed-parser-railway/internal/shoptet"
)

func TestParseProductsGroupsVariantsAndMapsFields(t *testing.T) {
	input := `<SHOP generated="2026-05-21T12:57:57+02:00">
  <SHOPITEM>
    <ITEM_ID>5211112</ITEM_ID>
    <ITEMGROUP_ID>521</ITEMGROUP_ID>
    <PRODUCTNAME>Matrace Eliška 195x80x19 Úplet</PRODUCTNAME>
    <MANUFACTURER>DŘEVOČAL</MANUFACTURER>
    <PRICE_VAT>3588.00</PRICE_VAT>
    <CURRENCY>CZK</CURRENCY>
    <EAN>8596723002176</EAN>
    <DESCRIPTION><![CDATA[Eliška je ideální volbou.]]></DESCRIPTION>
    <URL>https://www.matrace-drevocal.cz/eliska</URL>
    <IMGURL>https://www.matrace-drevocal.cz/eliska.jpg</IMGURL>
    <AVAILABILITY>Skladem</AVAILABILITY>
    <GIFT>polštář Lukáš</GIFT>
    <PARAM><PARAM_NAME>Rozměr</PARAM_NAME><VAL>195x80</VAL></PARAM>
    <PARAM><PARAM_NAME>Výška</PARAM_NAME><VAL>19 cm</VAL></PARAM>
    <PARAM><PARAM_NAME>Potah</PARAM_NAME><VAL>Úplet</VAL></PARAM>
  </SHOPITEM>
  <SHOPITEM>
    <ITEM_ID>5211113</ITEM_ID>
    <ITEMGROUP_ID>521</ITEMGROUP_ID>
    <PRODUCTNAME>Matrace Eliška 200x90x19 Aloe Vera</PRODUCTNAME>
    <MANUFACTURER>DŘEVOČAL</MANUFACTURER>
    <PRICE_VAT>4590.00</PRICE_VAT>
    <CURRENCY>CZK</CURRENCY>
    <EAN>8596723002177</EAN>
    <DESCRIPTION><![CDATA[Eliška je ideální volbou.]]></DESCRIPTION>
    <URL>https://www.matrace-drevocal.cz/eliska</URL>
    <IMGURL>https://www.matrace-drevocal.cz/eliska.jpg</IMGURL>
    <AVAILABILITY>Momentálně nedostupné</AVAILABILITY>
    <GIFT>polštář Lukáš</GIFT>
    <PARAM><PARAM_NAME>Rozměr</PARAM_NAME><VAL>200x90</VAL></PARAM>
    <PARAM><PARAM_NAME>Výška</PARAM_NAME><VAL>19 cm</VAL></PARAM>
    <PARAM><PARAM_NAME>Potah</PARAM_NAME><VAL>Aloe Vera</VAL></PARAM>
  </SHOPITEM>
</SHOP>`

	feed, stats, err := ParseProducts(context.Background(), strings.NewReader(input), ProductsOptions{})
	if err != nil {
		t.Fatalf("parse Dřevočal products: %v", err)
	}
	if stats.ProductsRead != 2 || stats.ProductsEmitted != 1 || stats.ProductsSkipped != 0 || stats.ItemsWithVariants != 1 || stats.VariantsEmitted != 2 {
		t.Fatalf("unexpected stats: %#v", stats)
	}
	if len(feed.Items) != 1 {
		t.Fatalf("expected 1 grouped item, got %#v", feed.Items)
	}

	item := feed.Items[0]
	if item.Code != "DREVOCAL-521" || item.Name != "Matrace Eliška" {
		t.Fatalf("unexpected parent item: %#v", item)
	}
	if item.Supplier != "DŘEVOČAL" || item.Manufacturer != "DŘEVOČAL" {
		t.Fatalf("unexpected supplier/manufacturer: %#v", item)
	}
	if item.Description != "" {
		t.Fatalf("expected Dřevočal description to be omitted, got %q", item.Description)
	}
	if item.DefaultCategory == nil || *item.DefaultCategory != (shoptet.Category{ID: "1188", Path: "LOŽNICE > MATRACE"}) {
		t.Fatalf("unexpected category: %#v", item.DefaultCategory)
	}
	if len(item.Images) != 1 || item.Images[0].URL != "https://www.matrace-drevocal.cz/eliska.jpg" {
		t.Fatalf("unexpected images: %#v", item.Images)
	}
	if len(item.InformationParameters) != 1 || item.InformationParameters[0] != (shoptet.Parameter{Name: "Dárek", Value: "polštář Lukáš"}) {
		t.Fatalf("unexpected information parameters: %#v", item.InformationParameters)
	}
	if len(item.Variants) != 2 {
		t.Fatalf("expected 2 variants, got %#v", item.Variants)
	}

	first := item.Variants[0]
	if first.Code != "5211112" || first.EAN != "8596723002176" || first.PriceVAT != "3588.00" || first.Currency != "CZK" || first.VAT != "21" || first.Availability != "Skladem" {
		t.Fatalf("unexpected first variant: %#v", first)
	}
	expectedParameters := []shoptet.Parameter{
		{Name: "Rozměr", Value: "195x80"},
		{Name: "Výška", Value: "19 cm"},
		{Name: "Potah", Value: "Úplet"},
	}
	for index, expected := range expectedParameters {
		if first.Parameters[index] != expected {
			t.Fatalf("unexpected first variant parameter %d: %#v", index, first.Parameters[index])
		}
	}
	if item.Variants[1].Availability != "Momentálně nedostupné" {
		t.Fatalf("unexpected second variant availability: %#v", item.Variants[1])
	}
}

func TestParseProductsLimitsOutputProducts(t *testing.T) {
	input := `<SHOP>
  <SHOPITEM><ITEM_ID>1</ITEM_ID><ITEMGROUP_ID>401</ITEMGROUP_ID><PRODUCTNAME>Matrace Milena 195x80x10 Úplet</PRODUCTNAME><PRICE_VAT>2650.00</PRICE_VAT><PARAM><PARAM_NAME>Rozměr</PARAM_NAME><VAL>195x80</VAL></PARAM><PARAM><PARAM_NAME>Výška</PARAM_NAME><VAL>10 cm</VAL></PARAM><PARAM><PARAM_NAME>Potah</PARAM_NAME><VAL>Úplet</VAL></PARAM></SHOPITEM>
  <SHOPITEM><ITEM_ID>2</ITEM_ID><ITEMGROUP_ID>403</ITEMGROUP_ID><PRODUCTNAME>Matrace Hana 195x80x14 Úplet</PRODUCTNAME><PRICE_VAT>3679.00</PRICE_VAT><PARAM><PARAM_NAME>Rozměr</PARAM_NAME><VAL>195x80</VAL></PARAM><PARAM><PARAM_NAME>Výška</PARAM_NAME><VAL>14 cm</VAL></PARAM><PARAM><PARAM_NAME>Potah</PARAM_NAME><VAL>Úplet</VAL></PARAM></SHOPITEM>
</SHOP>`

	feed, stats, err := ParseProducts(context.Background(), strings.NewReader(input), ProductsOptions{MaxProducts: 1})
	if err != nil {
		t.Fatalf("parse limited Dřevočal products: %v", err)
	}
	if stats.ProductsRead != 2 || stats.ProductsEmitted != 1 {
		t.Fatalf("unexpected stats: %#v", stats)
	}
	if len(feed.Items) != 1 || feed.Items[0].Code != "DREVOCAL-401" {
		t.Fatalf("unexpected feed: %#v", feed)
	}
}

func TestParseProductsTrimsVariantsAboveLimit(t *testing.T) {
	var input bytes.Buffer
	input.WriteString(`<SHOP>`)
	for index := range 3 {
		fmt.Fprintf(&input, `<SHOPITEM><ITEM_ID>521%d</ITEM_ID><ITEMGROUP_ID>521</ITEMGROUP_ID><PRODUCTNAME>Matrace Eliška 195x80x19 Úplet</PRODUCTNAME><PRICE_VAT>3588.00</PRICE_VAT><PARAM><PARAM_NAME>Rozměr</PARAM_NAME><VAL>195x80</VAL></PARAM><PARAM><PARAM_NAME>Výška</PARAM_NAME><VAL>19 cm</VAL></PARAM><PARAM><PARAM_NAME>Potah</PARAM_NAME><VAL>Úplet %d</VAL></PARAM></SHOPITEM>`, index, index)
	}
	input.WriteString(`</SHOP>`)

	feed, stats, err := ParseProducts(context.Background(), &input, ProductsOptions{MaxVariantsPerProduct: 2})
	if err != nil {
		t.Fatalf("parse Dřevočal products with variant limit: %v", err)
	}
	if len(feed.Items) != 1 || len(feed.Items[0].Variants) != 2 {
		t.Fatalf("unexpected feed: %#v", feed)
	}
	if stats.ProductsTrimmed != 1 || stats.VariantsTrimmed != 1 || stats.VariantsEmitted != 2 {
		t.Fatalf("unexpected trim stats: %#v", stats)
	}
}

func TestParseProductsSkipsMissingIdentityOrRequiredParameter(t *testing.T) {
	input := `<SHOP>
  <SHOPITEM><ITEMGROUP_ID>521</ITEMGROUP_ID><PRODUCTNAME>Matrace bez kódu</PRODUCTNAME><PARAM><PARAM_NAME>Rozměr</PARAM_NAME><VAL>195x80</VAL></PARAM><PARAM><PARAM_NAME>Výška</PARAM_NAME><VAL>19 cm</VAL></PARAM><PARAM><PARAM_NAME>Potah</PARAM_NAME><VAL>Úplet</VAL></PARAM></SHOPITEM>
  <SHOPITEM><ITEM_ID>5211112</ITEM_ID><ITEMGROUP_ID>521</ITEMGROUP_ID><PRODUCTNAME>Matrace bez potahu</PRODUCTNAME><PARAM><PARAM_NAME>Rozměr</PARAM_NAME><VAL>195x80</VAL></PARAM><PARAM><PARAM_NAME>Výška</PARAM_NAME><VAL>19 cm</VAL></PARAM></SHOPITEM>
  <SHOPITEM><ITEM_ID>5211113</ITEM_ID><ITEMGROUP_ID>521</ITEMGROUP_ID><PRODUCTNAME>Matrace Eliška 195x80x19 Úplet</PRODUCTNAME><PARAM><PARAM_NAME>Rozměr</PARAM_NAME><VAL>195x80</VAL></PARAM><PARAM><PARAM_NAME>Výška</PARAM_NAME><VAL>19 cm</VAL></PARAM><PARAM><PARAM_NAME>Potah</PARAM_NAME><VAL>Úplet</VAL></PARAM></SHOPITEM>
</SHOP>`

	feed, stats, err := ParseProducts(context.Background(), strings.NewReader(input), ProductsOptions{})
	if err != nil {
		t.Fatalf("parse Dřevočal products: %v", err)
	}
	if stats.ProductsRead != 3 || stats.ProductsSkipped != 2 || stats.VariantsSkipped != 2 || stats.ProductsEmitted != 1 {
		t.Fatalf("unexpected stats: %#v", stats)
	}
	if len(feed.Items) != 1 || len(feed.Items[0].Variants) != 1 {
		t.Fatalf("unexpected feed: %#v", feed)
	}
}

func TestParseProductsKeepsVariantWithoutEAN(t *testing.T) {
	input := `<SHOP>
  <SHOPITEM>
    <ITEM_ID>473114</ITEM_ID>
    <ITEMGROUP_ID>473</ITEMGROUP_ID>
    <PRODUCTNAME>Matrace Ester 200x100x25 Úplet</PRODUCTNAME>
    <PRICE_VAT>9990.00</PRICE_VAT>
    <PARAM><PARAM_NAME>Rozměr</PARAM_NAME><VAL>200x100</VAL></PARAM>
    <PARAM><PARAM_NAME>Výška</PARAM_NAME><VAL>25 cm</VAL></PARAM>
    <PARAM><PARAM_NAME>Potah</PARAM_NAME><VAL>Úplet</VAL></PARAM>
  </SHOPITEM>
</SHOP>`

	feed, stats, err := ParseProducts(context.Background(), strings.NewReader(input), ProductsOptions{})
	if err != nil {
		t.Fatalf("parse Dřevočal products: %v", err)
	}
	if stats.ProductsSkipped != 0 || stats.VariantsEmitted != 1 {
		t.Fatalf("unexpected stats: %#v", stats)
	}
	if feed.Items[0].Variants[0].EAN != "" {
		t.Fatalf("expected empty EAN to be preserved, got %#v", feed.Items[0].Variants[0])
	}
}

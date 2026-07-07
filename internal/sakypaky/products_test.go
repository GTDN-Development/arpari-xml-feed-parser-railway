package sakypaky

import (
	"context"
	"strings"
	"testing"
)

func TestParseProductsSimpleProductMapsFields(t *testing.T) {
	feed, stats, err := ParseProducts(context.Background(), strings.NewReader(`<SHOP>
  <SHOPITEM>
    <PRODUCTNAME>SakyPaky náhradní náplň EPS</PRODUCTNAME>
    <DESCRIPTION>  Náhradní&#160;náplň do vaku. </DESCRIPTION>
    <IMGURL>https://example.test/main.jpg</IMGURL>
    <IMGURL_ALTERNATIVE>https://example.test/alt.jpg</IMGURL_ALTERNATIVE>
    <IMGURL_ALTERNATIVE>https://example.test/alt.jpg</IMGURL_ALTERNATIVE>
    <PRICE_VAT>399.00</PRICE_VAT>
    <DELIVERY_DATE>7</DELIVERY_DATE>
    <CATEGORYTEXT>Půjčovna a servis - Náhradní náplně do vaků, opravné sety</CATEGORYTEXT>
    <MANUFACTURER>SakyPaky</MANUFACTURER>
    <CODE>NAPLN-01</CODE>
    <EAN>1234567890123</EAN>
  </SHOPITEM>
</SHOP>`), ProductsOptions{})
	if err != nil {
		t.Fatalf("parse Sakypaky products: %v", err)
	}
	if stats.ProductsRead != 1 || stats.ProductsEmitted != 1 || stats.ProductsSkipped != 0 {
		t.Fatalf("unexpected stats: %#v", stats)
	}
	if len(feed.Items) != 1 {
		t.Fatalf("expected 1 item, got %#v", feed.Items)
	}

	item := feed.Items[0]
	if item.Code != "NAPLN-01" || item.Name != "SakyPaky náhradní náplň EPS" {
		t.Fatalf("unexpected item identity: %#v", item)
	}
	if item.Description != "Náhradní náplň do vaku." || item.PriceVAT != "399" || item.Currency != "CZK" || item.VAT != "21" {
		t.Fatalf("unexpected item fields: %#v", item)
	}
	if item.Availability != "Dodání 7 dnů" || item.EAN != "1234567890123" || item.Supplier != supplierName {
		t.Fatalf("unexpected logistics fields: %#v", item)
	}
	if len(item.Images) != 2 || item.Images[0].URL != "https://example.test/main.jpg" || item.Images[1].URL != "https://example.test/alt.jpg" {
		t.Fatalf("unexpected images: %#v", item.Images)
	}
	if item.DefaultCategory == nil || item.DefaultCategory.ID != "914" {
		t.Fatalf("unexpected category: %#v", item.DefaultCategory)
	}
}

func TestParseProductsGroupsVariantsByItemGroupIDAndColor(t *testing.T) {
	feed, stats, err := ParseProducts(context.Background(), strings.NewReader(`<SHOP>
  <SHOPITEM>
    <PRODUCTNAME>SakyPaky Hruška červená</PRODUCTNAME>
    <DESCRIPTION>Popis vaku.</DESCRIPTION>
    <IMGURL>https://example.test/red.jpg</IMGURL>
    <PRICE_VAT>1790.00</PRICE_VAT>
    <DELIVERY_DATE>0</DELIVERY_DATE>
    <CATEGORYTEXT>Nábytek | Obývací pokoj | Křesla a taburety | Sedací vaky</CATEGORYTEXT>
    <CODE>HRUSKA-RED</CODE>
    <ITEMGROUP_ID>G-1</ITEMGROUP_ID>
    <EAN>111</EAN>
    <PARAM><PARAM_NAME>Barva</PARAM_NAME><VAL>červená</VAL></PARAM>
  </SHOPITEM>
  <SHOPITEM>
    <PRODUCTNAME>SakyPaky Hruška modrá</PRODUCTNAME>
    <IMGURL>https://example.test/blue.jpg</IMGURL>
    <PRICE_VAT>1790.00</PRICE_VAT>
    <DELIVERY_DATE>3</DELIVERY_DATE>
    <CATEGORYTEXT>Nábytek | Obývací pokoj | Křesla a taburety | Sedací vaky</CATEGORYTEXT>
    <CODE>HRUSKA-BLUE</CODE>
    <ITEMGROUP_ID>G-1</ITEMGROUP_ID>
    <EAN>222</EAN>
    <PARAM><PARAM_NAME>Barva</PARAM_NAME><VAL>modrá</VAL></PARAM>
  </SHOPITEM>
</SHOP>`), ProductsOptions{})
	if err != nil {
		t.Fatalf("parse Sakypaky products: %v", err)
	}
	if stats.ProductsRead != 2 || stats.ProductsEmitted != 1 || stats.ItemsWithVariants != 1 || stats.VariantsEmitted != 2 {
		t.Fatalf("unexpected stats: %#v", stats)
	}
	if len(feed.Items) != 1 {
		t.Fatalf("expected 1 grouped item, got %#v", feed.Items)
	}

	item := feed.Items[0]
	if item.Code != "SAKYPAKY-G-1" || item.Name != "SakyPaky Hruška" {
		t.Fatalf("unexpected grouped item: %#v", item)
	}
	if len(item.Variants) != 2 {
		t.Fatalf("expected 2 variants, got %#v", item.Variants)
	}
	first := item.Variants[0]
	if first.Code != "HRUSKA-RED" || first.EAN != "111" || first.Availability != "Skladem" || first.ImageRef != "https://example.test/red.jpg" {
		t.Fatalf("unexpected first variant: %#v", first)
	}
	if len(first.Parameters) != 1 || first.Parameters[0].Name != "Barva" || first.Parameters[0].Value != "červená" {
		t.Fatalf("unexpected first variant parameters: %#v", first.Parameters)
	}
	second := item.Variants[1]
	if second.Code != "HRUSKA-BLUE" || second.Availability != "Dodání 3 dnů" || second.Parameters[0].Value != "modrá" {
		t.Fatalf("unexpected second variant: %#v", second)
	}
}

func TestParseProductsTrimsPartialColorFromParentName(t *testing.T) {
	feed, _, err := ParseProducts(context.Background(), strings.NewReader(`<SHOP>
  <SHOPITEM>
    <PRODUCTNAME>SakyPaky Klííídek LUX sedací vak šedá</PRODUCTNAME>
    <CATEGORYTEXT>Nábytek | Obývací pokoj | Křesla a taburety | Sedací vaky</CATEGORYTEXT>
    <CODE>LUX-BEIGE</CODE>
    <ITEMGROUP_ID>G-PARTIAL</ITEMGROUP_ID>
    <PARAM><PARAM_NAME>Barva</PARAM_NAME><VAL>béžová</VAL></PARAM>
  </SHOPITEM>
  <SHOPITEM>
    <PRODUCTNAME>SakyPaky Klííídek LUX sedací vak šedá</PRODUCTNAME>
    <CATEGORYTEXT>Nábytek | Obývací pokoj | Křesla a taburety | Sedací vaky</CATEGORYTEXT>
    <CODE>LUX-GREY</CODE>
    <ITEMGROUP_ID>G-PARTIAL</ITEMGROUP_ID>
    <PARAM><PARAM_NAME>Barva</PARAM_NAME><VAL>šedá ocelová</VAL></PARAM>
  </SHOPITEM>
</SHOP>`), ProductsOptions{})
	if err != nil {
		t.Fatalf("parse Sakypaky products: %v", err)
	}
	if got := feed.Items[0].Name; got != "SakyPaky Klííídek LUX sedací vak" {
		t.Fatalf("expected partial color to be trimmed from parent name, got %q", got)
	}
}

func TestParseProductsNormalizesDusinkaVariantSeries(t *testing.T) {
	feed, _, err := ParseProducts(context.Background(), strings.NewReader(`<SHOP>
  <SHOPITEM>
    <PRODUCTNAME>SakyPaky taburet Dušinka ANTONIE</PRODUCTNAME>
    <CATEGORYTEXT>Nábytek | Obývací pokoj | Křesla a taburety | Taburety</CATEGORYTEXT>
    <CODE>8005207902</CODE>
    <ITEMGROUP_ID>G-DUSINKA-TABURET</ITEMGROUP_ID>
    <PARAM><PARAM_NAME>Barva</PARAM_NAME><VAL>Dušinka ANTONIE</VAL></PARAM>
  </SHOPITEM>
  <SHOPITEM>
    <PRODUCTNAME>SakyPaky taburet Dušinka EMA</PRODUCTNAME>
    <CATEGORYTEXT>Nábytek | Obývací pokoj | Křesla a taburety | Taburety</CATEGORYTEXT>
    <CODE>8005207925</CODE>
    <ITEMGROUP_ID>G-DUSINKA-TABURET</ITEMGROUP_ID>
    <PARAM><PARAM_NAME>Barva</PARAM_NAME><VAL>Dušinka EMA</VAL></PARAM>
  </SHOPITEM>
</SHOP>`), ProductsOptions{})
	if err != nil {
		t.Fatalf("parse Sakypaky products: %v", err)
	}
	if len(feed.Items) != 1 {
		t.Fatalf("expected 1 grouped item, got %#v", feed.Items)
	}

	item := feed.Items[0]
	if item.Name != "SakyPaky taburet Dušinka" {
		t.Fatalf("expected Dušinka to remain in parent name, got %q", item.Name)
	}
	if len(item.Variants) != 2 {
		t.Fatalf("expected 2 variants, got %#v", item.Variants)
	}
	if item.Variants[0].Parameters[0].Value != "ANTONIE" || item.Variants[1].Parameters[0].Value != "EMA" {
		t.Fatalf("expected Dušinka prefix to be removed from variant values, got %#v", item.Variants)
	}
}

func TestParseProductsNormalizesZiziDusinkaVariantSeries(t *testing.T) {
	feed, _, err := ParseProducts(context.Background(), strings.NewReader(`<SHOP>
  <SHOPITEM>
    <PRODUCTNAME>SakyPaky sedací vak Žiži Dušinka ANTONIE</PRODUCTNAME>
    <CATEGORYTEXT>Nábytek | Obývací pokoj | Křesla a taburety | Sedací vaky</CATEGORYTEXT>
    <CODE>8005212902</CODE>
    <ITEMGROUP_ID>G-DUSINKA-ZIZI</ITEMGROUP_ID>
    <PARAM><PARAM_NAME>Barva</PARAM_NAME><VAL>Dušinka ANTONIE</VAL></PARAM>
  </SHOPITEM>
  <SHOPITEM>
    <PRODUCTNAME>SakyPaky sedací vak Žiži Dušinka EMA</PRODUCTNAME>
    <CATEGORYTEXT>Nábytek | Obývací pokoj | Křesla a taburety | Sedací vaky</CATEGORYTEXT>
    <CODE>8005212925</CODE>
    <ITEMGROUP_ID>G-DUSINKA-ZIZI</ITEMGROUP_ID>
    <PARAM><PARAM_NAME>Barva</PARAM_NAME><VAL>Dušinka EMA</VAL></PARAM>
  </SHOPITEM>
</SHOP>`), ProductsOptions{})
	if err != nil {
		t.Fatalf("parse Sakypaky products: %v", err)
	}
	if len(feed.Items) != 1 {
		t.Fatalf("expected 1 grouped item, got %#v", feed.Items)
	}

	item := feed.Items[0]
	if item.Name != "SakyPaky sedací vak Žiži Dušinka" {
		t.Fatalf("expected Dušinka to remain in parent name, got %q", item.Name)
	}
	if item.Variants[0].Parameters[0].Value != "ANTONIE" || item.Variants[1].Parameters[0].Value != "EMA" {
		t.Fatalf("expected Dušinka prefix to be removed from variant values, got %#v", item.Variants)
	}
}

func TestParseProductsKeepsMixedDusinkaVariantValues(t *testing.T) {
	feed, _, err := ParseProducts(context.Background(), strings.NewReader(`<SHOP>
  <SHOPITEM>
    <PRODUCTNAME>SakyPaky sedací vak Klííídek KYTI, Dušinka PŘÁTELSTVÍ 2</PRODUCTNAME>
    <CATEGORYTEXT>Nábytek | Obývací pokoj | Křesla a taburety | Sedací vaky</CATEGORYTEXT>
    <CODE>KYTI-DUSINKA</CODE>
    <ITEMGROUP_ID>G-MIXED</ITEMGROUP_ID>
    <PARAM><PARAM_NAME>Barva</PARAM_NAME><VAL>Dušinka PŘÁTELSTVÍ 2</VAL></PARAM>
  </SHOPITEM>
  <SHOPITEM>
    <PRODUCTNAME>SakyPaky sedací vak Klííídek KYTI, Camo</PRODUCTNAME>
    <CATEGORYTEXT>Nábytek | Obývací pokoj | Křesla a taburety | Sedací vaky</CATEGORYTEXT>
    <CODE>KYTI-CAMO</CODE>
    <ITEMGROUP_ID>G-MIXED</ITEMGROUP_ID>
    <PARAM><PARAM_NAME>Barva</PARAM_NAME><VAL>Camo</VAL></PARAM>
  </SHOPITEM>
</SHOP>`), ProductsOptions{})
	if err != nil {
		t.Fatalf("parse Sakypaky products: %v", err)
	}
	if len(feed.Items) != 1 {
		t.Fatalf("expected 1 grouped item, got %#v", feed.Items)
	}

	variants := feed.Items[0].Variants
	if variants[0].Parameters[0].Value != "Dušinka PŘÁTELSTVÍ 2" || variants[1].Parameters[0].Value != "Camo" {
		t.Fatalf("expected mixed group values to stay unchanged, got %#v", variants)
	}
}

func TestParseProductsDisambiguatesDuplicateColors(t *testing.T) {
	feed, _, err := ParseProducts(context.Background(), strings.NewReader(`<SHOP>
  <SHOPITEM>
    <PRODUCTNAME>SakyPaky Hruška zelená</PRODUCTNAME>
    <CATEGORYTEXT>Nábytek | Obývací pokoj | Křesla a taburety | Sedací vaky</CATEGORYTEXT>
    <CODE>HRUSKA-1</CODE>
    <ITEMGROUP_ID>G-2</ITEMGROUP_ID>
    <PARAM><PARAM_NAME>Barva</PARAM_NAME><VAL>zelená</VAL></PARAM>
  </SHOPITEM>
  <SHOPITEM>
    <PRODUCTNAME>SakyPaky Hruška zelená</PRODUCTNAME>
    <CATEGORYTEXT>Nábytek | Obývací pokoj | Křesla a taburety | Sedací vaky</CATEGORYTEXT>
    <CODE>HRUSKA-2</CODE>
    <ITEMGROUP_ID>G-2</ITEMGROUP_ID>
    <PARAM><PARAM_NAME>Barva</PARAM_NAME><VAL>zelená</VAL></PARAM>
  </SHOPITEM>
</SHOP>`), ProductsOptions{})
	if err != nil {
		t.Fatalf("parse Sakypaky products: %v", err)
	}

	variants := feed.Items[0].Variants
	if variants[0].Parameters[0].Value != "zelená (HRUSKA-1)" || variants[1].Parameters[0].Value != "zelená (HRUSKA-2)" {
		t.Fatalf("expected duplicate colors to be disambiguated, got %#v", variants)
	}
}

func TestParseProductsSkipsExcludedAndUnknownCategories(t *testing.T) {
	feed, stats, err := ParseProducts(context.Background(), strings.NewReader(`<SHOP>
  <SHOPITEM>
    <PRODUCTNAME>Jmenovka pelechvaku</PRODUCTNAME>
    <CATEGORYTEXT>Hobby » Obalové materiály » Etikety</CATEGORYTEXT>
    <CODE>ETIKETA</CODE>
  </SHOPITEM>
  <SHOPITEM>
    <PRODUCTNAME>Pelech pro psa</PRODUCTNAME>
    <CATEGORYTEXT>Hobby | Chovatelské potřeby | Pro psy | Pelíšky</CATEGORYTEXT>
    <CODE>PELECH</CODE>
  </SHOPITEM>
  <SHOPITEM>
    <PRODUCTNAME>Neznámý produkt</PRODUCTNAME>
    <CATEGORYTEXT>Neznámá kategorie</CATEGORYTEXT>
    <CODE>UNKNOWN</CODE>
  </SHOPITEM>
</SHOP>`), ProductsOptions{})
	if err != nil {
		t.Fatalf("parse Sakypaky products: %v", err)
	}
	if len(feed.Items) != 0 {
		t.Fatalf("expected no emitted items, got %#v", feed.Items)
	}
	if stats.ProductsRead != 3 || stats.ProductsSkipped != 3 {
		t.Fatalf("unexpected stats: %#v", stats)
	}
}

func TestParseProductsMapsCategories(t *testing.T) {
	feed, _, err := ParseProducts(context.Background(), strings.NewReader(`<SHOP>
  <SHOPITEM><PRODUCTNAME>SakyPaky Taburet Queen</PRODUCTNAME><CATEGORYTEXT>Nábytek | Obývací pokoj | Křesla a taburety | Taburety</CATEGORYTEXT><CODE>TAB</CODE></SHOPITEM>
  <SHOPITEM><PRODUCTNAME>SakyPaky houpačka</PRODUCTNAME><CATEGORYTEXT>Sedací vaky a taburety - Houpačky</CATEGORYTEXT><CODE>HOU</CODE></SHOPITEM>
  <SHOPITEM><PRODUCTNAME>SakyPaky skládací stolek</PRODUCTNAME><CATEGORYTEXT>Sedací vaky a taburety - Ostatní produkty</CATEGORYTEXT><CODE>STOL</CODE></SHOPITEM>
  <SHOPITEM><PRODUCTNAME>SakyPaky SET45</PRODUCTNAME><CATEGORYTEXT>Sedací vaky a taburety - Výhodné sety</CATEGORYTEXT><CODE>SET45</CODE></SHOPITEM>
</SHOP>`), ProductsOptions{})
	if err != nil {
		t.Fatalf("parse Sakypaky products: %v", err)
	}
	if len(feed.Items) != 4 {
		t.Fatalf("expected 4 mapped items, got %#v", feed.Items)
	}

	expected := []string{"1155", "1227", "1269", "914"}
	for index, want := range expected {
		if feed.Items[index].DefaultCategory == nil || feed.Items[index].DefaultCategory.ID != want {
			t.Fatalf("item %d expected category %s, got %#v", index, want, feed.Items[index].DefaultCategory)
		}
	}
}

func TestParseProductsLimitsOutputProductsAndPrefersVariantItems(t *testing.T) {
	feed, stats, err := ParseProducts(context.Background(), strings.NewReader(`<SHOP>
  <SHOPITEM><PRODUCTNAME>SakyPaky náhradní náplň</PRODUCTNAME><CATEGORYTEXT>Půjčovna a servis - Náhradní náplně do vaků, opravné sety</CATEGORYTEXT><CODE>SIMPLE</CODE></SHOPITEM>
  <SHOPITEM><PRODUCTNAME>SakyPaky Hruška červená</PRODUCTNAME><CATEGORYTEXT>Nábytek | Obývací pokoj | Křesla a taburety | Sedací vaky</CATEGORYTEXT><CODE>V1</CODE><ITEMGROUP_ID>G-3</ITEMGROUP_ID><PARAM><PARAM_NAME>Barva</PARAM_NAME><VAL>červená</VAL></PARAM></SHOPITEM>
  <SHOPITEM><PRODUCTNAME>SakyPaky Hruška modrá</PRODUCTNAME><CATEGORYTEXT>Nábytek | Obývací pokoj | Křesla a taburety | Sedací vaky</CATEGORYTEXT><CODE>V2</CODE><ITEMGROUP_ID>G-3</ITEMGROUP_ID><PARAM><PARAM_NAME>Barva</PARAM_NAME><VAL>modrá</VAL></PARAM></SHOPITEM>
</SHOP>`), ProductsOptions{MaxProducts: 1, PreferVariantItems: true})
	if err != nil {
		t.Fatalf("parse Sakypaky products: %v", err)
	}
	if len(feed.Items) != 1 || feed.Items[0].Code != "SAKYPAKY-G-3" {
		t.Fatalf("expected variant group to be emitted first, got %#v", feed.Items)
	}
	if stats.ProductsEmitted != 1 {
		t.Fatalf("unexpected stats: %#v", stats)
	}
}

package feed

import (
	"bytes"
	"context"
	"testing"
)

func TestSakypakyTestGenerateLimitsOutputItems(t *testing.T) {
	downloader := &fakeStimaDownloader{
		body: `<SHOP>
  <SHOPITEM><PRODUCTNAME>SakyPaky náhradní náplň</PRODUCTNAME><CATEGORYTEXT>Půjčovna a servis - Náhradní náplně do vaků, opravné sety</CATEGORYTEXT><CODE>SIMPLE</CODE></SHOPITEM>
  <SHOPITEM><PRODUCTNAME>SakyPaky Hruška červená</PRODUCTNAME><CATEGORYTEXT>Nábytek | Obývací pokoj | Křesla a taburety | Sedací vaky</CATEGORYTEXT><CODE>V1</CODE><ITEMGROUP_ID>G-1</ITEMGROUP_ID><PARAM><PARAM_NAME>Barva</PARAM_NAME><VAL>červená</VAL></PARAM></SHOPITEM>
  <SHOPITEM><PRODUCTNAME>SakyPaky Hruška modrá</PRODUCTNAME><CATEGORYTEXT>Nábytek | Obývací pokoj | Křesla a taburety | Sedací vaky</CATEGORYTEXT><CODE>V2</CODE><ITEMGROUP_ID>G-1</ITEMGROUP_ID><PARAM><PARAM_NAME>Barva</PARAM_NAME><VAL>modrá</VAL></PARAM></SHOPITEM>
</SHOP>`,
	}
	generator := SakypakyTest{
		Downloader:  downloader,
		SourceURL:   "https://example.test/sakypaky.xml",
		MaxProducts: 1,
	}

	var output bytes.Buffer
	result, err := generator.Generate(context.Background(), &output)
	if err != nil {
		t.Fatalf("generate Sakypaky test: %v", err)
	}
	if downloader.lastURL != "https://example.test/sakypaky.xml" {
		t.Fatalf("expected custom source URL, got %q", downloader.lastURL)
	}
	if result.ItemsProcessed != 1 || result.ItemsSkipped != 0 {
		t.Fatalf("unexpected result: %#v", result)
	}

	parsed := parseGeneratedDrevocal(t, output.Bytes())
	if len(parsed.Items) != 1 {
		t.Fatalf("expected 1 generated item, got %#v", parsed.Items)
	}
	if parsed.Items[0].ExternalID != "SAKYPAKY-G-1" || parsed.Items[0].Name != "SakyPaky Hruška" {
		t.Fatalf("unexpected generated item: %#v", parsed.Items[0])
	}
	if len(parsed.Items[0].Variants) != 2 || parsed.Items[0].Variants[0].Code != "V1" {
		t.Fatalf("unexpected generated variants: %#v", parsed.Items[0].Variants)
	}
	if parsed.Items[0].Categories.Default.ID != "914" {
		t.Fatalf("unexpected default category: %#v", parsed.Items[0].Categories.Default)
	}
}

func TestSakypakyGenerateUsesFixtureBackedDownloader(t *testing.T) {
	downloader := &fakeStimaDownloader{
		body: `<SHOP>
  <SHOPITEM>
    <PRODUCTNAME>SakyPaky Hruška červená</PRODUCTNAME>
    <DESCRIPTION>Popis vaku.</DESCRIPTION>
    <IMGURL>https://example.test/red.jpg</IMGURL>
    <PRICE_VAT>1790.00</PRICE_VAT>
    <DELIVERY_DATE>0</DELIVERY_DATE>
    <CATEGORYTEXT>Nábytek | Obývací pokoj | Křesla a taburety | Sedací vaky</CATEGORYTEXT>
    <CODE>HRUSKA-RED</CODE>
    <ITEMGROUP_ID>G-2</ITEMGROUP_ID>
    <EAN>111</EAN>
    <PARAM><PARAM_NAME>Barva</PARAM_NAME><VAL>červená</VAL></PARAM>
  </SHOPITEM>
</SHOP>`,
	}
	generator := Sakypaky{
		Downloader: downloader,
		SourceURL:  "https://example.test/sakypaky.xml",
	}

	var output bytes.Buffer
	result, err := generator.Generate(context.Background(), &output)
	if err != nil {
		t.Fatalf("generate Sakypaky: %v", err)
	}
	if downloader.lastURL != "https://example.test/sakypaky.xml" {
		t.Fatalf("expected custom source URL, got %q", downloader.lastURL)
	}
	if result.ItemsProcessed != 1 || result.ItemsSkipped != 0 {
		t.Fatalf("unexpected result: %#v", result)
	}

	parsed := parseGeneratedDrevocal(t, output.Bytes())
	if len(parsed.Items) != 1 || parsed.Items[0].ExternalID != "SAKYPAKY-G-2" {
		t.Fatalf("unexpected generated items: %#v", parsed.Items)
	}
	variant := parsed.Items[0].Variants[0]
	if variant.Code != "HRUSKA-RED" || variant.EAN != "111" || variant.PriceVAT != "1790" || variant.Currency != "CZK" {
		t.Fatalf("unexpected generated variant: %#v", variant)
	}
	if len(variant.Parameters) != 1 || variant.Parameters[0].Name != "Barva" || variant.Parameters[0].Value != "červená" {
		t.Fatalf("unexpected variant parameters: %#v", variant.Parameters)
	}
}

package feed

import (
	"bytes"
	"context"
	"encoding/xml"
	"testing"
)

func TestDrevocalTestGenerateLimitsOutputItems(t *testing.T) {
	downloader := &fakeStimaDownloader{
		body: `<SHOP>
  <SHOPITEM><ITEM_ID>1</ITEM_ID><ITEMGROUP_ID>401</ITEMGROUP_ID><PRODUCTNAME>Matrace Milena 195x80x10 Úplet</PRODUCTNAME><PRICE_VAT>2650.00</PRICE_VAT><PARAM><PARAM_NAME>Rozměr</PARAM_NAME><VAL>195x80</VAL></PARAM><PARAM><PARAM_NAME>Výška</PARAM_NAME><VAL>10 cm</VAL></PARAM><PARAM><PARAM_NAME>Potah</PARAM_NAME><VAL>Úplet</VAL></PARAM></SHOPITEM>
  <SHOPITEM><ITEM_ID>2</ITEM_ID><ITEMGROUP_ID>403</ITEMGROUP_ID><PRODUCTNAME>Matrace Hana 195x80x14 Úplet</PRODUCTNAME><PRICE_VAT>3679.00</PRICE_VAT><PARAM><PARAM_NAME>Rozměr</PARAM_NAME><VAL>195x80</VAL></PARAM><PARAM><PARAM_NAME>Výška</PARAM_NAME><VAL>14 cm</VAL></PARAM><PARAM><PARAM_NAME>Potah</PARAM_NAME><VAL>Úplet</VAL></PARAM></SHOPITEM>
</SHOP>`,
	}
	generator := DrevocalTest{
		Downloader:  downloader,
		SourceURL:   "https://example.test/drevocal.xml",
		MaxProducts: 1,
	}

	var output bytes.Buffer
	result, err := generator.Generate(context.Background(), &output)
	if err != nil {
		t.Fatalf("generate Dřevočal test: %v", err)
	}
	if downloader.lastURL != "https://example.test/drevocal.xml" {
		t.Fatalf("expected custom source URL, got %q", downloader.lastURL)
	}
	if result.ItemsProcessed != 1 || result.ItemsSkipped != 0 {
		t.Fatalf("unexpected result: %#v", result)
	}

	parsed := parseGeneratedDrevocal(t, output.Bytes())
	if len(parsed.Items) != 1 {
		t.Fatalf("expected 1 generated item, got %#v", parsed.Items)
	}
	if parsed.Items[0].ExternalID != "DREVOCAL-401" || parsed.Items[0].Name != "Matrace Milena" {
		t.Fatalf("unexpected generated item: %#v", parsed.Items[0])
	}
	if len(parsed.Items[0].Variants) != 1 || parsed.Items[0].Variants[0].Code != "1" {
		t.Fatalf("unexpected generated variants: %#v", parsed.Items[0].Variants)
	}
	if parsed.Items[0].Categories.Default.ID != "1188" {
		t.Fatalf("unexpected default category: %#v", parsed.Items[0].Categories.Default)
	}
}

func TestDrevocalGenerateUsesFixtureBackedDownloader(t *testing.T) {
	downloader := &fakeStimaDownloader{
		body: `<SHOP>
  <SHOPITEM>
    <ITEM_ID>5211112</ITEM_ID>
    <ITEMGROUP_ID>521</ITEMGROUP_ID>
    <PRODUCTNAME>Matrace Eliška 195x80x19 Úplet</PRODUCTNAME>
    <MANUFACTURER>DŘEVOČAL</MANUFACTURER>
    <PRICE_VAT>3588.00</PRICE_VAT>
    <CURRENCY>CZK</CURRENCY>
    <EAN>8596723002176</EAN>
    <DESCRIPTION>Eliška je ideální volbou.</DESCRIPTION>
    <IMGURL>https://www.matrace-drevocal.cz/eliska.jpg</IMGURL>
    <AVAILABILITY>Skladem</AVAILABILITY>
    <GIFT>polštář Lukáš</GIFT>
    <PARAM><PARAM_NAME>Rozměr</PARAM_NAME><VAL>195x80</VAL></PARAM>
    <PARAM><PARAM_NAME>Výška</PARAM_NAME><VAL>19 cm</VAL></PARAM>
    <PARAM><PARAM_NAME>Potah</PARAM_NAME><VAL>Úplet</VAL></PARAM>
  </SHOPITEM>
</SHOP>`,
	}
	generator := Drevocal{
		Downloader: downloader,
		SourceURL:  "https://example.test/drevocal.xml",
	}

	var output bytes.Buffer
	result, err := generator.Generate(context.Background(), &output)
	if err != nil {
		t.Fatalf("generate Dřevočal: %v", err)
	}
	if result.ItemsProcessed != 1 || result.ItemsSkipped != 0 {
		t.Fatalf("unexpected result: %#v", result)
	}

	parsed := parseGeneratedDrevocal(t, output.Bytes())
	if len(parsed.Items) != 1 || parsed.Items[0].ExternalID != "DREVOCAL-521" {
		t.Fatalf("unexpected generated items: %#v", parsed.Items)
	}
	if parsed.Items[0].Description != "" {
		t.Fatalf("expected generated Dřevočal output without DESCRIPTION, got %q", parsed.Items[0].Description)
	}
	variant := parsed.Items[0].Variants[0]
	if variant.Code != "5211112" || variant.EAN != "8596723002176" || variant.PriceVAT != "3588" || variant.Currency != "CZK" || variant.Availability != "Skladem" {
		t.Fatalf("unexpected generated variant: %#v", variant)
	}
	if len(variant.Parameters) != 3 || variant.Parameters[0].Name != "Rozměr" || variant.Parameters[0].Value != "195x80" {
		t.Fatalf("unexpected variant parameters: %#v", variant.Parameters)
	}
	if len(parsed.Items[0].InformationParameters) != 1 || parsed.Items[0].InformationParameters[0].Name != "Dárek" || parsed.Items[0].InformationParameters[0].Values[0] != "polštář Lukáš" {
		t.Fatalf("unexpected information parameters: %#v", parsed.Items[0].InformationParameters)
	}
}

func parseGeneratedDrevocal(t *testing.T, data []byte) generatedDrevocalShop {
	t.Helper()
	var parsed generatedDrevocalShop
	if err := xml.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("generated output is not XML: %v", err)
	}
	return parsed
}

type generatedDrevocalShop struct {
	Items []generatedDrevocalItem `xml:"SHOPITEM"`
}

type generatedDrevocalItem struct {
	ExternalID            string                              `xml:"EXTERNAL_ID"`
	Name                  string                              `xml:"NAME"`
	Description           string                              `xml:"DESCRIPTION"`
	Categories            generatedDrevocalCategories         `xml:"CATEGORIES"`
	InformationParameters []generatedDrevocalInformationParam `xml:"INFORMATION_PARAMETERS>INFORMATION_PARAMETER"`
	Variants              []generatedDrevocalVariant          `xml:"VARIANTS>VARIANT"`
}

type generatedDrevocalCategories struct {
	Default generatedDrevocalCategory `xml:"DEFAULT_CATEGORY"`
}

type generatedDrevocalCategory struct {
	ID   string `xml:"id,attr"`
	Path string `xml:",chardata"`
}

type generatedDrevocalVariant struct {
	Code         string                       `xml:"CODE"`
	EAN          string                       `xml:"EAN"`
	PriceVAT     string                       `xml:"PRICE_VAT"`
	Currency     string                       `xml:"CURRENCY"`
	Availability string                       `xml:"AVAILABILITY"`
	Parameters   []generatedDrevocalParameter `xml:"PARAMETERS>PARAMETER"`
}

type generatedDrevocalParameter struct {
	Name  string `xml:"NAME"`
	Value string `xml:"VALUE"`
}

type generatedDrevocalInformationParam struct {
	Name   string   `xml:"NAME"`
	Values []string `xml:"VALUE"`
}

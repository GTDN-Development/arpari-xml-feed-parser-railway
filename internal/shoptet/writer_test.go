package shoptet

import (
	"bytes"
	"encoding/xml"
	"strings"
	"testing"
)

func TestWriteSimpleProduct(t *testing.T) {
	var output bytes.Buffer

	err := Write(&output, Feed{
		Items: []Item{
			{
				Code:             "HELLO-001",
				Name:             "Hello world product",
				ShortDescription: "Short product text",
				Description:      "Long product text",
				PriceVAT:         "123.45",
				Stock:            "7",
				Availability:     "Skladem",
				EAN:              "8590000000001",
				Categories: []Category{
					{ID: "902", Path: "ŽIDLE"},
					{ID: "905", Path: "ŽIDLE > DŘEVĚNÉ ŽIDLE"},
				},
				DefaultCategory: &Category{ID: "905", Path: "ŽIDLE > DŘEVĚNÉ ŽIDLE"},
				Images: []Image{
					{URL: "https://www.stima.cz/userfiles/xml/pictures/simple.jpg"},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("write simple product: %v", err)
	}

	var parsed shopXML
	if err := xml.Unmarshal(output.Bytes(), &parsed); err != nil {
		t.Fatalf("output is not well-formed XML: %v", err)
	}
	if len(parsed.Items) != 1 {
		t.Fatalf("expected 1 SHOPITEM, got %d", len(parsed.Items))
	}

	item := parsed.Items[0]
	if item.Code != "HELLO-001" {
		t.Fatalf("expected item code HELLO-001, got %q", item.Code)
	}
	if item.Name != "Hello world product" {
		t.Fatalf("expected item name, got %q", item.Name)
	}
	if item.ShortDescription != "Short product text" {
		t.Fatalf("expected item short description, got %q", item.ShortDescription)
	}
	if item.Description != "Long product text" {
		t.Fatalf("expected item description, got %q", item.Description)
	}
	if item.PriceVAT != "123.45" {
		t.Fatalf("expected item PRICE_VAT, got %q", item.PriceVAT)
	}
	if item.Stock == nil || item.Stock.Value != "7" {
		t.Fatalf("expected item STOCK, got %#v", item.Stock)
	}
	if !strings.Contains(output.String(), "<AMOUNT>7</AMOUNT>") {
		t.Fatalf("expected structured stock AMOUNT, got:\n%s", output.String())
	}
	if item.Availability != "Skladem" {
		t.Fatalf("expected item AVAILABILITY, got %q", item.Availability)
	}
	if item.EAN != "8590000000001" {
		t.Fatalf("expected item EAN, got %q", item.EAN)
	}
	if item.Categories == nil || len(item.Categories.Items) != 2 {
		t.Fatalf("expected item categories, got %#v", item.Categories)
	}
	if item.Categories.Items[1].ID != "905" || item.Categories.Items[1].Path != "ŽIDLE > DŘEVĚNÉ ŽIDLE" {
		t.Fatalf("expected mapped category, got %#v", item.Categories.Items[1])
	}
	if item.Categories.Default == nil || item.Categories.Default.ID != "905" || item.Categories.Default.Path != "ŽIDLE > DŘEVĚNÉ ŽIDLE" {
		t.Fatalf("expected default category, got %#v", item.Categories.Default)
	}
	if item.Images == nil || len(item.Images.Items) != 1 || item.Images.Items[0].URL != "https://www.stima.cz/userfiles/xml/pictures/simple.jpg" {
		t.Fatalf("expected item images, got %#v", item.Images)
	}
}

func TestWriteProductWithVariants(t *testing.T) {
	var output bytes.Buffer

	err := Write(&output, Feed{
		Items: []Item{
			{
				Code: "CHAIR-001",
				Name: "Chair",
				Variants: []Variant{
					{
						Code:         "CHAIR-001-OAK",
						EAN:          "8590000000002",
						PriceVAT:     "1000.00",
						Warehouses:   []Warehouse{{Name: "HLAVNÍ SKLAD", Value: "3.000"}},
						Availability: "Skladem",
						Parameters: []Parameter{
							{Name: "KOSTRA", Value: "dub"},
							{Name: "Sedák", Value: "raven 15 šedá"},
						},
					},
					{
						Code:         "CHAIR-001-BEECH",
						PriceVAT:     "1100.00",
						Stock:        "0",
						Availability: "Na zakázku",
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("write product with variants: %v", err)
	}

	var parsed shopXML
	if err := xml.Unmarshal(output.Bytes(), &parsed); err != nil {
		t.Fatalf("output is not well-formed XML: %v", err)
	}
	if len(parsed.Items) != 1 {
		t.Fatalf("expected 1 SHOPITEM, got %d", len(parsed.Items))
	}
	if parsed.Items[0].Variants == nil {
		t.Fatal("expected VARIANTS element")
	}
	if parsed.Items[0].ExternalID != "CHAIR-001" {
		t.Fatalf("expected parent EXTERNAL_ID, got %q", parsed.Items[0].ExternalID)
	}
	if parsed.Items[0].Code != "" {
		t.Fatalf("expected no parent CODE for variant product, got %q", parsed.Items[0].Code)
	}
	if len(parsed.Items[0].Variants.Items) != 2 {
		t.Fatalf("expected 2 VARIANT elements, got %d", len(parsed.Items[0].Variants.Items))
	}

	first := parsed.Items[0].Variants.Items[0]
	if first.Code != "CHAIR-001-OAK" {
		t.Fatalf("expected first variant code, got %q", first.Code)
	}
	if first.Stock == nil || len(first.Stock.Warehouses) != 1 || first.Stock.Warehouses[0].Name != "HLAVNÍ SKLAD" || first.Stock.Warehouses[0].Value != "3.000" {
		t.Fatalf("expected first variant structured stock, got %#v", first.Stock)
	}
	if first.Parameters == nil || len(first.Parameters.Items) != 2 {
		t.Fatalf("expected first variant parameters, got %#v", first.Parameters)
	}
	if first.Parameters.Items[0].Name != "KOSTRA" || first.Parameters.Items[0].Value != "dub" {
		t.Fatalf("expected KOSTRA parameter, got %#v", first.Parameters.Items[0])
	}
}

func TestWriteEscapesXMLText(t *testing.T) {
	var output bytes.Buffer

	err := Write(&output, Feed{
		Items: []Item{
			{
				Code:     "A&B-<1>",
				Name:     `Chair "A&B" <test>`,
				PriceVAT: "123.45",
				Stock:    "7",
			},
		},
	})
	if err != nil {
		t.Fatalf("write escaped XML: %v", err)
	}

	xmlText := output.String()
	if !strings.Contains(xmlText, "A&amp;B-&lt;1&gt;") {
		t.Fatalf("expected escaped code in XML, got:\n%s", xmlText)
	}
	if !strings.Contains(xmlText, "Chair &#34;A&amp;B&#34; &lt;test&gt;") {
		t.Fatalf("expected escaped name in XML, got:\n%s", xmlText)
	}

	var parsed shopXML
	if err := xml.Unmarshal(output.Bytes(), &parsed); err != nil {
		t.Fatalf("output is not well-formed XML: %v", err)
	}
	if parsed.Items[0].Code != "A&B-<1>" {
		t.Fatalf("expected unescaped code after parse, got %q", parsed.Items[0].Code)
	}
}

func TestValidateRejectsEmptyFeed(t *testing.T) {
	err := Validate(Feed{})
	if err == nil {
		t.Fatal("expected empty feed validation error")
	}
	if !strings.Contains(err.Error(), "at least one SHOPITEM") {
		t.Fatalf("expected readable empty feed error, got %q", err.Error())
	}
}

func TestValidateRejectsMissingItemCode(t *testing.T) {
	err := Validate(Feed{Items: []Item{{Name: "Missing code"}}})
	if err == nil {
		t.Fatal("expected missing CODE validation error")
	}
	if !strings.Contains(err.Error(), "CODE is required") {
		t.Fatalf("expected readable CODE error, got %q", err.Error())
	}
}

func TestValidateRejectsTooManyItems(t *testing.T) {
	err := ValidateWithLimits(Feed{
		Items: []Item{
			{Code: "A"},
			{Code: "B"},
		},
	}, Limits{MaxItems: 1})
	if err == nil {
		t.Fatal("expected item limit validation error")
	}
	if !strings.Contains(err.Error(), "limit is 1") {
		t.Fatalf("expected readable item limit error, got %q", err.Error())
	}
}

func TestValidateRejectsTooManyVariants(t *testing.T) {
	err := ValidateWithLimits(Feed{
		Items: []Item{
			{
				Code: "CHAIR-001",
				Variants: []Variant{
					{Code: "CHAIR-001-A"},
					{Code: "CHAIR-001-B"},
				},
			},
		},
	}, Limits{MaxVariantsPerItem: 1})
	if err == nil {
		t.Fatal("expected variant limit validation error")
	}
	if !strings.Contains(err.Error(), "has 2 variants; limit is 1") {
		t.Fatalf("expected readable variant limit error, got %q", err.Error())
	}
}

func TestValidateRejectsMissingVariantCode(t *testing.T) {
	err := Validate(Feed{
		Items: []Item{
			{
				Code:     "CHAIR-001",
				Variants: []Variant{{Stock: "1"}},
			},
		},
	})
	if err == nil {
		t.Fatal("expected missing variant CODE validation error")
	}
	if !strings.Contains(err.Error(), "VARIANT[0] CODE is required") {
		t.Fatalf("expected readable variant CODE error, got %q", err.Error())
	}
}

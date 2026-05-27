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
				Code:         "HELLO-001",
				Name:         "Hello world product",
				PriceVAT:     "123.45",
				Stock:        "7",
				Availability: "Skladem",
				EAN:          "8590000000001",
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
	if item.PriceVAT != "123.45" {
		t.Fatalf("expected item PRICE_VAT, got %q", item.PriceVAT)
	}
	if item.Stock != "7" {
		t.Fatalf("expected item STOCK, got %q", item.Stock)
	}
	if item.Availability != "Skladem" {
		t.Fatalf("expected item AVAILABILITY, got %q", item.Availability)
	}
	if item.EAN != "8590000000001" {
		t.Fatalf("expected item EAN, got %q", item.EAN)
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
						Stock:        "3",
						Availability: "Skladem",
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
	if len(parsed.Items[0].Variants.Items) != 2 {
		t.Fatalf("expected 2 VARIANT elements, got %d", len(parsed.Items[0].Variants.Items))
	}

	first := parsed.Items[0].Variants.Items[0]
	if first.Code != "CHAIR-001-OAK" {
		t.Fatalf("expected first variant code, got %q", first.Code)
	}
	if first.Stock != "3" {
		t.Fatalf("expected first variant stock, got %q", first.Stock)
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

package shoptet

import (
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

const (
	DefaultMaxItems           = 20000
	DefaultMaxVariantsPerItem = 512
)

type Limits struct {
	MaxItems           int
	MaxVariantsPerItem int
}

type Feed struct {
	Items []Item
}

type Item struct {
	Code         string
	Name         string
	PriceVAT     string
	Stock        string
	Availability string
	EAN          string
	Variants     []Variant
}

type Variant struct {
	Code         string
	PriceVAT     string
	Stock        string
	Availability string
	EAN          string
}

func Write(w io.Writer, feed Feed) error {
	return WriteWithLimits(w, feed, Limits{})
}

func WriteWithLimits(w io.Writer, feed Feed, limits Limits) error {
	limits = normalizeLimits(limits)
	if err := ValidateWithLimits(feed, limits); err != nil {
		return err
	}

	if _, err := io.WriteString(w, xml.Header); err != nil {
		return fmt.Errorf("write XML header: %w", err)
	}

	encoder := xml.NewEncoder(w)
	encoder.Indent("", "  ")
	if err := encoder.Encode(toShop(feed)); err != nil {
		return fmt.Errorf("encode Shoptet XML: %w", err)
	}
	if err := encoder.Flush(); err != nil {
		return fmt.Errorf("flush Shoptet XML: %w", err)
	}
	if _, err := io.WriteString(w, "\n"); err != nil {
		return fmt.Errorf("write XML newline: %w", err)
	}

	return nil
}

func Validate(feed Feed) error {
	return ValidateWithLimits(feed, Limits{})
}

func ValidateWithLimits(feed Feed, limits Limits) error {
	limits = normalizeLimits(limits)
	if len(feed.Items) == 0 {
		return fmt.Errorf("Shoptet feed must contain at least one SHOPITEM")
	}
	if len(feed.Items) > limits.MaxItems {
		return fmt.Errorf("Shoptet feed contains %d SHOPITEMs; limit is %d", len(feed.Items), limits.MaxItems)
	}

	for itemIndex, item := range feed.Items {
		if strings.TrimSpace(item.Code) == "" {
			return fmt.Errorf("SHOPITEM[%d] CODE is required", itemIndex)
		}
		if len(item.Variants) > limits.MaxVariantsPerItem {
			return fmt.Errorf("SHOPITEM[%d] %q has %d variants; limit is %d", itemIndex, item.Code, len(item.Variants), limits.MaxVariantsPerItem)
		}
		for variantIndex, variant := range item.Variants {
			if strings.TrimSpace(variant.Code) == "" {
				return fmt.Errorf("SHOPITEM[%d] %q VARIANT[%d] CODE is required", itemIndex, item.Code, variantIndex)
			}
		}
	}

	return nil
}

func normalizeLimits(limits Limits) Limits {
	if limits.MaxItems == 0 {
		limits.MaxItems = DefaultMaxItems
	}
	if limits.MaxVariantsPerItem == 0 {
		limits.MaxVariantsPerItem = DefaultMaxVariantsPerItem
	}
	return limits
}

func toShop(feed Feed) shopXML {
	items := make([]shopItemXML, 0, len(feed.Items))
	for _, item := range feed.Items {
		shopItem := shopItemXML{
			Code:         item.Code,
			Name:         item.Name,
			PriceVAT:     item.PriceVAT,
			Stock:        item.Stock,
			Availability: item.Availability,
			EAN:          item.EAN,
		}
		if len(item.Variants) > 0 {
			variants := make([]shopVariantXML, 0, len(item.Variants))
			for _, variant := range item.Variants {
				variants = append(variants, shopVariantXML{
					Code:         variant.Code,
					PriceVAT:     variant.PriceVAT,
					Stock:        variant.Stock,
					Availability: variant.Availability,
					EAN:          variant.EAN,
				})
			}
			shopItem.Variants = &shopVariantsXML{Items: variants}
		}
		items = append(items, shopItem)
	}
	return shopXML{Items: items}
}

type shopXML struct {
	XMLName xml.Name      `xml:"SHOP"`
	Items   []shopItemXML `xml:"SHOPITEM"`
}

type shopItemXML struct {
	Code         string           `xml:"CODE,omitempty"`
	Name         string           `xml:"NAME,omitempty"`
	PriceVAT     string           `xml:"PRICE_VAT,omitempty"`
	Stock        string           `xml:"STOCK,omitempty"`
	Availability string           `xml:"AVAILABILITY,omitempty"`
	EAN          string           `xml:"EAN,omitempty"`
	Variants     *shopVariantsXML `xml:"VARIANTS,omitempty"`
}

type shopVariantsXML struct {
	Items []shopVariantXML `xml:"VARIANT"`
}

type shopVariantXML struct {
	Code         string `xml:"CODE,omitempty"`
	EAN          string `xml:"EAN,omitempty"`
	PriceVAT     string `xml:"PRICE_VAT,omitempty"`
	Stock        string `xml:"STOCK,omitempty"`
	Availability string `xml:"AVAILABILITY,omitempty"`
}

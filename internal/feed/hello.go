package feed

import (
	"context"
	"encoding/xml"
	"io"
)

type Hello struct{}

type helloShop struct {
	XMLName xml.Name      `xml:"SHOP"`
	Item    helloShopItem `xml:"SHOPITEM"`
}

type helloShopItem struct {
	Code     string `xml:"CODE"`
	Name     string `xml:"NAME"`
	PriceVAT string `xml:"PRICE_VAT"`
	Stock    string `xml:"STOCK"`
}

func (Hello) Name() string {
	return "hello"
}

func (Hello) Filename() string {
	return "hello.xml"
}

func (Hello) Generate(_ context.Context, w io.Writer) error {
	if _, err := io.WriteString(w, xml.Header); err != nil {
		return err
	}

	encoder := xml.NewEncoder(w)
	encoder.Indent("", "  ")

	shop := helloShop{
		Item: helloShopItem{
			Code:     "HELLO-001",
			Name:     "Hello world product",
			PriceVAT: "123.45",
			Stock:    "7",
		},
	}

	if err := encoder.Encode(shop); err != nil {
		return err
	}

	if _, err := io.WriteString(w, "\n"); err != nil {
		return err
	}

	return encoder.Flush()
}

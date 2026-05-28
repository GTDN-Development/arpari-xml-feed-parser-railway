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
	Code             string
	Name             string
	ShortDescription string
	Description      string
	Price            string
	PriceVAT         string
	VAT              string
	Currency         string
	Stock            string
	Warehouses       []Warehouse
	Availability     string
	EAN              string
	Categories       []Category
	DefaultCategory  *Category
	Images           []Image
	Variants         []Variant
}

type Variant struct {
	Code         string
	Price        string
	PriceVAT     string
	VAT          string
	Currency     string
	Stock        string
	Warehouses   []Warehouse
	Availability string
	EAN          string
	ImageRef     string
	Parameters   []Parameter
}

type Warehouse struct {
	Name  string
	Value string
}

type Parameter struct {
	Name  string
	Value string
}

type Category struct {
	ID   string
	Path string
}

type Image struct {
	URL string
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
		if len(item.Variants) == 0 && strings.TrimSpace(item.Code) == "" {
			return fmt.Errorf("SHOPITEM[%d] CODE is required", itemIndex)
		}
		if len(item.Variants) > limits.MaxVariantsPerItem {
			return fmt.Errorf("SHOPITEM[%d] %q has %d variants; limit is %d", itemIndex, item.Code, len(item.Variants), limits.MaxVariantsPerItem)
		}
		for variantIndex, variant := range item.Variants {
			if strings.TrimSpace(variant.Code) == "" {
				return fmt.Errorf("SHOPITEM[%d] %q VARIANT[%d] CODE is required", itemIndex, item.Code, variantIndex)
			}
			for parameterIndex, parameter := range variant.Parameters {
				if strings.TrimSpace(parameter.Name) == "" {
					return fmt.Errorf("SHOPITEM[%d] %q VARIANT[%d] PARAMETER[%d] NAME is required", itemIndex, item.Code, variantIndex, parameterIndex)
				}
			}
		}
		for categoryIndex, category := range item.Categories {
			if strings.TrimSpace(category.Path) == "" {
				return fmt.Errorf("SHOPITEM[%d] %q CATEGORY[%d] path is required", itemIndex, item.Code, categoryIndex)
			}
		}
		if item.DefaultCategory != nil && strings.TrimSpace(item.DefaultCategory.Path) == "" {
			return fmt.Errorf("SHOPITEM[%d] %q DEFAULT_CATEGORY path is required", itemIndex, item.Code)
		}
		for imageIndex, image := range item.Images {
			if strings.TrimSpace(image.URL) == "" {
				return fmt.Errorf("SHOPITEM[%d] %q IMAGE[%d] URL is required", itemIndex, item.Code, imageIndex)
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
			Name:             item.Name,
			ShortDescription: item.ShortDescription,
			Description:      item.Description,
			Categories:       toCategoriesXML(item.Categories, item.DefaultCategory),
			Images:           toImagesXML(item.Images),
		}
		if len(item.Variants) > 0 {
			shopItem.ExternalID = item.Code
			variants := make([]shopVariantXML, 0, len(item.Variants))
			for _, variant := range item.Variants {
				variants = append(variants, shopVariantXML{
					Code:         variant.Code,
					EAN:          variant.EAN,
					Currency:     variant.Currency,
					VAT:          variant.VAT,
					Price:        variant.Price,
					PriceVAT:     variant.PriceVAT,
					Stock:        toStockXML(variant.Stock, variant.Warehouses),
					Availability: variant.Availability,
					ImageRef:     variant.ImageRef,
					Parameters:   toParametersXML(variant.Parameters),
				})
			}
			shopItem.Variants = &shopVariantsXML{Items: variants}
		} else {
			shopItem.Code = item.Code
			shopItem.EAN = item.EAN
			shopItem.Currency = item.Currency
			shopItem.VAT = item.VAT
			shopItem.Price = item.Price
			shopItem.PriceVAT = item.PriceVAT
			shopItem.Stock = toStockXML(item.Stock, item.Warehouses)
			shopItem.Availability = item.Availability
		}
		items = append(items, shopItem)
	}
	return shopXML{Items: items}
}

func toCategoriesXML(categories []Category, defaultCategory *Category) *shopCategoriesXML {
	if len(categories) == 0 && defaultCategory == nil {
		return nil
	}

	items := make([]shopCategoryXML, 0, len(categories))
	for _, category := range categories {
		path := strings.TrimSpace(category.Path)
		if path == "" {
			continue
		}
		items = append(items, shopCategoryXML{
			ID:   strings.TrimSpace(category.ID),
			Path: path,
		})
	}

	var defaultItem *shopCategoryXML
	if defaultCategory != nil {
		path := strings.TrimSpace(defaultCategory.Path)
		if path != "" {
			defaultItem = &shopCategoryXML{
				ID:   strings.TrimSpace(defaultCategory.ID),
				Path: path,
			}
		}
	}

	if len(items) == 0 && defaultItem == nil {
		return nil
	}
	return &shopCategoriesXML{Items: items, Default: defaultItem}
}

func toImagesXML(images []Image) *shopImagesXML {
	if len(images) == 0 {
		return nil
	}

	items := make([]shopImageXML, 0, len(images))
	for _, image := range images {
		url := strings.TrimSpace(image.URL)
		if url == "" {
			continue
		}
		items = append(items, shopImageXML{URL: url})
	}
	if len(items) == 0 {
		return nil
	}
	return &shopImagesXML{Items: items}
}

func toStockXML(stock string, warehouses []Warehouse) *shopStockXML {
	stock = strings.TrimSpace(stock)
	if len(warehouses) == 0 && stock == "" {
		return nil
	}

	result := &shopStockXML{
		Value: stock,
	}
	if len(warehouses) > 0 {
		result.Warehouses = make([]shopWarehouseXML, 0, len(warehouses))
		for _, warehouse := range warehouses {
			name := strings.TrimSpace(warehouse.Name)
			value := strings.TrimSpace(warehouse.Value)
			if name == "" && value == "" {
				continue
			}
			result.Warehouses = append(result.Warehouses, shopWarehouseXML{
				Name:  name,
				Value: value,
			})
		}
	}
	if len(result.Warehouses) == 0 && result.Value == "" {
		return nil
	}
	return result
}

func toParametersXML(parameters []Parameter) *shopParametersXML {
	if len(parameters) == 0 {
		return nil
	}

	items := make([]shopParameterXML, 0, len(parameters))
	for _, parameter := range parameters {
		name := strings.TrimSpace(parameter.Name)
		value := strings.TrimSpace(parameter.Value)
		if name == "" && value == "" {
			continue
		}
		items = append(items, shopParameterXML{
			Name:  name,
			Value: value,
		})
	}
	if len(items) == 0 {
		return nil
	}
	return &shopParametersXML{Items: items}
}

type shopXML struct {
	XMLName xml.Name      `xml:"SHOP"`
	Items   []shopItemXML `xml:"SHOPITEM"`
}

type shopItemXML struct {
	ExternalID       string             `xml:"EXTERNAL_ID,omitempty"`
	Code             string             `xml:"CODE,omitempty"`
	Name             string             `xml:"NAME,omitempty"`
	ShortDescription string             `xml:"SHORT_DESCRIPTION,omitempty"`
	Description      string             `xml:"DESCRIPTION,omitempty"`
	EAN              string             `xml:"EAN,omitempty"`
	Currency         string             `xml:"CURRENCY,omitempty"`
	VAT              string             `xml:"VAT,omitempty"`
	Price            string             `xml:"PRICE,omitempty"`
	PriceVAT         string             `xml:"PRICE_VAT,omitempty"`
	Stock            *shopStockXML      `xml:"STOCK,omitempty"`
	Availability     string             `xml:"AVAILABILITY,omitempty"`
	Categories       *shopCategoriesXML `xml:"CATEGORIES,omitempty"`
	Images           *shopImagesXML     `xml:"IMAGES,omitempty"`
	Variants         *shopVariantsXML   `xml:"VARIANTS,omitempty"`
}

type shopCategoriesXML struct {
	Items   []shopCategoryXML `xml:"CATEGORY"`
	Default *shopCategoryXML  `xml:"DEFAULT_CATEGORY,omitempty"`
}

type shopCategoryXML struct {
	ID   string `xml:"id,attr,omitempty"`
	Path string `xml:",chardata"`
}

type shopImagesXML struct {
	Items []shopImageXML `xml:"IMAGE"`
}

type shopImageXML struct {
	URL string `xml:",chardata"`
}

type shopVariantsXML struct {
	Items []shopVariantXML `xml:"VARIANT"`
}

type shopVariantXML struct {
	Code         string             `xml:"CODE,omitempty"`
	EAN          string             `xml:"EAN,omitempty"`
	Currency     string             `xml:"CURRENCY,omitempty"`
	VAT          string             `xml:"VAT,omitempty"`
	Price        string             `xml:"PRICE,omitempty"`
	PriceVAT     string             `xml:"PRICE_VAT,omitempty"`
	Stock        *shopStockXML      `xml:"STOCK,omitempty"`
	Availability string             `xml:"AVAILABILITY,omitempty"`
	ImageRef     string             `xml:"IMAGE_REF,omitempty"`
	Parameters   *shopParametersXML `xml:"PARAMETERS,omitempty"`
}

type shopStockXML struct {
	Value      string
	Warehouses []shopWarehouseXML
}

func (stock shopStockXML) MarshalXML(encoder *xml.Encoder, start xml.StartElement) error {
	if err := encoder.EncodeToken(start); err != nil {
		return err
	}
	if len(stock.Warehouses) > 0 {
		if err := encoder.EncodeElement(shopWarehousesXML{Items: stock.Warehouses}, xml.StartElement{Name: xml.Name{Local: "WAREHOUSES"}}); err != nil {
			return err
		}
	} else if stock.Value != "" {
		if err := encoder.EncodeElement(stock.Value, xml.StartElement{Name: xml.Name{Local: "AMOUNT"}}); err != nil {
			return err
		}
	}
	return encoder.EncodeToken(start.End())
}

func (stock *shopStockXML) UnmarshalXML(decoder *xml.Decoder, start xml.StartElement) error {
	var raw struct {
		Value      string             `xml:",chardata"`
		Amount     string             `xml:"AMOUNT"`
		Warehouses []shopWarehouseXML `xml:"WAREHOUSES>WAREHOUSE"`
	}
	if err := decoder.DecodeElement(&raw, &start); err != nil {
		return err
	}
	stock.Value = strings.TrimSpace(raw.Amount)
	if stock.Value == "" {
		stock.Value = strings.TrimSpace(raw.Value)
	}
	stock.Warehouses = raw.Warehouses
	return nil
}

type shopWarehousesXML struct {
	Items []shopWarehouseXML `xml:"WAREHOUSE"`
}

type shopWarehouseXML struct {
	Name  string `xml:"NAME,omitempty"`
	Value string `xml:"VALUE,omitempty"`
}

type shopParametersXML struct {
	Items []shopParameterXML `xml:"PARAMETER"`
}

type shopParameterXML struct {
	Name  string `xml:"NAME,omitempty"`
	Value string `xml:"VALUE,omitempty"`
}

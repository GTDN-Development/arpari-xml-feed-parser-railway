package sego

import (
	_ "embed"
	"encoding/csv"
	"strings"
)

//go:embed excluded_products.csv
var excludedProductsCSV string

var defaultExcludedProductCodes = mustParseExcludedProductCodes(excludedProductsCSV)

func mustParseExcludedProductCodes(data string) map[string]struct{} {
	reader := csv.NewReader(strings.NewReader(data))
	reader.FieldsPerRecord = -1

	records, err := reader.ReadAll()
	if err != nil {
		panic("parse SEGO excluded products: " + err.Error())
	}

	result := make(map[string]struct{}, len(records))
	for index, record := range records {
		if index == 0 || len(record) == 0 {
			continue
		}
		code := strings.TrimSpace(record[0])
		if code == "" {
			continue
		}
		result[code] = struct{}{}
	}
	return result
}

func excludedProductCodes(options ProductsOptions) map[string]struct{} {
	if options.ExcludedCodes != nil {
		return options.ExcludedCodes
	}
	return defaultExcludedProductCodes
}

func isProductCodeExcluded(code string, excludedCodes map[string]struct{}) bool {
	if len(excludedCodes) == 0 {
		return false
	}
	_, ok := excludedCodes[strings.TrimSpace(code)]
	return ok
}

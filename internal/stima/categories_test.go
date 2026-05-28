package stima

import (
	"testing"

	"github.com/fanda/arpari-xml-feed-parser-railway/internal/shoptet"
)

func TestTransformCategoriesMapsOnlyKnownShoptetCategories(t *testing.T) {
	categories, defaultCategory := transformCategories(sourceCategories{
		Items: []string{
			"Katalog > Katalog 2026",
			"Katalog > Židle",
			"Katalog > Židle > Dřevěné židle",
			"Katalog > Restaurační židle > Stále skladem",
		},
	}, "Židle Test")

	if len(categories) != 2 {
		t.Fatalf("expected 2 mapped categories, got %#v", categories)
	}
	if categories[0] != (shoptet.Category{ID: "902", Path: "ŽIDLE"}) {
		t.Fatalf("unexpected first category: %#v", categories[0])
	}
	if categories[1] != (shoptet.Category{ID: "905", Path: "ŽIDLE > DŘEVĚNÉ ŽIDLE"}) {
		t.Fatalf("unexpected second category: %#v", categories[1])
	}
	if defaultCategory == nil || *defaultCategory != (shoptet.Category{ID: "905", Path: "ŽIDLE > DŘEVĚNÉ ŽIDLE"}) {
		t.Fatalf("unexpected default category: %#v", defaultCategory)
	}
}

func TestTransformCategoriesFallsBackByProductName(t *testing.T) {
	categories, defaultCategory := transformCategories(sourceCategories{}, "Stůl Test")
	if len(categories) != 1 || categories[0] != (shoptet.Category{ID: "971", Path: "STOLY"}) {
		t.Fatalf("unexpected fallback categories: %#v", categories)
	}
	if defaultCategory == nil || *defaultCategory != categories[0] {
		t.Fatalf("unexpected default category: %#v", defaultCategory)
	}
}

func TestTransformCategoriesReturnsEmptyForUnknownProduct(t *testing.T) {
	categories, defaultCategory := transformCategories(sourceCategories{}, "manipulační poplatek")
	if len(categories) != 0 || defaultCategory != nil {
		t.Fatalf("expected no categories, got %#v / %#v", categories, defaultCategory)
	}
}

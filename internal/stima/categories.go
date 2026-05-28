package stima

import (
	"strings"

	"github.com/fanda/arpari-xml-feed-parser-railway/internal/shoptet"
)

var sourceCategoryMap = map[string]shoptet.Category{
	"Katalog > Židle":                      {ID: "902", Path: "ŽIDLE"},
	"Katalog > Židle > Dřevěné židle":      {ID: "905", Path: "ŽIDLE > DŘEVĚNÉ ŽIDLE"},
	"Katalog > Restaurační židle":          {ID: "1128", Path: "ŽIDLE > RESTAURAČNÍ ŽIDLE"},
	"Katalog > Židle > Plastové židle":     {ID: "911", Path: "ŽIDLE > PLASTOVÉ ŽIDLE"},
	"Katalog > Židle > Barové židle":       {ID: "1143", Path: "ŽIDLE > BAROVÉ ŽIDLE"},
	"Katalog > Židle > Kovové židle":       {ID: "908", Path: "ŽIDLE > KOVOVÉ ŽIDLE"},
	"Katalog > Židle > Taburety, stoličky": {ID: "1155", Path: "ŽIDLE > TABURETY"},
	"Katalog > Stoly":                      {ID: "971", Path: "STOLY"},
	"Katalog > Stoly > Plastové stoly":     {ID: "1251", Path: "STOLY > PLASTOVÉ STOLY"},
	"Katalog > Stoly > Konferenční stoly":  {ID: "1263", Path: "STOLY > KONFERENČNÍ STOLY"},
	"Katalog > Stoly > Stolové podnože":    {ID: "1266", Path: "STOLY > STOLOVÉ PODNOŽE"},
}

func transformCategories(source sourceCategories, productName string) ([]shoptet.Category, *shoptet.Category) {
	categories := make([]shoptet.Category, 0, len(source.Items))
	seen := make(map[string]struct{}, len(source.Items))

	for _, sourcePath := range source.Items {
		category, ok := sourceCategoryMap[strings.TrimSpace(sourcePath)]
		if !ok {
			continue
		}
		if _, ok := seen[category.ID]; ok {
			continue
		}
		seen[category.ID] = struct{}{}
		categories = append(categories, category)
	}

	if len(categories) == 0 {
		if fallback, ok := fallbackCategory(productName); ok {
			categories = append(categories, fallback)
		}
	}

	if len(categories) == 0 {
		return nil, nil
	}

	defaultCategory := chooseDefaultCategory(categories)
	return categories, &defaultCategory
}

func chooseDefaultCategory(categories []shoptet.Category) shoptet.Category {
	selected := categories[0]
	selectedDepth := categoryDepth(selected)
	for _, category := range categories[1:] {
		depth := categoryDepth(category)
		if depth > selectedDepth {
			selected = category
			selectedDepth = depth
		}
	}
	return selected
}

func categoryDepth(category shoptet.Category) int {
	if strings.TrimSpace(category.Path) == "" {
		return 0
	}
	return strings.Count(category.Path, ">") + 1
}

func fallbackCategory(productName string) (shoptet.Category, bool) {
	name := strings.ToLower(strings.TrimSpace(productName))
	switch {
	case strings.HasPrefix(name, "stůl"), strings.HasPrefix(name, "stol"):
		return sourceCategoryMap["Katalog > Stoly"], true
	case strings.HasPrefix(name, "židle"), strings.HasPrefix(name, "křesílko"), strings.HasPrefix(name, "lavice"):
		return sourceCategoryMap["Katalog > Židle"], true
	default:
		return shoptet.Category{}, false
	}
}

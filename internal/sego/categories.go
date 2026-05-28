package sego

import (
	"strings"

	"github.com/fanda/arpari-xml-feed-parser-railway/internal/shoptet"
)

var segoCategoryMap = map[string]shoptet.Category{
	"office":             {ID: "881", Path: "KANCELÁŘSKÉ ŽIDLE A KŘESLA"},
	"office-mesh":        {ID: "1284", Path: "KANCELÁŘSKÉ ŽIDLE A KŘESLA > SÍŤOVANÉ KANCELÁŘSKÉ ŽIDLE"},
	"office-upholstered": {ID: "1275", Path: "KANCELÁŘSKÉ ŽIDLE A KŘESLA > ČALOUNĚNÁ KANCELÁŘSKÁ KŘESLA"},
	"office-child":       {ID: "1125", Path: "KANCELÁŘSKÉ ŽIDLE A KŘESLA > DĚTSKÉ ŽIDLE"},
	"office-work":        {ID: "1140", Path: "KANCELÁŘSKÉ ŽIDLE A KŘESLA > PRACOVNÍ A PRŮMYSLOVÉ ŽIDLE"},
	"office-parts":       {ID: "896", Path: "KANCELÁŘSKÉ ŽIDLE A KŘESLA > NÁHRADNÍ DÍLY A PODLOŽKY"},
	"conference":         {ID: "1146", Path: "ŽIDLE > KONFERENČNÍ ŽIDLE"},
	"lab":                {ID: "1233", Path: "LABORATORNÍ ŽIDLE A LAVICE > LABORATORNÍ ŽIDLE"},
	"medical":            {ID: "1230", Path: "LABORATORNÍ ŽIDLE A LAVICE > ZDRAVOTNÍ ŽIDLE"},
}

func targetCategory(source sourceItem) shoptet.Category {
	baseName := normalizeCategoryText(baseProductName(source.ProductName))
	description := normalizeCategoryText(normalizeDescription(source.Description))
	text := strings.TrimSpace(baseName + " " + description)

	switch {
	case containsAny(baseName, "kolečko", "kluzák", "kříž", "píst", "podnož", "mechanika"):
		return segoCategoryMap["office-parts"]
	case containsAny(text, "dětsk", "malé neposedy", "od 3 let") || containsAny(baseName, "junior", "kinder"):
		return segoCategoryMap["office-child"]
	case containsAny(text, "jednací", "konferenční", "stohovateln", "veřejných prostor"):
		return segoCategoryMap["conference"]
	case containsAny(baseName, "medical", "rescuer", "nursy", "aid", "care", "area", "pill", "drop", "sanit", "support"):
		return segoCategoryMap["medical"]
	case containsAny(text, "zdravotnick", "zdravotn", "lékař", "ordinac"):
		return segoCategoryMap["medical"]
	case containsAny(text, "laborator", "dílensk"):
		return segoCategoryMap["lab"]
	case strings.EqualFold(source.ParameterValue("Čalounění opěráku"), "Síťovaný"):
		return segoCategoryMap["office-mesh"]
	case strings.EqualFold(source.ParameterValue("Čalounění opěráku"), "Látkový"):
		return segoCategoryMap["office-upholstered"]
	case containsAny(text, "křeslo"):
		return segoCategoryMap["office-upholstered"]
	case containsAny(text, "pracovní", "průmyslov"):
		return segoCategoryMap["office-work"]
	default:
		return segoCategoryMap["office"]
	}
}

func baseProductName(productName string) string {
	base, _, _ := strings.Cut(strings.TrimSpace(productName), "|")
	return strings.TrimSpace(base)
}

func normalizeCategoryText(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, "\u00a0", " ")
	return strings.Join(strings.Fields(value), " ")
}

func containsAny(value string, needles ...string) bool {
	for _, needle := range needles {
		if strings.Contains(value, needle) {
			return true
		}
	}
	return false
}

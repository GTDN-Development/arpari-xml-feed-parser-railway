package stima

import "strings"

type FabricWhitelist map[string][]string

func DefaultFabricWhitelist() FabricWhitelist {
	return FabricWhitelist{
		"ART13626-K001-L244": {"borg"},
		"ART13674-K009-L010": {"diosa", "now or never"},
		"ART13678-K009-L010": {"lux"},
		"ART12714-K001-L017": {"lux"},
		"ART12713-K009-L017": {"lux"},
		"ART05978-01-109":    {"lux"},
		"ART12939-K171-L010": {"lux"},
		"ART05585-09-109":    {"lux"},
		"ART10436-K221-L174": {"lux"},
		"ART12455-K005-L010": {"lux"},
		"ART13780-K009-L113": {"lux"},
		"ART13571-K009-L113": {"lux"},
		"ART13751-K171-L010": {"lux"},
		"ART05971-K001-L113": {"lux"},
		"ART01728-K001-L174": {"lux"},
		"ART08728-K001-L113": {"boss"},
		"ART13726-K001-L010": {"boss"},
		"ART11240-K009-L113": {"boss"},
		"ART13610-K001-L113": {"raven"},
		"ART00961-10-109":    {"lux"},
		"ART13770-K009-L010": {"lux"},
		"ART06011-K007-L010": {"lux", "tristan"},
		"ART06514-K001-L113": {"boss"},
		"ART13725-K001-L010": {"boss"},
		"ART13612-K001-L113": {"borg"},
		"ART13445-K001-L174": {"lux"},
		"ART06556-K009-L113": {"lux", "boss"},
		"ART05363-10-109":    {"lux"},
		"ART13588-K001-L113": {"boss"},
		"ART13312-K001-L118": {"borg"},
		"ART13314-K001-L113": {"borg"},
		"ART13809-K001-L113": {"borg"},
		"ART13639-K009-L113": {"borg", "raven"},
		"ART13812-K171-L113": {"borg"},
		"ART13813-K171-L113": {"borg"},
		"ART13814-K171-L113": {"borg"},
		"ART13779-K009-L113": {"borg"},
		"ART13625-K001-L244": {"borg"},
		"ART13739-K171-L244": {"borg"},
		"ART13737-K001-L244": {"borg"},
		"ART13455-K001-L011": {"lux"},
		"ART11798-K001-L113": {"boss"},
		"ART13681-K009-L010": {"lux"},
		"ART13627-K001-L113": {"raven"},
		"ART13472-K171-L113": {"borg"},
		"ART13470-K009-L113": {"borg"},
		"ART05980-09-109":    {"lux"},
		"ART01081-K001-L010": {"lux"},
		"ART12987-K221-L010": {"lux"},
		"ART13622-K001-L010": {"borg"},
		"ART13623-K001-L010": {"borg"},
		"ART05081-01-109":    {"lux"},
		"ART13690-K001-L010": {"lux"},
		"ART12708-K009-L010": {"velvet"},
		"ART12696-K001-L010": {"velvet"},
		"ART06473-K001-L113": {"boss"},
		"ART06472-09-289":    {"boss"},
		"ART01624-K001-L010": {"lux"},
		"ART01887-01-109":    {"lux", "tristan"},
		"ART07091-K009-L113": {"velvet"},
		"ART07014-K001-L153": {"boss"},
	}
}

func (whitelist FabricWhitelist) ProductFabrics(source sourceShopItem) []string {
	if len(whitelist) == 0 {
		return nil
	}
	for _, variant := range source.Variants {
		if fabrics, exists := whitelist[normalizeFabricKey(variant.Code)]; exists {
			return fabrics
		}
	}
	return nil
}

func AllowsFabric(fabrics []string, parameters []sourceParameter) bool {
	if len(fabrics) == 0 {
		return true
	}
	sedak := normalizeFabricValue(parameterValue(parameters, "Sedák"))
	for _, prefix := range fabrics {
		if fabricValueHasPrefix(sedak, prefix) {
			return true
		}
	}
	return false
}

func parameterValue(parameters []sourceParameter, name string) string {
	for _, parameter := range parameters {
		if strings.EqualFold(strings.TrimSpace(parameter.Name), name) {
			return strings.TrimSpace(parameter.Value)
		}
	}
	return ""
}

func normalizeFabricKey(value string) string {
	return strings.ToUpper(strings.TrimSpace(value))
}

func normalizeFabricValue(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func fabricValueHasPrefix(value, prefix string) bool {
	prefix = normalizeFabricValue(prefix)
	if prefix == "" {
		return false
	}
	return value == prefix || strings.HasPrefix(value, prefix+" ") || strings.HasPrefix(value, prefix+"-")
}

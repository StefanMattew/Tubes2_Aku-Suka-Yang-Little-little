package utility

import (
	"backend/src/model"
	"encoding/json"
	"os"
	"strconv"
	"strings"
)

func LoadElementsFromFile(path string) (*model.ElementsDatabase, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	raw := map[string]model.ScrapeElement{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	db := &model.ElementsDatabase{
		Elements: make(map[string]model.Element),
	}

	for name, se := range raw {
		elem := model.ConvertToElement(name, se)
		db.Elements[name] = elem
	}

	return db, nil
}

func SortByTier(db *model.ElementsDatabase) *model.ElementsDatabase {
	orderedTiers := []string{
		"Starting elements",
		"Tier 1 elements",
		"Tier 2 elements",
		"Tier 3 elements",
		"Tier 4 elements",
		"Tier 5 elements",
		"Tier 6 elements",
		"Tier 7 elements",
		"Tier 8 elements",
		"Tier 9 elements",
		"Tier 10 elements",
		"Tier 11 elements",
		"Tier 12 elements",
		"Tier 13 elements",
		"Tier 14 elements",
		"Tier 15 elements",
		"Special element",
	}

	orderedDb := &model.ElementsDatabase{
		Elements: make(map[string]model.Element),
	}

	seen := make(map[string]bool)

	for _, tier := range orderedTiers {
		for name, elem := range db.Elements {
			if elem.Tier == tier && !seen[name] {
				orderedDb.Elements[name] = elem
				seen[name] = true
			}
		}
	}

	// Tambahkan elemen yang tidak memiliki tier atau tidak termasuk dalam daftar tier
	for name, elem := range db.Elements {
		if !seen[name] {
			orderedDb.Elements[name] = elem
		}
	}

	return orderedDb
}

func ParseTier(tierStr string) int {
	if strings.HasPrefix(tierStr, "Tier ") {
		// contoh: "Tier 3 Element"
		parts := strings.Split(tierStr, " ")
		if len(parts) >= 2 {
			if n, err := strconv.Atoi(parts[1]); err == nil {
				return n
			}
		}
	}
	// Tier default rendah untuk Starting Elements atau error
	if tierStr == "Starting elements" {
		return 0
	}
	return 999
}

// func LoadTiers(path string, db *model.ElementsDatabase) error {
// 	data, err := os.ReadFile(path)
// 	if err != nil {
// 	return err
// 	}
// 	var tiers map[string][]string
// 	if err := json.Unmarshal(data, &tiers); err != nil {
// 		return err
// 	}

// 	tierLevel := 0
// 	for _, levelName := range []string{"Starting elements", "Tier 1 elements", "Tier 2 elements", "Tier 3 elements", "Special element"} {
// 		elements := tiers[levelName]
// 		for _, name := range elements {
// 			if e, ok := db.Elements[name]; ok {
// 				e.Tier = tierLevel
// 				db.Elements[name] = e
// 			}
// 		}
// 		tierLevel++
// 	}

// 	return nil
// 	}

// // ReloadElements from default file path
// func ReloadElements() error {
// 	return LoadElements(DEFAULT_ELEMENTS_PATH)
// }

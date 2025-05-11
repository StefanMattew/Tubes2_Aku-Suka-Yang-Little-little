// package model

// type Element struct {
// 	Combos   [][2]string `json:"combos"`
// 	ImageURL string      `json:"image_url"`
// }

// type SearchRequest struct {
// 	StartElements []string `json:"startElements"`
// 	Target        string   `json:"target"`
// 	Method        string   `json:"method"`    // BFS / DFS / BIDIR
// 	Mode          string   `json:"mode"`      // single / multiple
// 	MaxRecipes    int      `json:"maxRecipe"` // untuk multiple
// }

// type SearchResult struct {
// 	Recipes      [][]string `json:"recipes"`
// 	ElapsedTime  int64      `json:"elapsedTime"`  // dalam ms
// 	VisitedNodes int        `json:"visitedNodes"` // jumlah node yang dikunjungi
// }

package model

type Recipe struct {
	Element1 string `json:"element1"`
	Element2 string `json:"element2"`
}

type Element struct {
	ID      string   `json:"id"`             // "Brick"
	Name    string   `json:"name"`           // "Brick" (bisa beda jika ingin nama dengan kapitalisasi)
	IsBasic bool     `json:"isBasic"`        // default false, di-set true untuk air/fire/water/earth
	Recipes []Recipe `json:"recipes"`        // parsed dari [2]string
	Icon    string   `json:"icon,omitempty"` // path ke gambar lokal, dari "image"
	Tier    string   `json:"tier,omitempty"` // 0 untuk basic, 1 untuk hasil kombinasi basic, dst
}

type ElementsDatabase struct {
	Elements map[string]Element `json:"elements"`
}

type ScrapeElement struct {
	Combos [][2]string `json:"combos"`
	Image  string      `json:"image"`
	Tier   string      `json:"tier"`
}

type TreeNode struct {
	Name     string      `json:"name"`
	Recipe   *Recipe     `json:"recipe,omitempty"` // step untuk membentuk node ini
	Children []*TreeNode `json:"children,omitempty"`
}

func ConvertToElement(id string, scraped ScrapeElement) Element {
	recipes := make([]Recipe, 0, len(scraped.Combos))
	for _, combo := range scraped.Combos {
		recipes = append(recipes, Recipe{
			Element1: combo[0],
			Element2: combo[1],
		})
	}

	return Element{
		ID:      id,
		Name:    id,
		IsBasic: id == "air" || id == "water" || id == "fire" || id == "earth",
		Recipes: recipes,
		Icon:    scraped.Image,
		Tier:    scraped.Tier,
	}
}

// func getBasicElement(e ElementsDatabase) Element {
// 	if e.IsBasic= true {
// 		 Element{
// 			ID:      id,
// 			Name:    id,
// 			IsBasic: id == "air" || id == "water" || id == "fire" || id == "earth",
// 			Recipes: recipes,
// 			Icon:    scraped.Image,
// 		}
// }

package main

import (
	// "backend/src/scrapper"
	// "log"
	"backend/src/algorithm"
	"backend/src/model"
	"backend/src/utility"
	"fmt"
	"time"
)

func printBFSResult(result *algorithm.BFSResult, db *model.ElementsDatabase) {
	fmt.Printf("Target: %s\n", result.TargetElement)
	fmt.Printf("Visited nodes: %d\n", result.VisitedNodes)
	fmt.Printf("Paths found: %d\n\n", len(result.Paths))

	for i, path := range result.Paths {
		fmt.Printf("Path %d (%d steps):\n", i+1, len(path))
		for j, recipe := range path {
			resultElement := findRecipeResult(recipe, db)
			fmt.Printf("  Step %d: %s + %s -> %s\n",
				j+1, recipe.Element1, recipe.Element2, resultElement)
		}
		fmt.Println()
	}
}
func findRecipeResult(recipe model.Recipe, db *model.ElementsDatabase) string {
	for elementName, element := range db.Elements {
		for _, r := range element.Recipes {
			if (r.Element1 == recipe.Element1 && r.Element2 == recipe.Element2) ||
				(r.Element1 == recipe.Element2 && r.Element2 == recipe.Element1) {
				return elementName
			}
		}
	}
	return "unknown"
}
func main() {

	// err := scrapper.RunScrapperAndSave()
	// if err != nil {
	// 	log.Fatalf("Scraper failed: %v", err)
	// }
	db, err := utility.LoadElementsFromFile("src/data/elements.json")

	// errs := utility.LoadTiers("src/data/tiers.json", db)
	if err != nil {
		fmt.Println("Gagal load file:", err)
	}
	// sortedDb := utility.SortByTier(db)
	// output, err := json.MarshalIndent(db, "", "  ")
	// if err != nil {
	// 	fmt.Println("Gagal encode JSON:", err)
	// 	return
	// }
	// fmt.Println(string(output))

	target := "Vinegar" // Ganti dengan elemen target yang kamu ingin cari
	maxPaths := 1       // Jumlah maksimal jalur yang dicari

	fmt.Println("\n===== Hasil BFS =====")
	startTime := time.Now()
	bfsResult := algorithm.MultiBFS(db, target, maxPaths, nil)
	duration := time.Since(startTime)
	fmt.Printf("BFS with complete expansion took %s\n", duration)
	fmt.Printf("Found %d paths to %s\n", len(bfsResult.Paths), bfsResult.TargetElement)
	for i, path := range bfsResult.Paths {
		fmt.Printf("\nPath %d (%d steps):\n", i+1, len(path))
		for j, recipe := range path {
			resultElement := findRecipeResult(recipe, db)
			fmt.Printf("  Step %d: %s + %s -> %s\n", j+1, recipe.Element1, recipe.Element2, resultElement)
		}
	}

	//fmt.Println("\n===== Hasil DFS =====")
	//dfsResult := algorithm.MultiDFS(db, target, maxPaths, nil)
	//printResult(dfsResult.Paths, dfsResult.VisitedNodes)

}

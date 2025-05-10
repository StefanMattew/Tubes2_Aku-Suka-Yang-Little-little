package main

import (
	// "backend/src/scrapper"
	// "log"
	"backend/src/algorithm"
	"backend/src/model"
	"backend/src/utility"
	"fmt"
)

func printResult(paths [][]model.Recipe, visited int) {
	fmt.Printf("Visited Nodes: %d\n", visited)
	fmt.Printf("Total Paths Found: %d\n", len(paths))
	for i, path := range paths {
		fmt.Printf("Path %d:\n", i+1)
		for _, recipe := range path {
			fmt.Printf("  %s + %s\n", recipe.Element1, recipe.Element2)
		}
	}
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

	target := "Energy" // Ganti dengan elemen target yang kamu ingin cari
	maxPaths := 100    // Jumlah maksimal jalur yang dicari

	fmt.Println("\n===== Hasil BFS =====")
	bfsResult := algorithm.MultiBFS(db, target, maxPaths, nil)
	printResult(bfsResult.Paths, bfsResult.VisitedNodes)

	fmt.Println("\n===== Hasil DFS =====")
	dfsResult := algorithm.MultiDFS(db, target, maxPaths, nil)
	printResult(dfsResult.Paths, dfsResult.VisitedNodes)

}

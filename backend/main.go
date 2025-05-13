package main

import (
	// "backend/src/scrapper"
	// "log"
	"backend/src/algorithm"
	"backend/src/utility"
	"fmt"
	"time"
)

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

	target := "Grilled cheese" // Ganti dengan elemen target yang kamu ingin cari
	maxPaths := 10     // Jumlah maksimal jalur yang dicari

	fmt.Println("\n===== Hasil BFS =====")
	startTime := time.Now()
	bfsResult := algorithm.MultiBFS(db, target, maxPaths, nil)
	duration := time.Since(startTime)
	fmt.Printf("BFS with complete expansion took %s\n", duration)
	fmt.Printf("Found %d paths to %s\n", len(bfsResult.Paths), bfsResult.TargetElement)
	for i, path := range bfsResult.Paths {
		fmt.Printf("\nPath %d (%d steps):\n", i+1, len(path))
		for j, recipe := range path {
			resultElement := recipe.Result
			fmt.Printf("  Step %d: %s + %s -> %s\n", j+1, recipe.Recipe.Element1, recipe.Recipe.Element2, resultElement)
		}
	}

	//fmt.Println("\n===== Hasil DFS =====")
	//dfsResult := algorithm.MultiDFS(db, target, maxPaths, nil)
	//printResult(dfsResult.Paths, dfsResult.VisitedNodes)

}

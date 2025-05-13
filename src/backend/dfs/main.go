package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"shared/algorithm"
	"shared/model"
	"shared/utility"
	"time"
)

var db *model.ElementsDatabase

func main() {
	db = utility.LoadDatabase()
	/*
		http.HandleFunc("/search", handleSearch)

		log.Println("DFS Server listening at http://localhost:8082")
		log.Fatal(http.ListenAndServe(":8082", nil))
	*/
	target := "Animal" // Ganti dengan elemen target yang kamu ingin cari
	maxPaths := 1      // Jumlah maksimal jalur yang dicari

	fmt.Println("\n===== Hasil DFS =====")
	startTime := time.Now()
	bfsResult := algorithm.DFSDriver(db, target, maxPaths, nil)
	duration := time.Since(startTime)
	fmt.Printf("BFS with complete expansion took %s\n", duration)
	fmt.Printf("Found %d paths to %s\n", len(bfsResult.Paths), bfsResult.TargetElement)
	for i, path := range bfsResult.Paths {
		fmt.Printf("\nPath %d (%d steps):\n", i+1, len(path))
		for j, recipe := range path {
			resultElement := recipe.Result
			fmt.Printf("  Step %d: %s + %s -> %s\n", j+1, recipe.Element1, recipe.Element2, resultElement)
		}
	}
	if db == nil || db.Elements == nil {
		log.Fatal("Database elemen gagal dimuat atau kosong.")
	}
}

func handleSearch(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req model.SearchRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if len(req.StartElements) == 0 {
		req.StartElements = []string{"Air", "Water", "Fire", "Earth"}
	}

	log.Printf("Target: %s, Mode: %s, Max: %d\n", req.Target, req.Mode, req.MaxRecipes)

	start := time.Now()
	var res *algorithm.DFSResult
	if req.Mode == "multiple" {
		res = algorithm.DFSDriver(db, req.Target, req.MaxRecipes, nil)
	} else {
		// single = 1 recipe saja
		res = algorithm.DFSDriver(db, req.Target, 1, nil)
	}
	elapsed := time.Since(start)

	json.NewEncoder(w).Encode(model.SearchResult{
		Recipes:      res.Paths,
		ElapsedTime:  elapsed.Milliseconds(),
		VisitedNodes: res.VisitedNodes,
	})
}

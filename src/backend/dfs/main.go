package main

import (
	"encoding/json"
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

	http.HandleFunc("/search", handleSearch)

	log.Println("DFS Server listening at http://localhost:8082")
	log.Fatal(http.ListenAndServe(":8082", nil))
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
		res = algorithm.MultiDFS(db, req.Target, req.MaxRecipes, nil)
	} else {
		// single = 1 recipe saja
		res = algorithm.MultiDFS(db, req.Target, 1, nil)
	}
	elapsed := time.Since(start)

	json.NewEncoder(w).Encode(model.SearchResult{
		Recipes:      res.Paths,
		ElapsedTime:  elapsed.Milliseconds(),
		VisitedNodes: res.VisitedNodes,
	})
}
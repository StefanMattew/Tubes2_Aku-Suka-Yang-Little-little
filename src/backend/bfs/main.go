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

const IMAGE_DIRECTORY_PATH = "../shared/data/images"

var db *model.ElementsDatabase

func main() {
	db = utility.LoadDatabase()

	http.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir(IMAGE_DIRECTORY_PATH))))
	http.HandleFunc("/element-images", handleElementImages)

	http.HandleFunc("/search", handleSearch)
	http.HandleFunc("/elements", handleElements)

	log.Println("BFS Server listening at http://localhost:8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
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

	start := time.Now()
	var res *algorithm.BFSResult
	if req.Mode == "multiple" {
		res = algorithm.MultiBFS(db, req.Target, req.MaxRecipes, nil)
	} else {
		// single = 1 recipe saja
		res = algorithm.MultiBFS(db, req.Target, 1, nil)
	}
	elapsed := time.Since(start)

	json.NewEncoder(w).Encode(model.SearchResult{
		Recipes:      convertRecipes(res.Paths),
		ElapsedTime:  elapsed.Milliseconds(),
		VisitedNodes: res.VisitedNodes,
	})
}

func handleElements(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var names []string
	for name := range db.Elements {
		names = append(names, name)
	}

	json.NewEncoder(w).Encode(struct {
		Elements []string `json:"elements"`
	}{
		Elements: names,
	})
}

func convertRecipes(input [][]model.Recipe) [][]string {
	var out [][]string
	for _, recipeList := range input {
		var steps []string
		for _, r := range recipeList {
			steps = append(steps, r.Element1+" + "+r.Element2)
		}
		out = append(out, steps)
	}
	return out
}

func handleElementImages(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	imageMap := make(map[string]string)
	for name, el := range db.Elements {
		imageMap[name] = "/images/" + el.Icon
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(imageMap)
}

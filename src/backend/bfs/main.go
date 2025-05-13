package main

import (
	"encoding/json"
	"log"
	"net/http"
	"path/filepath"
	"shared/algorithm"
	"shared/model"
	"shared/utility"
	"sort"
	"strings"
	"time"
)

const DATA_DIRECTORY_PATH = "../shared/data"
const IMAGE_DIRECTORY_SERVE_PATH = "/images/"

var db *model.ElementsDatabase
var tiersData map[string][]string

type ElementInfo struct {
	Name      string `json:"name"`
	ImagePath string `json:"imagePath"` // URL lengkap ke gambar
	Tier      string `json:"tier"`
}

func main() {
	db = utility.LoadDatabase()
	if db == nil || db.Elements == nil {
		log.Fatal("Database elemen gagal dimuat atau kosong.")
	}

	imageDirPath := filepath.Join(DATA_DIRECTORY_PATH, "images")
	http.Handle(IMAGE_DIRECTORY_SERVE_PATH,
		http.StripPrefix(IMAGE_DIRECTORY_SERVE_PATH, http.FileServer(http.Dir(imageDirPath))))

	http.HandleFunc("/search", handleSearch)
	http.HandleFunc("/elements-info", handleElementsInfo)

	log.Println("BFS Server listening at http://localhost:8081")
	log.Fatal(http.ListenAndServe(":8081", corsMiddleware(http.DefaultServeMux)))
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000") // Atau "*" jika lebih fleksibel
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func handleElementsInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var elementsInfoList []ElementInfo
	elementNames := make([]string, 0, len(db.Elements)) // Untuk sorting nama

	for name := range db.Elements {
		elementNames = append(elementNames, name)
	}
	sort.Strings(elementNames) // Urutkan nama elemen agar konsisten

	for _, name := range elementNames {
		el, ok := db.Elements[name]
		if !ok {
			continue // Seharusnya tidak terjadi jika iterasi dari keys db.Elements
		}

		currentTier := el.Tier // Asumsi Tier tidak ada yang null atau kosong

		var imagePath string
		if el.Icon != "" && !strings.HasPrefix(el.Icon, "/") { // Jika Icon adalah nama file saja
			imagePath = "http://localhost:8081" + IMAGE_DIRECTORY_SERVE_PATH + el.Icon
		} else if strings.HasPrefix(el.Icon, "/") { // Jika Icon sudah punya leading slash
			imagePath = "http://localhost:8081" + el.Icon
		} else {
			imagePath = "http://localhost:8081" + IMAGE_DIRECTORY_SERVE_PATH + "placeholder.png" // Fallback
		}

		elementsInfoList = append(elementsInfoList, ElementInfo{
			Name:      name,
			ImagePath: imagePath,
			Tier:      currentTier,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(elementsInfoList) // Kirim array objek ElementInfo
}

func handleSearch(w http.ResponseWriter, r *http.Request) {
	// CORS sudah ditangani oleh middleware
	if r.Method != http.MethodPost { // Method POST untuk search
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
	var res *algorithm.BFSResult
	if req.Mode == "multiple" {
		res = algorithm.Driver(db, req.Target, req.MaxRecipes, nil)
	} else {
		res = algorithm.Driver(db, req.Target, 1, nil) // single = 1 recipe
	}
	elapsed := time.Since(start)

	// Pastikan res.Paths tidak nil sebelum mengirim
	pathsToSend := res.Paths
	if pathsToSend == nil {
		pathsToSend = [][]model.Recipe{} // Kirim array kosong jika nil
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(model.SearchResult{
		Recipes:      pathsToSend,
		ElapsedTime:  elapsed.Milliseconds(),
		VisitedNodes: res.VisitedNodes,
	})
}

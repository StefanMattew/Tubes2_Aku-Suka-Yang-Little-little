// package algorithm

// import (
// 	"fmt"
// 	"shared/model"
// 	"shared/utility"
// )

// type DFSResult struct {
// 	TargetElement string           `json:"target_element"`
// 	Paths         [][]model.Recipe `json:"recipes"`
// 	VisitedNodes  int              `json:"visited_nodes"`
// }

// type DFSNode struct {
// 	Element    string         `json:"element"`
// 	Path       []model.Recipe `json:"path"`
// 	ParentNode *DFSNode
// }

// func DFS(db *model.ElementsDatabase, startElements []string, targetElement string, maxPath int, result chan<- *DFSResult, step chan<- *SearchProgress) {
// 	target, exists := db.Elements[targetElement]
// 	if !exists {
// 		result <- &DFSResult{
// 			TargetElement: targetElement,
// 			Paths:         [][]model.Recipe{},
// 			VisitedNodes:  0,
// 		}
// 		close(result)
// 		return
// 	}
// 	if target.IsBasic {
// 		result <- &DFSResult{
// 			TargetElement: targetElement,
// 			Paths:         [][]model.Recipe{},
// 			VisitedNodes:  1,
// 		}
// 		close(result)
// 		return
// 	}

// 	visitedCombinations := make(map[string]bool)
// 	paths := make([][]model.Recipe, 0)
// 	visitedCount := 0

// 	var dfsRecursive func(current string, path []model.Recipe, depth int)
// 	dfsRecursive = func(current string, path []model.Recipe, depth int) {
// 		if len(paths) >= maxPath {
// 			return // Hentikan jika sudah menemukan cukup banyak jalur.
// 		}

// 		visitedCount++
// 		if step != nil {
// 			step <- &SearchProgress{
// 				CurrentElement: current,
// 				Visited:        visitedCount,
// 				PathsFound:     len(paths),
// 				VisitedNodes:   visitedCombinations,
// 			}
// 		}

// 		if current == targetElement {
// 			newPath := make([]model.Recipe, len(path))
// 			copy(newPath, path)
// 			paths = append(paths, newPath)
// 			return
// 		}

// 		for elementID := range db.Elements {
// 			//Cek kombinasi current dengan element lain.
// 			e1, e2 := current, elementID
// 			if e1 > e2 {
// 				e1, e2 = e2, e1
// 			}
// 			combinationKey := fmt.Sprintf("%s+%s", e1, e2)

// 			if visitedCombinations[combinationKey] {
// 				continue
// 			}

// 			visitedCombinations[combinationKey] = true

// 			for resultElementID, resultElement := range db.Elements {
// 				resultTier := utility.ParseTier(resultElement.Tier)
// 				for _, recipe := range resultElement.Recipes {
// 					if (recipe.Element1 == e1 && recipe.Element2 == e2) || (recipe.Element1 == e2 && recipe.Element2 == e1) {
// 						r1, ok1 := db.Elements[recipe.Element1]
// 						r2, ok2 := db.Elements[recipe.Element2]
// 						if !ok1 || !ok2 {
// 							continue
// 						}
// 						t1 := utility.ParseTier(r1.Tier)
// 						t2 := utility.ParseTier(r2.Tier)
// 						if t1 >= resultTier || t2 >= resultTier {
// 							continue
// 						}
// 						newPath := make([]model.Recipe, len(path)+1)
// 						copy(newPath, path)
// 						newPath[len(path)] = recipe
// 						dfsRecursive(resultElementID, newPath, depth+1)
// 					}
// 				}
// 			}
// 		}
// 	}

// 	for _, basicElement := range startElements {
// 		dfsRecursive(basicElement, []model.Recipe{}, 0)
// 	}

// 	for i := range paths {
// 		paths[i] = expandPath(paths[i], db, startElements)
// 	}

// 	result <- &DFSResult{
// 		TargetElement: targetElement,
// 		Paths:         paths,
// 		VisitedNodes:  visitedCount,
// 	}
// 	close(result)
// }

// // Multithreading DFS
// func MultiDFS(db *model.ElementsDatabase, targetElement string, maxPath int, step chan<- *SearchProgress) *DFSResult {
// 	sortedDb := utility.SortByTier(db)
// 	result := make(chan *DFSResult, 1)
// 	startElement := []string{"Air", "Water", "Fire", "Earth"}

// 	go DFS(sortedDb, startElement, targetElement, maxPath, result, step)

// 	results := <-result

// 	return results
// }

// ======================== FIX ============================
package algorithm

import (
	"shared/model"
	"shared/utility"
	// "log" // Uncomment jika perlu logging
)

type DFSResult struct {
	TargetElement string           `json:"target_element"`
	Paths         [][]model.Recipe `json:"recipes"`
	VisitedNodes  int              `json:"visited_nodes"`
}

// DFSNode tidak secara eksplisit digunakan dalam implementasi rekursif DFS ini,
// tapi path dan elemen saat ini dikelola melalui parameter fungsi.
// type DFSNode struct {
// 	Element    string         `json:"element"`
// 	Path       []model.Recipe `json:"path"`
// 	ParentNode *DFSNode
// }

func DFS(db *model.ElementsDatabase, startElements []string, targetElement string, maxPath int, resultChan chan<- *DFSResult, step chan<- *SearchProgress) { // Ganti nama var result
	target, exists := db.Elements[targetElement]
	if !exists {
		resultChan <- &DFSResult{ // Menggunakan resultChan
			TargetElement: targetElement,
			Paths:         [][]model.Recipe{},
			VisitedNodes:  0,
		}
		close(resultChan) // Menggunakan resultChan
		return
	}
	if target.IsBasic {
		resultChan <- &DFSResult{ // Menggunakan resultChan
			TargetElement: targetElement,
			Paths:         [][]model.Recipe{},
			VisitedNodes:  1,
		}
		close(resultChan) // Menggunakan resultChan
		return
	}

	// Untuk DFS, visitedCombinations lebih cocok untuk melacak *kombinasi bahan* yang sudah dieksplorasi
	// dalam satu cabang pencarian untuk mencegah loop tak terbatas dari kombinasi yang sama.
	// Untuk melacak node yang dikunjungi secara global bisa berbeda, tergantung strategi.
	// visitedPathElements digunakan untuk mencegah siklus dalam path saat ini.

	paths := make([][]model.Recipe, 0)
	visitedCount := 0 // Jumlah total pemanggilan rekursif atau node yang dieksplorasi

	var dfsRecursive func(current string, currentPath []model.Recipe, visitedPathElements map[string]bool)
	dfsRecursive = func(current string, currentPath []model.Recipe, visitedPathElements map[string]bool) {
		if maxPath > 0 && len(paths) >= maxPath {
			return
		}

		visitedCount++
		if step != nil {
			// Untuk DFS, 'visitedNodes' di SearchProgress mungkin lebih baik merefleksikan
			// elemen unik dalam path saat ini atau kedalaman, bukan semua kombinasi global.
			// Ini perlu disesuaikan dengan apa yang ingin ditampilkan di frontend.
			tempVisitedForStep := make(map[string]bool)
			for k, v := range visitedPathElements {
				tempVisitedForStep[k] = v
			}
			step <- &SearchProgress{
				CurrentElement: current,
				Visited:        visitedCount, // Atau kedalaman: len(currentPath)
				PathsFound:     len(paths),
				VisitedNodes:   tempVisitedForStep,
			}
		}

		// Tandai elemen saat ini sebagai dikunjungi dalam path ini
		newVisitedPathElements := make(map[string]bool)
		for k, v := range visitedPathElements {
			newVisitedPathElements[k] = v
		}
		newVisitedPathElements[current] = true

		if current == targetElement {
			// Path ditemukan, salin dan simpan
			// currentPath sudah berisi Recipe dengan Result yang terisi
			pathToAppend := make([]model.Recipe, len(currentPath))
			copy(pathToAppend, currentPath)
			paths = append(paths, pathToAppend)
			return
		}

		// Coba kombinasikan 'current' dengan semua elemen lain yang ada di DB
		// (atau elemen yang sudah ditemukan/dasar, tergantung strategi)
		// Untuk DFS yang mencari resep, kita biasanya mencoba mengkombinasikan elemen saat ini
		// dengan elemen dasar atau elemen lain yang sudah bisa dibuat.

		// Iterasi melalui semua elemen di database sebagai otherElement
		// Ini mungkin sangat tidak efisien untuk DFS. Strategi yang lebih baik adalah
		// mencoba mengkombinasikan 'current' dengan elemen dasar atau elemen yang sudah ada di 'available'
		// Namun, mengikuti struktur kode awal:
		elementsToCombineWith := make([]string, 0, len(db.Elements))
		for elName := range db.Elements {
			elementsToCombineWith = append(elementsToCombineWith, elName)
		}

		for _, otherElement := range elementsToCombineWith {
			if maxPath > 0 && len(paths) >= maxPath {
				return
			}

			e1, e2 := current, otherElement
			if e1 > e2 {
				e1, e2 = e2, e1
			}
			// combinationKey := fmt.Sprintf("%s+%s", e1, e2)
			// visitedCombinations (global map) bisa digunakan di sini jika ingin mencegah
			// eksplorasi ulang kombinasi bahan secara global, tapi hati-hati dengan DFS.

			// Iterasi melalui semua kemungkinan hasil di database
			for resultElementID, resultElementData := range db.Elements {
				if resultElementData.IsBasic {
					continue
				} // Hasil tidak mungkin elemen dasar
				if newVisitedPathElements[resultElementID] { // Jika hasil sudah ada di path saat ini, lewati untuk mencegah siklus langsung
					continue
				}

				resultTier := utility.ParseTier(resultElementData.Tier)

				for _, dbRecipe := range resultElementData.Recipes { // dbRecipe dari database
					if (dbRecipe.Element1 == e1 && dbRecipe.Element2 == e2) || (dbRecipe.Element1 == e2 && dbRecipe.Element2 == e1) {
						// Validasi Tier
						r1Data, ok1 := db.Elements[dbRecipe.Element1]
						r2Data, ok2 := db.Elements[dbRecipe.Element2]
						if !ok1 || !ok2 {
							continue
						}

						t1 := utility.ParseTier(r1Data.Tier)
						t2 := utility.ParseTier(r2Data.Tier)

						// Logika tier yang sama seperti di BFS
						if t1 >= resultTier && !isSpecialCombination(dbRecipe.Element1, resultTier, db) {
							continue
						}
						if t2 >= resultTier && !isSpecialCombination(dbRecipe.Element2, resultTier, db) {
							continue
						}

						// Buat langkah resep dengan Result
						stepRecipe := model.Recipe{
							Element1: dbRecipe.Element1,
							Element2: dbRecipe.Element2,
							Result:   resultElementID,
						}

						newCurrentPath := make([]model.Recipe, len(currentPath)+1)
						copy(newCurrentPath, currentPath)
						newCurrentPath[len(currentPath)] = stepRecipe

						dfsRecursive(resultElementID, newCurrentPath, newVisitedPathElements)
						if maxPath > 0 && len(paths) >= maxPath {
							return
						}
					}
				}
			}
		}
	}

	// Mulai DFS dari setiap elemen dasar
	for _, basicElement := range startElements {
		dfsRecursive(basicElement, []model.Recipe{}, make(map[string]bool))
		if maxPath > 0 && len(paths) >= maxPath {
			break
		}
	}

	expandedPaths := [][]model.Recipe{}
	for _, p := range paths {
		// Pastikan fungsi expandPath dan dependensinya (addDependencies, isBasicElement, dll.)
		// tersedia untuk paket ini, atau pindahkan ke utility jika digunakan bersama.
		// Mengasumsikan expandPath dapat diakses (misalnya, dari paket yang sama atau utilitas).
		// Fungsi expandPath yang ada di bfs.go akan digunakan.
		expandedPaths = append(expandedPaths, expandPath(p, db, startElements))
	}

	resultChan <- &DFSResult{ // Menggunakan resultChan
		TargetElement: targetElement,
		Paths:         expandedPaths,
		VisitedNodes:  visitedCount,
	}
	close(resultChan) // Menggunakan resultChan
}

func MultiDFS(db *model.ElementsDatabase, targetElement string, maxPath int, step chan<- *SearchProgress) *DFSResult {
	// utility.SortByTier(db) // Mungkin tidak diperlukan
	resultChan := make(chan *DFSResult, 1) // Menggunakan resultChan
	startElement := []string{"Air", "Water", "Fire", "Earth"}

	go DFS(db, startElement, targetElement, maxPath, resultChan, step) // Menggunakan resultChan

	results := <-resultChan // Menggunakan resultChan

	return results
}

// Fungsi isSpecialCombination dan canCombineWithSelf (jika diperlukan untuk DFS dengan cara yang sama)
// dapat disalin dari bfs.go atau dipindahkan ke paket utility jika digunakan bersama.
// Untuk saat ini, saya asumsikan fungsi isSpecialCombination dari bfs.go dapat diakses jika berada dalam paket yang sama.

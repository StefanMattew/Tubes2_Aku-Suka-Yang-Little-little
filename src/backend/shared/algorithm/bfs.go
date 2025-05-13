// package algorithm

// import (
// 	"container/list"
// 	"fmt"
// 	"log"
// 	"shared/model"
// 	"shared/utility"
// 	"slices"
// )

// type BFSResult struct {
// 	TargetElement string           `json:"target_element"`
// 	Paths         [][]model.Recipe `json:"recipes"`
// 	VisitedNodes  int              `json:"visited_nodes"`
// }

// type BFSNode struct {
// 	Element    string         `json:"element"`
// 	Path       []model.Recipe `json:"path"`
// 	ParentNode *BFSNode
// }

// type SearchProgress struct {
// 	CurrentElement string          `json:"currentElement"`
// 	Visited        int             `json:"visited"`
// 	PathsFound     int             `json:"pathsFound"`
// 	VisitedNodes   map[string]bool `json:"visitedNodes"`
// }

// func BFS(db *model.ElementsDatabase, startElements []string, targetElement string, maxPath int, result chan<- *BFSResult, step chan<- *SearchProgress) {
// 	target, exists := db.Elements[targetElement]
// 	if !exists {
// 		result <- &BFSResult{
// 			TargetElement: targetElement,
// 			Paths:         [][]model.Recipe{},
// 			VisitedNodes:  0,
// 		}
// 		close(result)
// 		return
// 	}
// 	if target.IsBasic {
// 		result <- &BFSResult{
// 			TargetElement: targetElement,
// 			Paths:         [][]model.Recipe{},
// 			VisitedNodes:  1,
// 		}
// 		close(result)
// 		return
// 	}

// 	visitedCombinations := make(map[string]bool)
// 	discoveredElements := make(map[string]bool)
// 	queue := list.New()

// 	// Masukkan elemen dasar ke queue dan discovered
// 	for _, basic := range startElements {
// 		node := &BFSNode{
// 			Element:    basic,
// 			Path:       []model.Recipe{},
// 			ParentNode: nil,
// 		}
// 		queue.PushBack(node)
// 		discoveredElements[basic] = true
// 	}

// 	paths := [][]model.Recipe{}
// 	visitedCount := 0

// 	for queue.Len() > 0 {
// 		visitedCount++
// 		node := queue.Remove(queue.Front()).(*BFSNode)

// 		if step != nil {
// 			step <- &SearchProgress{
// 				CurrentElement: node.Element,
// 				Visited:        visitedCount,
// 				PathsFound:     len(paths),
// 				VisitedNodes:   discoveredElements,
// 			}
// 		}

// 		if node.Element == targetElement {
// 			paths = append(paths, node.Path)
// 			break
// 		}

// 		// Kombinasikan node ini dengan semua elemen yang sudah ditemukan sebelumnya
// 		for otherElementID := range discoveredElements {
// 			e1, e2 := node.Element, otherElementID
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
// 					// Check if the recipe uses the current element and another discovered element
// 					if (recipe.Element1 == e1 && recipe.Element2 == e2) || (recipe.Element1 == e2 && recipe.Element2 == e1) {
// 						// Cek tier agar tidak lompat ke atas
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

// 						// Kombinasi valid, buat node baru
// 						newPath := make([]model.Recipe, len(node.Path)+1)
// 						copy(newPath, node.Path)
// 						newPath[len(node.Path)] = recipe

// 						newNode := &BFSNode{
// 							Element:    resultElementID,
// 							Path:       newPath,
// 							ParentNode: node,
// 						}
// 						queue.PushBack(newNode)

// 						if !discoveredElements[resultElementID] {
// 							discoveredElements[resultElementID] = true
// 						}
// 					}
// 				}
// 			}
// 		}
// 	}
// 	for i := range paths {
// 		paths[i] = expandPath(paths[i], db, startElements)
// 	}
// 	result <- &BFSResult{
// 		TargetElement: targetElement,
// 		Paths:         paths,
// 		VisitedNodes:  visitedCount,
// 	}
// 	close(result)
// }
// func expandPath(path []model.Recipe, db *model.ElementsDatabase, startElements []string) []model.Recipe {
// 	// Track what we can make
// 	available := make(map[string]bool)

// 	// Mark basic elements as available
// 	for _, elem := range startElements {
// 		available[elem] = true
// 	}

// 	expandedPath := []model.Recipe{}

// 	// Process each recipe in the original path
// 	for _, recipe := range path {
// 		// Track elements being processed to detect cycles
// 		processing := make(map[string]bool)

// 		// Recursively add missing dependencies for Element1
// 		addDependencies(&expandedPath, recipe.Element1, available, processing, db, startElements)

// 		// Recursively add missing dependencies for Element2
// 		addDependencies(&expandedPath, recipe.Element2, available, processing, db, startElements)

// 		// Now add the main recipe
// 		expandedPath = append(expandedPath, recipe)

// 		// Mark the result as available
// 		result := findRecipeResult(recipe, db)
// 		if result != "" {
// 			available[result] = true
// 		}
// 	}

// 	return expandedPath
// }

// func addDependencies(expandedPath *[]model.Recipe, elementID string, available, processing map[string]bool, db *model.ElementsDatabase, startElements []string) {
// 	// Check for cycles
// 	if processing[elementID] {
// 		log.Printf("Cycle detected: %s is already being processed", elementID)
// 		return
// 	}

// 	// If already available or basic, nothing to do
// 	if available[elementID] || isBasicElement(elementID, startElements) {
// 		return
// 	}

// 	// Mark as being processed
// 	processing[elementID] = true

// 	// Find the recipe to make this element
// 	element, exists := db.Elements[elementID]
// 	if !exists || len(element.Recipes) == 0 {
// 		processing[elementID] = false
// 		return
// 	}

// 	recipe := element.Recipes[0] // Use first available recipe

// 	// Check if this recipe would create a cycle
// 	if recipe.Element1 == elementID || recipe.Element2 == elementID {
// 		log.Printf("Self-referential recipe detected for %s", elementID)
// 		processing[elementID] = false
// 		return
// 	}

// 	// First, recursively add dependencies for the ingredients of this recipe
// 	addDependencies(expandedPath, recipe.Element1, available, processing, db, startElements)
// 	addDependencies(expandedPath, recipe.Element2, available, processing, db, startElements)

// 	// Now add this recipe
// 	*expandedPath = append(*expandedPath, recipe)

// 	// Mark this element as available and done processing
// 	available[elementID] = true
// 	processing[elementID] = false
// }
// func isBasicElement(elementID string, startElements []string) bool {
// 	return slices.Contains(startElements, elementID)
// }
// func findRecipeResult(recipe model.Recipe, db *model.ElementsDatabase) string {
// 	for elementName, element := range db.Elements {
// 		for _, r := range element.Recipes {
// 			if (r.Element1 == recipe.Element1 && r.Element2 == recipe.Element2) ||
// 				(r.Element1 == recipe.Element2 && r.Element2 == recipe.Element1) {
// 				return elementName
// 			}
// 		}
// 	}
// 	return ""
// }
// func MultiBFS(db *model.ElementsDatabase, targetElement string, maxPaths int, step chan<- *SearchProgress) *BFSResult {
// 	sortedDb := utility.SortByTier(db)
// 	result := make(chan *BFSResult, 1)
// 	startElement := []string{"Air", "Water", "Fire", "Earth"}
// 	// Run BFS in a goroutine
// 	go BFS(sortedDb, startElement, targetElement, maxPaths, result, step)

// 	// Wait for result
// 	results := <-result

// 	return results
// }

// ================================= FIX ==========================

package algorithm

import (
	"container/list"
	"fmt"
	"log"
	"shared/model"
	"shared/utility"
	"slices"
)

type BFSResult struct {
	TargetElement string           `json:"target_element"`
	Paths         [][]model.Recipe `json:"recipes"`
	VisitedNodes  int              `json:"visited_nodes"`
}

type BFSNode struct {
	Element    string         `json:"element"`
	Path       []model.Recipe `json:"path"` // Akan berisi Recipe dengan Result
	ParentNode *BFSNode
}

type SearchProgress struct {
	CurrentElement string          `json:"currentElement"`
	Visited        int             `json:"visited"`
	PathsFound     int             `json:"pathsFound"`
	VisitedNodes   map[string]bool `json:"visitedNodes"`
}

func BFS(db *model.ElementsDatabase, startElements []string, targetElement string, maxPath int, resultChan chan<- *BFSResult, step chan<- *SearchProgress) { // Ganti nama variabel result menjadi resultChan
	target, exists := db.Elements[targetElement]
	if !exists {
		resultChan <- &BFSResult{ // Menggunakan resultChan
			TargetElement: targetElement,
			Paths:         [][]model.Recipe{},
			VisitedNodes:  0,
		}
		close(resultChan) // Menggunakan resultChan
		return
	}
	if target.IsBasic {
		resultChan <- &BFSResult{ // Menggunakan resultChan
			TargetElement: targetElement,
			Paths:         [][]model.Recipe{},
			VisitedNodes:  1,
		}
		close(resultChan) // Menggunakan resultChan
		return
	}

	visitedCombinations := make(map[string]bool)
	discoveredElements := make(map[string]bool)
	queue := list.New()

	for _, basic := range startElements {
		node := &BFSNode{
			Element:    basic,
			Path:       []model.Recipe{},
			ParentNode: nil,
		}
		queue.PushBack(node)
		discoveredElements[basic] = true
	}

	paths := [][]model.Recipe{}
	visitedCount := 0

	for queue.Len() > 0 {
		if len(paths) >= maxPath && maxPath > 0 { // Tambahkan kondisi untuk maxPath jika diperlukan
			break
		}
		visitedCount++
		node := queue.Remove(queue.Front()).(*BFSNode)

		if step != nil {
			step <- &SearchProgress{
				CurrentElement: node.Element,
				Visited:        visitedCount,
				PathsFound:     len(paths),
				VisitedNodes:   discoveredElements, // Kirim discoveredElements agar lebih akurat
			}
		}

		// Cek apakah node saat ini adalah targetElement
		// Penting: Jika elemen ditemukan, Path di node sudah berisi langkah-langkah dengan Result-nya.
		if node.Element == targetElement {
			// Salin path agar aman
			finalPath := make([]model.Recipe, len(node.Path))
			copy(finalPath, node.Path)
			paths = append(paths, finalPath)
			if len(paths) >= maxPath && maxPath > 0 { // Cek lagi setelah menambah path
				break
			}
			continue // Lanjutkan mencari path lain jika mode multiple
		}

		// Kombinasikan node ini dengan semua elemen yang sudah ditemukan sebelumnya
		// atau dengan elemen dasar jika itu adalah strategi yang diinginkan
		elementsToCombineWith := make([]string, 0, len(discoveredElements))
		for el := range discoveredElements {
			elementsToCombineWith = append(elementsToCombineWith, el)
		}
		// Jika ingin selalu bisa kombinasi dengan elemen dasar juga:
		// baseElementsMap := make(map[string]bool)
		// for _, se := range startElements {
		// 	baseElementsMap[se] = true
		// }
		// for el := range baseElementsMap {
		// 	if !discoveredElements[el] { // Hanya tambahkan jika belum ada di discovered
		// 		elementsToCombineWith = append(elementsToCombineWith, el)
		// 	}
		// }

		for _, otherElementID := range elementsToCombineWith {
			// Hindari mengkombinasikan elemen dengan dirinya sendiri jika tidak menghasilkan apa-apa
			// atau jika kombinasi tersebut sudah dikunjungi
			if node.Element == otherElementID && !canCombineWithSelf(node.Element, db) {
				// (canCombineWithSelf adalah fungsi hipotetis, Anda mungkin perlu logika spesifik)
				// Atau cukup andalkan visitedCombinations
			}

			e1, e2 := node.Element, otherElementID
			if e1 > e2 {
				e1, e2 = e2, e1
			}
			combinationKey := fmt.Sprintf("%s+%s", e1, e2)
			if visitedCombinations[combinationKey] {
				continue
			}
			// Tandai kombinasi ini sebagai dikunjungi SEBELUM masuk ke loop resep
			// untuk mencegah pemrosesan berulang dari antrian yang sama.
			// Namun, untuk BFS yang benar, penandaan visit idealnya saat node di-pop.
			// Di sini, kita menandai *kombinasi bahan* yang telah dicoba.

			// Iterasi melalui semua kemungkinan hasil di database
			for resultElementID, resultElementData := range db.Elements {
				if resultElementData.IsBasic { // Hasil tidak mungkin elemen dasar
					continue
				}
				resultTier := utility.ParseTier(resultElementData.Tier)

				for _, dbRecipe := range resultElementData.Recipes { // dbRecipe adalah resep dari database
					// Cek apakah resep ini menggunakan e1 dan e2
					if (dbRecipe.Element1 == e1 && dbRecipe.Element2 == e2) || (dbRecipe.Element1 == e2 && dbRecipe.Element2 == e1) {
						// Validasi tier (elemen pembentuk harus memiliki tier lebih rendah dari hasilnya)
						r1Data, ok1 := db.Elements[dbRecipe.Element1]
						r2Data, ok2 := db.Elements[dbRecipe.Element2]
						if !ok1 || !ok2 {
							continue
						}
						t1 := utility.ParseTier(r1Data.Tier)
						t2 := utility.ParseTier(r2Data.Tier)

						// Tier elemen pembentuk harus lebih kecil dari tier hasil, atau sama jika merupakan kombinasi dasar (misal Air+Air=Pressure)
						// Logika tier ini penting untuk mencegah loop atau pencarian yang tidak efisien.
						// Tier hasil harus lebih besar dari atau sama dengan tier pembentuk terbesar.
						// Dan tier pembentuk tidak boleh lebih besar dari tier hasil.
						if t1 >= resultTier && !isSpecialCombination(dbRecipe.Element1, resultTier, db) {
							continue
						}
						if t2 >= resultTier && !isSpecialCombination(dbRecipe.Element2, resultTier, db) {
							continue
						}

						// Buat langkah resep dengan Result yang diisi
						currentStepRecipe := model.Recipe{
							Element1: dbRecipe.Element1, // Gunakan bahan dari dbRecipe
							Element2: dbRecipe.Element2,
							Result:   resultElementID, // resultElementID adalah hasil dari kombinasi ini
						}

						newPath := make([]model.Recipe, len(node.Path)+1)
						copy(newPath, node.Path)
						newPath[len(node.Path)] = currentStepRecipe

						// Buat node baru untuk elemen hasil
						newNode := &BFSNode{
							Element:    resultElementID,
							Path:       newPath,
							ParentNode: node,
						}

						// Hanya tambahkan ke antrian jika elemen hasil belum ditemukan
						// atau jika kita ingin menemukan semua jalur (maka discoveredElements mungkin perlu penyesuaian)
						// Untuk BFS standar mencari jalur terpendek, kita biasanya tidak menambahkan jika sudah di visit (discovered)
						// Tapi karena kita ingin tree, kita mungkin perlu logika berbeda atau mengandalkan visitedCombinations
						if !discoveredElements[resultElementID] {
							discoveredElements[resultElementID] = true // Tandai sebagai ditemukan
							queue.PushBack(newNode)
						} else if resultElementID == targetElement { // Jika target, tetap tambahkan untuk jalur alternatif
							queue.PushBack(newNode)
						}
						// Jika ingin mengizinkan semua jalur, bahkan yang lebih panjang, selalu PushBack
						// queue.PushBack(newNode)
					}
				}
			}
			visitedCombinations[combinationKey] = true // Tandai kombinasi bahan telah diproses
		}
	}

	expandedPaths := [][]model.Recipe{}
	for _, p := range paths {
		expandedPaths = append(expandedPaths, expandPath(p, db, startElements))
	}

	resultChan <- &BFSResult{ // Menggunakan resultChan
		TargetElement: targetElement,
		Paths:         expandedPaths, // Kirim jalur yang sudah diekspansi
		VisitedNodes:  visitedCount,
	}
	close(resultChan) // Menggunakan resultChan
}

// Fungsi pembantu untuk memeriksa apakah sebuah elemen dapat dikombinasikan dengan dirinya sendiri
func canCombineWithSelf(elementID string, db *model.ElementsDatabase) bool {
	elementData, exists := db.Elements[elementID]
	if !exists {
		return false
	}
	for _, recipe := range elementData.Recipes {
		if recipe.Element1 == elementID && recipe.Element2 == elementID {
			return true // Jika elemen ini adalah hasil dari dirinya sendiri + dirinya sendiri
		}
	}
	// Cek juga apakah elemen ini adalah bahan untuk dirinya sendiri (seharusnya tidak terjadi dalam data yang valid)
	for _, otherElementData := range db.Elements {
		for _, r := range otherElementData.Recipes {
			if r.Element1 == elementID && r.Element2 == elementID && otherElementData.Name == elementID {
				return true
			}
		}
	}
	return false
}

// Fungsi pembantu untuk kasus khusus tiering (misal, elemen dasar menghasilkan elemen dengan tier sama)
func isSpecialCombination(elementName string, resultTier int, db *model.ElementsDatabase) bool {
	// Contoh: Jika elemen dasar (tier 0 atau 1) bisa menghasilkan elemen lain dengan tier yang sama.
	// Ini perlu disesuaikan dengan aturan spesifik game Anda.
	elData, exists := db.Elements[elementName]
	if !exists {
		return false
	}
	elTier := utility.ParseTier(elData.Tier)
	// Misal, izinkan jika elemen pembentuk adalah tier 1 dan hasil juga tier 1 (seperti Air + Air = Pressure)
	if elTier <= 1 && resultTier <= 1 { // Sesuaikan angka tier sesuai kebutuhan
		return true
	}
	return false
}

func expandPath(path []model.Recipe, db *model.ElementsDatabase, startElements []string) []model.Recipe {
	available := make(map[string]bool)
	for _, elem := range startElements {
		available[elem] = true
	}

	finalExpandedPath := []model.Recipe{}

	// Iterasi melalui path yang sudah memiliki Result
	for _, recipeFromPath := range path {
		// recipeFromPath sudah memiliki Element1, Element2, dan Result

		// Pastikan dependensi (Element1 & Element2 dari recipeFromPath) ada
		// Ini akan menambahkan langkah-langkah untuk membuat Element1 dan Element2 jika belum tersedia
		addDependencies(&finalExpandedPath, recipeFromPath.Element1, available, make(map[string]bool), db, startElements)
		addDependencies(&finalExpandedPath, recipeFromPath.Element2, available, make(map[string]bool), db, startElements)

		// Tambahkan langkah resep saat ini (yang sudah memiliki Result)
		// Hanya tambahkan jika belum ada di finalExpandedPath (untuk menghindari duplikasi jika dependensi tumpang tindih)
		// Perlu cara yang lebih baik untuk cek duplikasi langkah spesifik ini.
		// Untuk sementara, kita asumsikan urutan dari BFS sudah cukup baik.
		if !isRecipeInPath(recipeFromPath, finalExpandedPath) {
			finalExpandedPath = append(finalExpandedPath, recipeFromPath)
		}

		// Tandai hasil dari resep ini sebagai tersedia
		available[recipeFromPath.Result] = true
	}

	return finalExpandedPath
}

// Fungsi pembantu untuk mengecek apakah resep sudah ada di path
func isRecipeInPath(recipe model.Recipe, path []model.Recipe) bool {
	for _, r := range path {
		if r.Element1 == recipe.Element1 && r.Element2 == recipe.Element2 && r.Result == recipe.Result {
			return true
		}
		// Pertimbangkan kasus urutan elemen berbeda jika data tidak selalu terurut
		if r.Element1 == recipe.Element2 && r.Element2 == recipe.Element1 && r.Result == recipe.Result {
			return true
		}
	}
	return false
}

func addDependencies(expandedPath *[]model.Recipe, elementID string, available map[string]bool, processing map[string]bool, db *model.ElementsDatabase, startElements []string) {
	if available[elementID] || isBasicElement(elementID, startElements) {
		return
	}

	if processing[elementID] {
		log.Printf("Cycle detected during expandPath for: %s", elementID)
		return
	}
	processing[elementID] = true

	elementData, exists := db.Elements[elementID]
	if !exists || len(elementData.Recipes) == 0 { // Jika elemen tidak ada atau tidak punya resep (seharusnya elemen dasar)
		log.Printf("Element %s has no recipes or does not exist, considered basic or unreachable in expandPath.", elementID)
		processing[elementID] = false // Reset status processing
		available[elementID] = true   // Anggap tersedia jika elemen dasar atau tidak bisa dibuat lebih lanjut
		return
	}

	// Pilih satu resep untuk membuat elementID (misalnya, yang pertama)
	// Penting: resep dari DB ini belum memiliki field 'Result' yang diisi secara kontekstual untuk *langkah ini*.
	dbRecipe := elementData.Recipes[0]

	// Buat instance Recipe untuk langkah ini, dengan Result diisi
	stepRecipe := model.Recipe{
		Element1: dbRecipe.Element1,
		Element2: dbRecipe.Element2,
		Result:   elementID, // elementID adalah yang ingin kita buat (hasil dari langkah ini)
	}

	// Rekursif tambahkan dependensi untuk bahan-bahan dari stepRecipe
	addDependencies(expandedPath, stepRecipe.Element1, available, processing, db, startElements)
	addDependencies(expandedPath, stepRecipe.Element2, available, processing, db, startElements)

	// Tambahkan stepRecipe ke expandedPath setelah dependensinya terpenuhi
	// Hanya tambahkan jika belum ada untuk menghindari duplikasi
	if !isRecipeInPath(stepRecipe, *expandedPath) {
		*expandedPath = append(*expandedPath, stepRecipe)
	}

	available[elementID] = true // Tandai elemen ini sekarang tersedia
	processing[elementID] = false
}

func isBasicElement(elementID string, startElements []string) bool {
	return slices.Contains(startElements, elementID)
}

// findRecipeResult tidak lagi dipanggil secara langsung oleh expandPath jika Result sudah ada.
// Namun, bisa berguna untuk keperluan lain atau jika ada bagian yang belum diubah.
func findRecipeResult(recipe model.Recipe, db *model.ElementsDatabase) string {
	// Fungsi ini mencari elemen apa yang dihasilkan oleh kombinasi recipe.Element1 dan recipe.Element2
	// Ini berguna jika Anda memiliki Recipe struct yang hanya berisi Element1 dan Element2.
	// Namun, jika Recipe sudah memiliki field Result, Anda bisa langsung menggunakannya.
	if recipe.Result != "" { // Jika Result sudah ada, kembalikan itu.
		return recipe.Result
	}

	// Logika fallback jika Result belum diisi di struct input recipe
	for elementName, elementData := range db.Elements {
		for _, r := range elementData.Recipes { // r adalah resep dari database
			// Cocokkan bahan
			if (r.Element1 == recipe.Element1 && r.Element2 == recipe.Element2) ||
				(r.Element1 == recipe.Element2 && r.Element2 == recipe.Element1) {
				return elementName // elementName adalah hasil dari resep r
			}
		}
	}
	return "" // Seharusnya tidak terjadi jika data valid dan resep ditemukan
}

func MultiBFS(db *model.ElementsDatabase, targetElement string, maxPaths int, step chan<- *SearchProgress) *BFSResult {
	// utility.SortByTier(db) // Sorting mungkin tidak diperlukan lagi jika logika tier sudah benar
	resultChan := make(chan *BFSResult, 1) // Menggunakan resultChan
	startElement := []string{"Air", "Water", "Fire", "Earth"}

	go BFS(db, startElement, targetElement, maxPaths, resultChan, step) // Menggunakan resultChan

	results := <-resultChan // Menggunakan resultChan

	return results
}

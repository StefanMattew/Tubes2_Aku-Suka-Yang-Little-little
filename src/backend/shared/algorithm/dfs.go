
package algorithm

import (
	"shared/model"
	"shared/utility"
)

type DFSResult struct {
	TargetElement string           `json:"target_element"`
	Paths         [][]model.Recipe `json:"recipes"`
	VisitedNodes  int              `json:"visited_nodes"`
}

func DFS(db *model.ElementsDatabase, startElements []string, targetElement string, maxPath int, result chan<- *DFSResult, step chan<- *SearchProgress) {
	target, exists := db.Elements[targetElement]
	if !exists {
		result <- &DFSResult{
			TargetElement: targetElement,
			Paths:         [][]model.Recipe{},
			VisitedNodes:  0,
		}
		close(result)
		return
	}
	if target.IsBasic {
		result <- &DFSResult{
			TargetElement: targetElement,
			Paths:         [][]model.Recipe{},
			VisitedNodes:  1,
		}
		close(result)
		return
	}

	visitedElements := make(map[string]bool) // Track visited elements, not combinations
	paths := make([][]model.Recipe, 0)
	visitedCount := 0

	// Inisialisasi dengan elemen dasar
	for _, el := range startElements {
		visitedElements[el] = true
	}

	var dfsRecursive func(current string, path []model.Recipe, depth int) bool
	dfsRecursive = func(current string, path []model.Recipe, depth int) bool {

		if depth > 100 {
			return false
		}

		visitedCount++

		if step != nil {
			step <- &SearchProgress{
				CurrentElement: current,
				Visited:        visitedCount,
				PathsFound:     len(paths),
				VisitedNodes:   visitedElements,
			}
		}

		if current == targetElement {
			paths = append(paths, path)
			return maxPath > 0 && len(paths) >= maxPath
		}

		for resultElementID, resultElement := range db.Elements {
			resultTier := utility.ParseTier(resultElement.Tier)

			for _, recipe := range resultElement.Recipes {
				// Check if the recipe creates the current element
				if resultElementID == current {
					// Check if ingredients are already discovered
					if visitedElements[recipe.Element1] && visitedElements[recipe.Element2] {
						r1, ok1 := db.Elements[recipe.Element1]
						r2, ok2 := db.Elements[recipe.Element2]

						if !ok1 || !ok2 {
							continue
						}

						t1 := utility.ParseTier(r1.Tier)
						t2 := utility.ParseTier(r2.Tier)

						if t1 >= resultTier || t2 >= resultTier {
							continue
						}

						// Simpan path
						newPath := make([]model.Recipe, len(path)+1)
						copy(newPath, path)
						newPath[len(path)] = recipe

						// Mark result element as visited (discovered)
						visitedElements[resultElementID] = true

						// Lanjut DFS ke elemen hasil kombinasi (result)
						if dfsRecursive(recipe.Element1, newPath, depth+1) {
							return true
						}
						if dfsRecursive(recipe.Element2, newPath, depth+1) {
							return true
						}

						// Backtrack: Unmark if path not found (optional, but can help with finding other paths)
						visitedElements[resultElementID] = false
					}
				}
			}
		}

		return false
	}

	// Start DFS from basic elements
	for _, basicElement := range startElements {
		dfsRecursive(basicElement, []model.Recipe{}, 0)

		// If we've found enough paths, exit
		if maxPath > 0 && len(paths) >= maxPath {
			break
		}
	}

	result <- &DFSResult{
		TargetElement: targetElement,
		Paths:         paths,
		VisitedNodes:  visitedCount,
	}
	close(result)
}

// Multithreading DFS
func MultiDFS(db *model.ElementsDatabase, targetElement string, maxPath int, step chan<- *SearchProgress) *DFSResult {
	sortedDb := utility.SortByTier(db)
	result := make(chan *DFSResult, 1)
	startElement := []string{"Air", "Water", "Fire", "Earth"}

	go DFS(sortedDb, startElement, targetElement, maxPath, result, step)

	results := <-result

	return results
}
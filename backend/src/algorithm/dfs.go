package algorithm

import (
	"backend/src/model"
	"backend/src/utility"
	"fmt"
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

	visited := make(map[string]bool)
	paths := make([][]model.Recipe, 0)
	visitedCount := 0

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
				VisitedNodes:   visited,
			}
		}

		if current == targetElement {
			paths = append(paths, path)
			return maxPath > 0 && len(paths) >= maxPath
		}

		element, exists := db.Elements[current]
		if !exists {
			return false
		}

		// If basic element
		// if element.IsBasic {
		// 	return false
		// }
		currentTier := utility.ParseTier(element.Tier)

		// Look for recipes to create this element
		for _, recipe := range element.Recipes {
			r1, ok1 := db.Elements[recipe.Element1]
			r2, ok2 := db.Elements[recipe.Element2]

			// validasi tier elemen pembentuk harus lebih kecil dari tier target
			if !ok1 || !ok2 || utility.ParseTier(r1.Tier) >= currentTier || utility.ParseTier(r2.Tier) >= currentTier {
				continue
			}

			recipeKey := fmt.Sprintf("%s-%s-%s", current, recipe.Element1, recipe.Element2)
			if visited[recipeKey] {
				continue
			}

			visited[recipeKey] = true

			newPath := make([]model.Recipe, len(path)+1)
			copy(newPath, path)
			newPath[len(path)] = recipe

			if dfsRecursive(recipe.Element1, newPath, depth+1) {
				return true
			}
			if dfsRecursive(recipe.Element2, newPath, depth+1) {
				return true
			}
		}

		return false
	}

	// Start DFS from basic elements
	for _, basicElement := range startElements {
		visited[basicElement] = true
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

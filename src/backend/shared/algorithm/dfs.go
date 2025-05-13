package algorithm

import (
	"fmt"
	"shared/model"
	"shared/utility"
)

type DFSResult struct {
	TargetElement string           `json:"target_element"`
	Paths         [][]model.Recipe `json:"recipes"`
	VisitedNodes  int              `json:"visited_nodes"`
}

type DFSNode struct {
	Element    string         `json:"element"`
	Path       []model.Recipe `json:"path"`
	ParentNode *DFSNode
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

	visitedCombinations := make(map[string]bool)
	paths := make([][]model.Recipe, 0)
	visitedCount := 0

	var dfsRecursive func(current string, path []model.Recipe, depth int)
	dfsRecursive = func(current string, path []model.Recipe, depth int) {
		if len(paths) >= maxPath {
			return // Hentikan jika sudah menemukan cukup banyak jalur.
		}

		visitedCount++
		if step != nil {
			step <- &SearchProgress{
				CurrentElement: current,
				Visited:        visitedCount,
				PathsFound:     len(paths),
				VisitedNodes:   visitedCombinations,
			}
		}

		if current == targetElement {
			newPath := make([]model.Recipe, len(path))
			copy(newPath, path)
			paths = append(paths, newPath)
			return
		}

		for elementID := range db.Elements {
			//Cek kombinasi current dengan element lain.
			e1, e2 := current, elementID
			if e1 > e2 {
				e1, e2 = e2, e1
			}
			combinationKey := fmt.Sprintf("%s+%s", e1, e2)

			if visitedCombinations[combinationKey] {
				continue
			}

			visitedCombinations[combinationKey] = true

			for resultElementID, resultElement := range db.Elements {
				resultTier := utility.ParseTier(resultElement.Tier)
				for _, recipe := range resultElement.Recipes {
					if (recipe.Element1 == e1 && recipe.Element2 == e2) || (recipe.Element1 == e2 && recipe.Element2 == e1) {
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
						newPath := make([]model.Recipe, len(path)+1)
						copy(newPath, path)
						newPath[len(path)] = recipe
						dfsRecursive(resultElementID, newPath, depth+1)
					}
				}
			}
		}
	}

	for _, basicElement := range startElements {
		dfsRecursive(basicElement, []model.Recipe{}, 0)
	}

	for i := range paths {
		paths[i] = expandPath(paths[i], db, startElements)
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

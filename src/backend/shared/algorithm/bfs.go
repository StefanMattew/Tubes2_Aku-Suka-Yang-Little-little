
package algorithm

import (
	"shared/model"
	"shared/utility"
	"container/list"
	"fmt"
)

type BFSResult struct {
	TargetElement string           `json:"target_element"`
	Paths         [][]model.Recipe `json:"recipes"`
	VisitedNodes  int              `json:"visited_nodes"`
}

type BFSNode struct {
	Element    string         `json:"element"`
	Path       []model.Recipe `json:"path"`
	ParentNode *BFSNode
}

type SearchProgress struct {
	CurrentElement string          `json:"currentElement"`
	Visited        int             `json:"visited"`
	PathsFound     int             `json:"pathsFound"`
	VisitedNodes   map[string]bool `json:"visitedNodes"`
}

func BFS(db *model.ElementsDatabase, startElements []string, targetElement string, maxPath int, result chan<- *BFSResult, step chan<- *SearchProgress) {
	target, exists := db.Elements[targetElement]
	if !exists {
		result <- &BFSResult{
			TargetElement: targetElement,
			Paths:         [][]model.Recipe{},
			VisitedNodes:  0,
		}
		close(result)
		return
	}
	if target.IsBasic {
		result <- &BFSResult{
			TargetElement: targetElement,
			Paths:         [][]model.Recipe{},
			VisitedNodes:  1,
		}
		close(result)
		return
	}

	visitedCombinations := make(map[string]bool)
	discoveredElements := make(map[string]bool) // Semua elemen yang sudah ditemukan dari hasil kombinasi
	queue := list.New()

	// Masukkan elemen dasar ke queue dan discovered
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

	for queue.Len() > 0 && (maxPath <= 0 || len(paths) < maxPath) {
		visitedCount++
		node := queue.Remove(queue.Front()).(*BFSNode)

		if step != nil {
			step <- &SearchProgress{
				CurrentElement: node.Element,
				Visited:        visitedCount,
				PathsFound:     len(paths),
				VisitedNodes:   discoveredElements,
			}
		}

		if node.Element == targetElement {
			paths = append(paths, node.Path)
			if maxPath > 0 && len(paths) >= maxPath {
				break
			}
			continue
		}

		// Kombinasikan node ini dengan semua elemen yang sudah ditemukan sebelumnya
		for otherElementID := range discoveredElements {
			e1, e2 := node.Element, otherElementID
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
					// Check if the recipe uses the current element and another discovered element
					if (recipe.Element1 == e1 && recipe.Element2 == e2) || (recipe.Element1 == e2 && recipe.Element2 == e1) {
						// Cek tier agar tidak lompat ke atas
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

						// Kombinasi valid, buat node baru
						newPath := make([]model.Recipe, len(node.Path)+1)
						copy(newPath, node.Path)
						newPath[len(node.Path)] = recipe

						newNode := &BFSNode{
							Element:    resultElementID,
							Path:       newPath,
							ParentNode: node,
						}
						queue.PushBack(newNode)

						if !discoveredElements[resultElementID] {
							discoveredElements[resultElementID] = true
						}
					}
				}
			}
		}
	}

	result <- &BFSResult{
		TargetElement: targetElement,
		Paths:         paths,
		VisitedNodes:  visitedCount,
	}
	close(result)
}
func MultiBFS(db *model.ElementsDatabase, targetElement string, maxPaths int, step chan<- *SearchProgress) *BFSResult {
	sortedDb := utility.SortByTier(db)
	result := make(chan *BFSResult, 1)
	startElement := []string{"Air", "Water", "Fire", "Earth"}
	// Run BFS in a goroutine
	go BFS(sortedDb, startElement, targetElement, maxPaths, result, step)

	// Wait for result
	results := <-result

	return results
}
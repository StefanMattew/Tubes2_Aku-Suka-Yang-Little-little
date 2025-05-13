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

func expandPath(path []model.Recipe, db *model.ElementsDatabase, startElements []string) []model.Recipe {
	isBasic := make(map[string]bool)
	elementRecipes := make(map[string]model.Recipe)
	visited := make(map[string]bool)
	// Mark basic elements
	for _, elem := range startElements {
		isBasic[elem] = true
	}

	// Build a complete recipe map with all dependencies
	var buildDependencies func(elementID string) []model.Recipe
	buildDependencies = func(elementID string) []model.Recipe {
		// If basic element or already processed, return empty
		if isBasic[elementID] {
			return []model.Recipe{}
		}
		if visited[elementID] {
			return nil
		}
		visited[elementID] = true

		// If we already have a recipe for this element, don't process again
		if _, exists := elementRecipes[elementID]; exists {
			return []model.Recipe{}
		}

		// Find recipe for this element
		element, exists := db.Elements[elementID]
		if !exists || len(element.Recipes) == 0 {
			return []model.Recipe{}
		}

		recipe := element.Recipes[0] // Take first recipe

		// Get dependencies for both ingredients
		deps := []model.Recipe{}
		deps = append(deps, buildDependencies(recipe.Element1)...)
		deps = append(deps, buildDependencies(recipe.Element2)...)

		// Add this recipe
		elementRecipes[elementID] = recipe
		deps = append(deps, recipe)

		return deps
	}

	// Build complete dependency list
	allRecipes := []model.Recipe{}
	for _, recipe := range path {
		// Add dependencies for both elements
		allRecipes = append(allRecipes, buildDependencies(recipe.Element1)...)
		allRecipes = append(allRecipes, buildDependencies(recipe.Element2)...)
		// Add the recipe itself
		allRecipes = append(allRecipes, recipe)
	}

	// Remove duplicates while preserving order
	seen := make(map[string]bool)
	uniqueRecipes := []model.Recipe{}
	for _, recipe := range allRecipes {
		key := fmt.Sprintf("%s+%s", recipe.Element1, recipe.Element2)
		if !seen[key] {
			seen[key] = true
			uniqueRecipes = append(uniqueRecipes, recipe)
		}
	}

	return uniqueRecipes
}

func MultiDFS(db *model.ElementsDatabase, targetElement string, maxPath int, step chan<- *SearchProgress) *DFSResult {
	sortedDb := utility.SortByTier(db)

	startElements := []string{"Air", "Water", "Fire", "Earth"}

	resultChan := make(chan *DFSResult, len(startElements))

	for _, elem := range startElements {
		go func(start string) {
			DFS(sortedDb, []string{start}, targetElement, maxPath, resultChan, step)
		}(elem)
	}

	finalPaths := [][]model.Recipe{}
	totalVisited := 0
	collected := 0

	for res := range resultChan {
		finalPaths = append(finalPaths, res.Paths...)
		totalVisited += res.VisitedNodes
		collected++

		if collected == len(startElements) {
			close(resultChan)
		}
	}

	return &DFSResult{
		TargetElement: targetElement,
		Paths:         finalPaths,
		VisitedNodes:  totalVisited,
	}
}

package algorithm

import (
	"context"
	"fmt"
	"log"
	"runtime"
	"shared/model"
	"shared/utility"
	"sync"
	"time"
)

type DFSResult struct {
	TargetElement string           `json:"target_element"`
	Paths         [][]model.Recipe `json:"recipes"`
	VisitedNodes  int              `json:"visited_nodes"`
}

// Bare DFS function for both single and multi-threaded
func DFSWithOptions(db *model.ElementsDatabase, startElements []string, targetElement string,
	bannedTargetRecipes []int, maxPaths int, result chan<- *DFSResult, progress chan<- *SearchProgress) {

	resultSent := false
	defer func() {
		if !resultSent {
			result <- &DFSResult{
				TargetElement: targetElement,
				Paths:         [][]model.Recipe{},
				VisitedNodes:  0,
			}
		}
		close(result)
	}()

	target, exists := db.Elements[targetElement]
	if !exists || target.IsBasic {
		result <- &DFSResult{
			TargetElement: targetElement,
			Paths:         [][]model.Recipe{},
			VisitedNodes:  0,
		}
		resultSent = true
		return
	}

	visitedCombinations := make(map[string]bool)

	// Add banned recipes to visitedCombinations with result
	for _, bannedIdx := range bannedTargetRecipes {
		if bannedIdx < len(target.Recipes) {
			recipe := target.Recipes[bannedIdx]
			e1, e2 := recipe.Element1, recipe.Element2
			if e1 > e2 {
				e1, e2 = e2, e1
			}
			// Include target element as result
			bannedKey := fmt.Sprintf("%s+%s->%s", e1, e2, targetElement)
			visitedCombinations[bannedKey] = true
			log.Printf("Banned combination: %s", bannedKey)
		}
	}

	paths := [][]model.Recipe{}
	visitedCount := 0

	// DFS recursive function
	var dfsSearch func(current string, currentPath []model.Recipe, availableElements map[string]bool)
	dfsSearch = func(current string, currentPath []model.Recipe, availableElements map[string]bool) {
		log.Printf("DFS: Exploring %s, path length: %d", current, len(currentPath))

		if maxPaths > 0 && len(paths) >= maxPaths {
			return
		}

		visitedCount++

		if progress != nil {
			select {
			case progress <- &SearchProgress{
				CurrentElement: current,
				Visited:        visitedCount,
				PathsFound:     len(paths),
				VisitedNodes:   availableElements,
			}:
			default:
			}
		}

		// Found target
		if current == targetElement {
			pathCopy := make([]model.Recipe, len(currentPath))
			copy(pathCopy, currentPath)
			paths = append(paths, pathCopy)
			log.Printf("Found path to %s with %d steps", targetElement, len(pathCopy))
			return
		}

		// Track available elements for this branch
		newAvailable := make(map[string]bool)
		for k, v := range availableElements {
			newAvailable[k] = v
		}
		newAvailable[current] = true

		// Try combinations with all available elements
		for otherElement := range newAvailable {
			if maxPaths > 0 && len(paths) >= maxPaths {
				return
			}

			e1, e2 := current, otherElement
			if e1 > e2 {
				e1, e2 = e2, e1
			}

			// Check all possible results for this combination
			for resultElementName, resultElement := range db.Elements {
				if resultElement.IsBasic {
					continue
				}

				//Skip if already available (removed this check per your requirements)
				if newAvailable[resultElementName] {
					continue
				}

				combinationKey := fmt.Sprintf("%s+%s->%s", e1, e2, resultElementName)
				if visitedCombinations[combinationKey] {
					continue
				}

				// Check recipes
				for _, recipe := range resultElement.Recipes {
					if (recipe.Element1 == e1 && recipe.Element2 == e2) ||
						(recipe.Element1 == e2 && recipe.Element2 == e1) {

						if !isValidTierProgression(recipe, resultElement, db) {
							continue
						}

						// Mark combination as visited for this search
						visitedCombinations[combinationKey] = true

						// Create recipe with result
						recipeWithResult := model.Recipe{
							Element1: recipe.Element1,
							Element2: recipe.Element2,
							Result:   resultElementName,
						}

						newPath := make([]model.Recipe, len(currentPath)+1)
						copy(newPath, currentPath)
						newPath[len(currentPath)] = recipeWithResult

						// Continue DFS
						dfsSearch(resultElementName, newPath, newAvailable)

						// Unmark to allow other paths
						if maxPaths > 1 {
							delete(visitedCombinations, combinationKey)
						}

						break // Only use first valid recipe
					}
				}
			}
		}
	}

	// Start DFS from each start element
	for _, startElement := range startElements {
		if maxPaths > 0 && len(paths) >= maxPaths {
			break
		}

		availableElements := make(map[string]bool)
		for _, elem := range startElements {
			availableElements[elem] = true
		}

		log.Printf("Starting DFS from %s", startElement)
		dfsSearch(startElement, []model.Recipe{}, availableElements)
	}

	log.Printf("DFS complete. Found %d paths", len(paths))

	// Apply iterative expansion to all paths
	expandedPaths := [][]model.Recipe{}
	for i, path := range paths {
		log.Printf("Expanding path %d", i+1)
		expanded := iterativeExpansion_DFS(path, db, startElements, progress)
		expandedPaths = append(expandedPaths, expanded)
	}

	result <- &DFSResult{
		TargetElement: targetElement,
		Paths:         expandedPaths,
		VisitedNodes:  visitedCount,
	}
	resultSent = true
}

// Single path DFS
func DFSSingle(db *model.ElementsDatabase, startElements []string, targetElement string,
	result chan<- *DFSResult, progress chan<- *SearchProgress) {
	DFSWithOptions(db, startElements, targetElement, []int{}, 1, result, progress)
}

// Multiple threaded DFS
func DFSMultipleThreaded(db *model.ElementsDatabase, startElements []string,
	targetElement string, maxPaths int, timeoutSeconds int,
	result chan<- *DFSResult) {

	numWorkers := runtime.NumCPU()
	pathsChan := make(chan []model.Recipe, maxPaths*2)
	tasks := make(chan DFSTask, 100)
	done := make(chan bool, 1)
	var mu sync.Mutex
	collectedPaths := make([][]model.Recipe, 0, maxPaths)
	totalVisited := 0
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSeconds)*time.Second)
	defer cancel()

	// Path collector
	go func() {
		for {
			select {
			case path, ok := <-pathsChan:
				if !ok {
					done <- true
					return
				}
				mu.Lock()
				if len(collectedPaths) < maxPaths && !isDuplicatePath(path, collectedPaths) {
					collectedPaths = append(collectedPaths, path)
					log.Printf("DFS: Found path %d/%d with %d steps", len(collectedPaths), maxPaths, len(path))
				}
				mu.Unlock()

			case <-ctx.Done():
				done <- true
				return
			}
		}
	}()

	// Worker pool
	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			progressChan := make(chan *SearchProgress, 100)
			go func() {
				for range progressChan {
				}
			}()

			for {
				select {
				case task, ok := <-tasks:
					if !ok {
						close(progressChan)
						return
					}

					mu.Lock()
					shouldStop := len(collectedPaths) >= maxPaths
					mu.Unlock()

					if shouldStop {
						continue
					}

					log.Printf("DFS Worker %d: Processing task", workerID)

					searchCtx, searchCancel := context.WithTimeout(ctx, 30*time.Second)
					resultChan := make(chan *DFSResult, 1)

					go DFSWithOptions(db, task.Shuffle, targetElement, task.BannedRecipes,
						1, resultChan, progressChan)

					select {
					case dfsResult := <-resultChan:
						searchCancel()

						mu.Lock()
						totalVisited += dfsResult.VisitedNodes
						mu.Unlock()

						if len(dfsResult.Paths) > 0 {
							select {
							case pathsChan <- dfsResult.Paths[0]:
							case <-ctx.Done():
								return
							}
						}

					case <-searchCtx.Done():
						searchCancel()
						log.Printf("DFS Worker %d: Search timeout", workerID)

					case <-ctx.Done():
						searchCancel()
						close(progressChan)
						return
					}

				case <-ctx.Done():
					close(progressChan)
					return
				}
			}
		}(i)
	}

	// Task generator
	go func() {
		defer close(tasks)

		shuffles := generateFixedShuffles(startElements)

		targetElem, exists := db.Elements[targetElement]
		targetRecipeCount := 0
		if exists {
			targetRecipeCount = len(targetElem.Recipes)
		}

		// Try each shuffle
		for i, shuffle := range shuffles {
			mu.Lock()
			currentCount := len(collectedPaths)
			mu.Unlock()

			if currentCount >= maxPaths {
				break
			}

			task := DFSTask{
				Shuffle:       shuffle,
				BannedRecipes: []int{},
			}

			log.Printf("DFS: Queuing shuffle task %d", i)

			select {
			case tasks <- task:
			case <-ctx.Done():
				return
			}
		}

		time.Sleep(1 * time.Second)

		// Recipe banning phase
		bannedRecipes := []int{}
		for recipeIdx := 0; recipeIdx < targetRecipeCount; recipeIdx++ {
			mu.Lock()
			currentCount := len(collectedPaths)
			mu.Unlock()

			if currentCount >= maxPaths {
				break
			}

			bannedRecipes = append(bannedRecipes, recipeIdx)
			log.Printf("DFS: Banning recipe %d", recipeIdx)

			task := DFSTask{
				Shuffle:       startElements,
				BannedRecipes: append([]int{}, bannedRecipes...),
			}

			select {
			case tasks <- task:
			case <-ctx.Done():
				return
			}
		}
	}()

	wg.Wait()
	close(pathsChan)

	select {
	case <-done:
	case <-time.After(5 * time.Second):
	}

	mu.Lock()
	finalPaths := make([][]model.Recipe, len(collectedPaths))
	copy(finalPaths, collectedPaths)
	finalVisited := totalVisited
	mu.Unlock()

	result <- &DFSResult{
		TargetElement: targetElement,
		Paths:         finalPaths,
		VisitedNodes:  finalVisited,
	}
	close(result)
}

type DFSTask struct {
	Shuffle       []string
	BannedRecipes []int
}

// DFS Driver
func DFSDriver(db *model.ElementsDatabase, targetElement string, maxPaths int, step chan<- *SearchProgress) *DFSResult {
	sortedDb := utility.SortByTier(db)
	result := make(chan *DFSResult, 1)
	startElements := []string{"Air", "Water", "Fire", "Earth"}

	log.Printf("DFS Driver: Starting search for %s with maxPaths=%d", targetElement, maxPaths)

	if maxPaths == 1 {
		go DFSSingle(sortedDb, startElements, targetElement, result, step)
	} else if maxPaths > 1 {
		go DFSMultipleThreaded(sortedDb, startElements, targetElement, maxPaths, 30, result)
	} else {
		result <- &DFSResult{
			TargetElement: targetElement,
			Paths:         [][]model.Recipe{},
			VisitedNodes:  0,
		}
		close(result)
		return &DFSResult{
			TargetElement: targetElement,
			Paths:         [][]model.Recipe{},
			VisitedNodes:  0,
		}
	}

	results := <-result
	return results
}

// Iteratively run DFS to search for a path to a missing element.
func iterativeExpansion_DFS(path []model.Recipe, db *model.ElementsDatabase, startElements []string, step chan<- *SearchProgress) []model.Recipe {
	workingPath := make([]model.Recipe, len(path))
	copy(workingPath, path)

	// Make a createdElements list to track which elements have been created
	createdElements := make(map[string]bool)
	for _, elem := range startElements {
		createdElements[elem] = true
	}

	iteration := 0
	// Find a single missing element in the path from bottom up
	for {
		iteration++
		log.Printf("Expansion iteration %d", iteration)
		missingElement := ""
		missingPosition := -1

		for i, recipe := range workingPath {
			if !createdElements[recipe.Element1] {
				missingElement = recipe.Element1
				missingPosition = i
				log.Printf("Found missing element %s at position %d", missingElement, i)
				break
			}

			if !createdElements[recipe.Element2] {
				missingElement = recipe.Element2
				missingPosition = i
				log.Printf("Found missing element %s at position %d", missingElement, i)
				break
			}

			// Add the element to the createdElements list using the Result from Recipe
			if recipe.Result != "" {
				createdElements[recipe.Result] = true
			}
		}

		// Loop breaking
		if missingElement == "" {
			log.Printf("No missing elements found, expansion complete")
			break
		}

		log.Printf("Searching for missing element: %s", missingElement)
		subResult := make(chan *DFSResult, 1)
		availableElements := keysFromMap(createdElements)

		// Run DFS to find the path to the missing element
		go DFSWithOptions(db, availableElements, missingElement, []int{}, 1, subResult, nil)

		dfsResult := <-subResult
		if len(dfsResult.Paths) == 0 {
			log.Printf("Could not find path for %s, skipping", missingElement)
			break
		}

		subPath := dfsResult.Paths[0]
		log.Printf("Found path for %s with %d steps", missingElement, len(subPath))

		newPath := []model.Recipe{}

		// Insert the subPath in the correct position
		for i, recipe := range workingPath {
			if i == missingPosition {
				newPath = append(newPath, subPath...)
			}
			newPath = append(newPath, recipe)
		}

		workingPath = newPath

		// Iteration limit
		if iteration > 50 {
			log.Printf("Max iterations reached, stopping expansion")
			break
		}
	}

	return workingPath
}

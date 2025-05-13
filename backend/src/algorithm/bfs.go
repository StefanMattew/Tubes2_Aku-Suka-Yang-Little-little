package algorithm

import (
	"backend/src/model"
	"backend/src/utility"
	"container/list"
	"context"
	"fmt"
	"log"
	"math/rand"
	"runtime"
	"slices"
	"sync"
	"time"
)

type SearchStrategy struct {
	Type             string
	Exclusions       map[string]bool
	PreferredTiers   []int
	ShuffledElements []string
}
type SearchParams struct {
	startElements []string
	strategy      SearchStrategy
}

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
	discoveredElements := make(map[string]bool)
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

	for queue.Len() > 0 {
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
			break
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
	//for i := range paths {
	//	paths[i] = expandPath(paths[i], db, startElements)
	//}
	result <- &BFSResult{
		TargetElement: targetElement,
		Paths:         paths,
		VisitedNodes:  visitedCount,
	}
	close(result)
}
func BFSWithCompleteExpansion(db *model.ElementsDatabase, startElements []string, targetElement string, result chan<- *BFSResult, step chan<- *SearchProgress) {
	// First BFS to find main path
	firstResult := make(chan *BFSResult, 1)
	BFS(db, startElements, targetElement, 1, firstResult, step)

	initialResult := <-firstResult
	if len(initialResult.Paths) == 0 {
		result <- initialResult
		close(result)
		return
	}

	// Get the first path
	path := initialResult.Paths[0]

	// Iteratively expand until all dependencies are satisfied
	expandedPath := iterativeExpansion(path, db, startElements, step)

	result <- &BFSResult{
		TargetElement: targetElement,
		Paths:         [][]model.Recipe{expandedPath},
		VisitedNodes:  initialResult.VisitedNodes,
	}
	close(result)
}

func iterativeExpansion(path []model.Recipe, db *model.ElementsDatabase, startElements []string, step chan<- *SearchProgress) []model.Recipe {
	// Keep track of available elements
	available := make(map[string]bool)

	// Mark basic elements as available
	for _, elem := range startElements {
		available[elem] = true
	}

	// Current working path
	workingPath := make([]model.Recipe, len(path))
	copy(workingPath, path)

	// Iterate until no more missing dependencies
	for {
		// Find all elements created by current path
		createdElements := make(map[string]bool)
		for _, recipe := range workingPath {
			result := findRecipeResult(recipe, db)
			if result != "" {
				createdElements[result] = true
			}
		}

		// Merge available and created elements
		allAvailable := make(map[string]bool)
		for elem := range available {
			allAvailable[elem] = true
		}
		for elem := range createdElements {
			allAvailable[elem] = true
		}

		// Find missing dependencies in the path
		missingElements := []string{}
		insertPositions := make(map[string]int) // Where to insert the sub-path for each missing element

		for i, recipe := range workingPath {
			// Check Element1
			if !allAvailable[recipe.Element1] {
				if !contains(missingElements, recipe.Element1) {
					missingElements = append(missingElements, recipe.Element1)
					result := findRecipeResult(recipe, db)
					insertPositions[recipe.Element1] = i
					if result != "" {
						allAvailable[result] = true
					}
					break
				}
			}

			// Check Element2
			if !allAvailable[recipe.Element2] {
				if !contains(missingElements, recipe.Element2) {
					missingElements = append(missingElements, recipe.Element2)
					insertPositions[recipe.Element2] = i
					result := findRecipeResult(recipe, db)
					if result != "" {
						allAvailable[result] = true
					}
					break
				}
			}

			// Mark this recipe's result as available for subsequent recipes
			result := findRecipeResult(recipe, db)
			if result != "" {
				allAvailable[result] = true
			}
		}

		// If no missing elements, we're done
		if len(missingElements) == 0 {
			break
		}

		// For each missing element, find how to make it
		newPath := []model.Recipe{}
		lastInsertPos := -1

		for i, recipe := range workingPath {
			// Insert any sub-paths needed before this recipe
			for _, missing := range missingElements {
				if insertPositions[missing] == i {
					// Only search if we haven't already inserted a path for this element
					if lastInsertPos < i {
						// Perform BFS to find how to make this missing element
						subResult := make(chan *BFSResult, 1)
						availableElements := keysFromMap(allAvailable)

						log.Printf("Searching for missing element: %s", missing)
						BFS(db, availableElements, missing, 1, subResult, step)

						subBFSResult := <-subResult
						if len(subBFSResult.Paths) > 0 {
							subPath := subBFSResult.Paths[0]
							log.Printf("Found path for %s: %d steps", missing, len(subPath))

							// Add all recipes from the sub-path
							for _, subRecipe := range subPath {
								newPath = append(newPath, subRecipe)

								// Update available elements
								subResult := findRecipeResult(subRecipe, db)
								if subResult != "" {
									allAvailable[subResult] = true
								}
							}
						}
					}
				}
			}

			// Add the original recipe
			newPath = append(newPath, recipe)
		}

		workingPath = newPath
	}

	return workingPath
}

// Helper function to check if slice contains element
func contains(slice []string, element string) bool {
	return slices.Contains(slice, element)
}

// Helper function to get keys from map
func keysFromMap(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
func findRecipeResult(recipe model.Recipe, db *model.ElementsDatabase) string {
	for elementName, element := range db.Elements {
		for _, r := range element.Recipes {
			if (r.Element1 == recipe.Element1 && r.Element2 == recipe.Element2) ||
				(r.Element1 == recipe.Element2 && r.Element2 == recipe.Element1) {
				return elementName
			}
		}
	}
	return ""
}

type BannedRecipes struct {
	mu     sync.Mutex
	banned map[string][]int // element -> list of banned recipe indices
}

func NewBannedRecipes() *BannedRecipes {
	return &BannedRecipes{
		banned: make(map[string][]int),
	}
}

func (br *BannedRecipes) BanRecipe(element string, recipeIndex int) {
	br.mu.Lock()
	defer br.mu.Unlock()

	if _, exists := br.banned[element]; !exists {
		br.banned[element] = make([]int, 0)
	}
	br.banned[element] = append(br.banned[element], recipeIndex)
}

func (br *BannedRecipes) IsRecipeBanned(element string, recipeIndex int) bool {
	br.mu.Lock()
	defer br.mu.Unlock()

	bannedIndices, exists := br.banned[element]
	if !exists {
		return false
	}

	return slices.Contains(bannedIndices, recipeIndex)
}

func BFSWithBannedRecipes(db *model.ElementsDatabase, startElements []string, targetElement string,
	bannedRecipes *BannedRecipes, result chan<- *BFSResult, progress chan<- *SearchProgress) {

	// Always ensure we send a result and close the channel
	resultSent := false

	// Always ensure we send a result and close the channel
	defer func() {
		if !resultSent {
			result <- &BFSResult{
				TargetElement: targetElement,
				Paths:         [][]model.Recipe{},
				VisitedNodes:  0,
			}
		}
		close(result)
	}()

	target, exists := db.Elements[targetElement]
	if !exists || target.IsBasic {
		result <- &BFSResult{
			TargetElement: targetElement,
			Paths:         [][]model.Recipe{},
			VisitedNodes:  0,
		}
		return
	}

	visitedCombinations := make(map[string]bool)
	discoveredElements := make(map[string]bool)
	queue := list.New()
	recipeUsed := make(map[string]int)

	// Initialize with start elements
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

	for queue.Len() > 0 && len(paths) == 0 {
		visitedCount++
		node := queue.Remove(queue.Front()).(*BFSNode)

		if progress != nil {
			select {
			case progress <- &SearchProgress{
				CurrentElement: node.Element,
				Visited:        visitedCount,
				PathsFound:     len(paths),
				VisitedNodes:   discoveredElements,
			}:
			default:
				// Don't block on progress
			}
		}

		if node.Element == targetElement {
			paths = append(paths, node.Path)
			break
		}

		// Try combinations
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

			// Check all possible results
			for resultElementID, resultElement := range db.Elements {
				// Skip if already discovered (unless it's the target)
				if discoveredElements[resultElementID] && resultElementID != targetElement {
					continue
				}

				// Track if we found any valid recipe for this result
				foundValidRecipe := false

				// Check each recipe in order
				for recipeIndex, recipe := range resultElement.Recipes {
					// Skip if this recipe is banned
					if bannedRecipes.IsRecipeBanned(resultElementID, recipeIndex) {
						log.Printf("Skipping banned recipe %d for %s", recipeIndex, resultElementID)
						continue
					}

					if (recipe.Element1 == e1 && recipe.Element2 == e2) ||
						(recipe.Element1 == e2 && recipe.Element2 == e1) {

						// Tier check
						if !isValidTierProgression(recipe, resultElement, db) {
							continue
						}

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
							recipeUsed[resultElementID] = recipeIndex
						}

						foundValidRecipe = true
						// Only use the first valid recipe found
						break
					}
				}

				if foundValidRecipe {
					// We found a recipe for this result, no need to check more
					break
				}
			}
		}
	}

	// Log if no path was found
	if len(paths) == 0 {
		log.Printf("No path found to %s with current banned recipes", targetElement)
	}

	result <- &BFSResult{
		TargetElement: targetElement,
		Paths:         paths,
		VisitedNodes:  visitedCount,
	}
}

// Multiple path BFS with better error handling
func BFSMultiplePathsWithBanning(db *model.ElementsDatabase, startElements []string,
	targetElement string, maxPaths int, timeoutSeconds int,
	result chan<- *BFSResult) {

	// Configuration
	numWorkers := runtime.NumCPU()

	// Channels for coordination
	pathsChan := make(chan []model.Recipe, maxPaths*2)
	workerTasks := make(chan int, maxPaths*3) // More buffer
	done := make(chan bool, 1)

	// Shared state
	var mu sync.Mutex
	collectedPaths := make([][]model.Recipe, 0, maxPaths)
	bannedRecipes := NewBannedRecipes()
	totalVisited := 0
	failedAttempts := 0

	// Context for timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSeconds)*time.Second)
	defer cancel()

	// Progress tracking
	progressChan := make(chan *SearchProgress, 1000)

	// Progress consumer
	go func() {
		for {
			select {
			case <-progressChan:
				// Consume progress messages
			case <-ctx.Done():
				return
			}
		}
	}()

	// Path collector goroutine
	go func() {
		for {
			select {
			case path, ok := <-pathsChan:
				if !ok {
					done <- true
					return
				}
				mu.Lock()
				if len(collectedPaths) < maxPaths && !isDuplicatePath(db, path, collectedPaths) {
					collectedPaths = append(collectedPaths, path)

					// Extract and ban recipes used in this path
					usedRecipes := extractUsedRecipes(path, db)
					for element, recipeIndex := range usedRecipes {
						log.Printf("Path %d: Banning recipe %d for element %s",
							len(collectedPaths), recipeIndex, element)
						bannedRecipes.BanRecipe(element, recipeIndex)
					}

					log.Printf("Found path %d/%d", len(collectedPaths), maxPaths)
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

			for {
				select {
				case taskID, ok := <-workerTasks:
					if !ok {
						return
					}

					// Check if we should stop
					mu.Lock()
					shouldStop := len(collectedPaths) >= maxPaths
					currentFailed := failedAttempts
					mu.Unlock()

					if shouldStop {
						return
					}

					// Stop if too many failures
					if currentFailed > maxPaths*2 {
						log.Printf("Worker %d: Too many failed attempts (%d), stopping", workerID, currentFailed)
						return
					}

					// Shuffle start elements for variety
					shuffled := make([]string, len(startElements))
					copy(shuffled, startElements)
					if taskID > 0 { // Don't shuffle on first attempt
						rand.Seed(time.Now().UnixNano() + int64(taskID))
						rand.Shuffle(len(shuffled), func(i, j int) {
							shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
						})
					}

					log.Printf("Worker %d: Starting search %d with elements %v", workerID, taskID, shuffled)

					// Perform search with current banned recipes
					searchCtx, searchCancel := context.WithTimeout(ctx, 20*time.Second) // Increased timeout
					resultChan := make(chan *BFSResult, 1)

					go BFSWithBannedRecipes(db, shuffled, targetElement,
						bannedRecipes, resultChan, progressChan)

					select {
					case bfsResult := <-resultChan:
						searchCancel()

						mu.Lock()
						totalVisited += bfsResult.VisitedNodes
						mu.Unlock()

						if len(bfsResult.Paths) > 0 {
							log.Printf("Worker %d: Found raw path with %d steps", workerID, len(bfsResult.Paths[0]))

							// Expand the path
							expandedPath := expandPathSimple(bfsResult.Paths[0], db, startElements)

							log.Printf("Worker %d: Expanded path to %d steps", workerID, len(expandedPath))

							select {
							case pathsChan <- expandedPath:
							case <-ctx.Done():
								return
							}
						} else {
							mu.Lock()
							failedAttempts++
							mu.Unlock()
							log.Printf("Worker %d: No path found (attempt %d)", workerID, taskID)
						}

					case <-searchCtx.Done():
						searchCancel()
						mu.Lock()
						failedAttempts++
						mu.Unlock()
						log.Printf("Worker %d: Search timeout for task %d", workerID, taskID)

					case <-ctx.Done():
						searchCancel()
						return
					}

				case <-ctx.Done():
					return
				}
			}
		}(i)
	}

	// Task generator
	go func() {
		defer close(workerTasks)

		maxAttempts := maxPaths * 5 // More attempts
		for i := 0; i < maxAttempts; i++ {
			mu.Lock()
			currentCount := len(collectedPaths)
			currentFailed := failedAttempts
			mu.Unlock()

			if currentCount >= maxPaths {
				log.Printf("Reached target path count: %d", currentCount)
				break
			}

			if currentFailed > maxPaths*2 {
				log.Printf("Too many failures, stopping task generation")
				break
			}

			select {
			case workerTasks <- i:
			case <-ctx.Done():
				log.Printf("Task generator: Context cancelled")
				return
			case <-time.After(100 * time.Millisecond):
				// Don't block forever on sending tasks
			}

			// Small delay between tasks
			time.Sleep(50 * time.Millisecond)
		}

		log.Printf("Task generator completed")
	}()

	// Wait for workers
	wg.Wait()
	log.Printf("All workers completed")

	// Close paths channel
	close(pathsChan)

	// Wait for collector
	select {
	case <-done:
		log.Printf("Collector completed")
	case <-time.After(5 * time.Second):
		log.Printf("Collector timeout")
	}

	close(progressChan)

	// Send final result
	mu.Lock()
	finalPaths := make([][]model.Recipe, len(collectedPaths))
	copy(finalPaths, collectedPaths)
	finalVisited := totalVisited
	mu.Unlock()

	log.Printf("Final result: %d paths found, %d nodes visited, %d failed attempts",
		len(finalPaths), finalVisited, failedAttempts)

	result <- &BFSResult{
		TargetElement: targetElement,
		Paths:         finalPaths,
		VisitedNodes:  finalVisited,
	}
	close(result)
}

// Extract used recipes from a path
func extractUsedRecipes(path []model.Recipe, db *model.ElementsDatabase) map[string]int {
	usedRecipes := make(map[string]int)

	for _, recipe := range path {
		// Find which element this recipe creates
		result := findRecipeResult(recipe, db)
		if result != "" {
			// Find the recipe index
			element := db.Elements[result]
			for recipeIndex, elementRecipe := range element.Recipes {
				if isSameRecipe(recipe, elementRecipe) {
					usedRecipes[result] = recipeIndex
					break
				}
			}
		}
	}

	return usedRecipes
}

// Check if path is duplicate
func isDuplicatePath(db *model.ElementsDatabase, newPath []model.Recipe, existingPaths [][]model.Recipe) bool {
	if len(existingPaths) == 0 {
		return false
	}

	// Create a set of recipes in the new path
	newRecipes := make(map[string]bool)
	for _, recipe := range newPath {
		result := findRecipeResult(recipe, db)
		key := fmt.Sprintf("%s+%s->%s", recipe.Element1, recipe.Element2, result)
		newRecipes[key] = true
	}

	// Check against existing paths
	for _, existingPath := range existingPaths {
		if len(existingPath) != len(newPath) {
			continue
		}

		matches := true
		for _, recipe := range existingPath {
			result := findRecipeResult(recipe, nil)
			key := fmt.Sprintf("%s+%s->%s", recipe.Element1, recipe.Element2, result)
			if !newRecipes[key] {
				matches = false
				break
			}
		}

		if matches {
			return true
		}
	}

	return false
}

// Simple path expansion
func expandPathSimple(path []model.Recipe, db *model.ElementsDatabase, startElements []string) []model.Recipe {
	return iterativeExpansion(path, db, startElements, nil)
}

// Check if two recipes are the same
func isSameRecipe(r1, r2 model.Recipe) bool {
	return (r1.Element1 == r2.Element1 && r1.Element2 == r2.Element2) ||
		(r1.Element1 == r2.Element2 && r1.Element2 == r2.Element1)
}

// Check tier progression
func isValidTierProgression(recipe model.Recipe, resultElement model.Element, db *model.ElementsDatabase) bool {
	r1, ok1 := db.Elements[recipe.Element1]
	r2, ok2 := db.Elements[recipe.Element2]
	if !ok1 || !ok2 {
		return false
	}

	resultTier := utility.ParseTier(resultElement.Tier)
	t1 := utility.ParseTier(r1.Tier)
	t2 := utility.ParseTier(r2.Tier)

	return t1 < resultTier && t2 < resultTier
}

func MultiBFS(db *model.ElementsDatabase, targetElement string, maxPaths int, step chan<- *SearchProgress) *BFSResult {
	sortedDb := utility.SortByTier(db)
	result := make(chan *BFSResult, 1)
	startElement := []string{"Air", "Water", "Fire", "Earth"}
	// Run BFS in a goroutine
	go BFSWithCompleteExpansion(sortedDb, startElement, targetElement, result, step)

	// Wait for result
	results := <-result

	return results
}

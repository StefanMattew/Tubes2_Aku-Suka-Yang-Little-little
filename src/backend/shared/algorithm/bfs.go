package algorithm

import (
	"container/list"
	"context"
	"fmt"
	"log"
	"runtime"
	"shared/model"
	"shared/utility"
	"sync"
	"time"
)

type SearchStrategy struct {
	Type             string
	Exclusions       map[string]bool
	PreferredTiers   []int
	ShuffledElements []string
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

// Iteratively run BFS to search for a path to a missing element.
func iterativeExpansion(path []model.Recipe, db *model.ElementsDatabase, startElements []string, step chan<- *SearchProgress) []model.Recipe {
	workingPath := make([]model.Recipe, len(path))
	copy(workingPath, path)

	//Make a createdElements list to track which elements have been created
	createdElements := make(map[string]bool)
	for _, elem := range startElements {
		createdElements[elem] = true
	}

	iteration := 0
	//Find a single missing element in the path from bottom up
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

			//Add the element to the createdElements list using the Result from Recipe
			if recipe.Result != "" {
				createdElements[recipe.Result] = true
			}
		}

		//Loop breaking
		if missingElement == "" {
			log.Printf("No missing elements found, expansion complete")
			break
		}

		log.Printf("Searching for missing element: %s", missingElement)
		subResult := make(chan *BFSResult, 1)
		availableElements := keysFromMap(createdElements)

		//Run BFS to find the path to the missing element
		go BFSWithOptions(db, availableElements, missingElement, []int{}, 1, subResult, nil)

		bfsResult := <-subResult
		if len(bfsResult.Paths) == 0 {
			log.Printf("Could not find path for %s, skipping", missingElement)
			break
		}

		subPath := bfsResult.Paths[0]
		log.Printf("Found path for %s with %d steps", missingElement, len(subPath))

		newPath := []model.Recipe{}

		//Insert the subPath in the correct position
		for i, recipe := range workingPath {
			if i == missingPosition {
				newPath = append(newPath, subPath...)
			}
			newPath = append(newPath, recipe)
		}

		workingPath = newPath

		//Iteration limit
		if iteration > 50 {
			log.Printf("Max iterations reached, stopping expansion")
			break
		}
	}

	return workingPath
}

// Helper to get keys from a map
func keysFromMap(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// Bare BFS function for both single and multi-threaded
func BFSWithOptions(db *model.ElementsDatabase, startElements []string, targetElement string,
	bannedTargetRecipes []int, maxPaths int, result chan<- *BFSResult, progress chan<- *SearchProgress) {

	resultSent := false
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
		resultSent = true
		return
	}

	visitedCombinations := make(map[string]bool)

	//Add banned recipes to visitedCombinations with result
	for _, bannedIdx := range bannedTargetRecipes {
		if bannedIdx < len(target.Recipes) {
			recipe := target.Recipes[bannedIdx]
			e1, e2 := recipe.Element1, recipe.Element2
			if e1 > e2 {
				e1, e2 = e2, e1
			}
			//Include target element as result
			bannedKey := fmt.Sprintf("%s+%s->%s", e1, e2, targetElement)
			visitedCombinations[bannedKey] = true
			log.Printf("Banned combination: %s", bannedKey)
		}
	}

	discoveredElements := make(map[string]bool)
	queue := list.New()

	//Initialize queue with start elements
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

	//BFS main loop
	for queue.Len() > 0 && (maxPaths <= 0 || len(paths) < maxPaths) {
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
			}
		}

		if node.Element == targetElement {
			paths = append(paths, node.Path)
			if maxPaths > 0 && len(paths) >= maxPaths {
				break
			}
			continue
		}

		//Try combinations with discovered elements
		for otherElementID := range discoveredElements {
			e1, e2 := node.Element, otherElementID
			if e1 > e2 {
				e1, e2 = e2, e1
			}

			//Check all possible results for this combination
			for resultElementName, resultElement := range db.Elements {
				if discoveredElements[resultElementName] && resultElementName != targetElement {
					continue
				}

				//Create combination key with result
				combinationKey := fmt.Sprintf("%s+%s->%s", e1, e2, resultElementName)
				//If recipe->result already visited, skip
				if visitedCombinations[combinationKey] {
					continue
				}

				for _, recipe := range resultElement.Recipes {
					if (recipe.Element1 == e1 && recipe.Element2 == e2) ||
						(recipe.Element1 == e2 && recipe.Element2 == e1) {

						if !isValidTierProgression(recipe, resultElement, db) {
							continue
						}
						//Mark this specific combination->result as visited
						visitedCombinations[combinationKey] = true

						//Create Recipe with result
						recipeWithResult := model.Recipe{
							Element1: recipe.Element1,
							Element2: recipe.Element2,
							Result:   resultElementName,
						}

						newPath := make([]model.Recipe, len(node.Path)+1)
						copy(newPath, node.Path)
						newPath[len(node.Path)] = recipeWithResult

						newNode := &BFSNode{
							Element:    resultElementName,
							Path:       newPath,
							ParentNode: node,
						}
						queue.PushBack(newNode)

						if !discoveredElements[resultElementName] {
							discoveredElements[resultElementName] = true
						}
						break
					}
				}
			}
		}
	}

	//Apply iterative expansion to all paths
	for i := range paths {
		paths[i] = iterativeExpansion(paths[i], db, startElements, progress)
	}

	result <- &BFSResult{
		TargetElement: targetElement,
		Paths:         paths,
		VisitedNodes:  visitedCount,
	}
	resultSent = true
}

func BFSSingle(db *model.ElementsDatabase, startElements []string, targetElement string,
	result chan<- *BFSResult, progress chan<- *SearchProgress) {
	BFSWithOptions(db, startElements, targetElement, []int{}, 1, result, progress)
}

func BFSMultipleThreaded(db *model.ElementsDatabase, startElements []string,
	//Init
	targetElement string, maxPaths int, timeoutSeconds int,
	result chan<- *BFSResult) {
	numWorkers := runtime.NumCPU()
	pathsChan := make(chan []model.Recipe, maxPaths*2)
	tasks := make(chan BFSTask, 100)
	done := make(chan bool, 1)
	var mu sync.Mutex
	collectedPaths := make([][]model.Recipe, 0, maxPaths)
	totalVisited := 0
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSeconds)*time.Second)
	defer cancel()
	//Threading Mumbo Jumbo
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
					log.Printf("Found path %d/%d with %d steps", len(collectedPaths), maxPaths, len(path))
				}
				mu.Unlock()

			case <-ctx.Done():
				done <- true
				return
			}
		}
	}()

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

					log.Printf("Worker %d: Processing task with shuffle %v, banned %v",
						workerID, task.Shuffle, task.BannedRecipes)

					searchCtx, searchCancel := context.WithTimeout(ctx, 15*time.Second)
					resultChan := make(chan *BFSResult, 1)

					go BFSWithOptions(db, task.Shuffle, targetElement, task.BannedRecipes,
						1, resultChan, progressChan)

					select {
					case bfsResult := <-resultChan:
						searchCancel()

						mu.Lock()
						totalVisited += bfsResult.VisitedNodes
						mu.Unlock()

						if len(bfsResult.Paths) > 0 {
							select {
							case pathsChan <- bfsResult.Paths[0]:
							case <-ctx.Done():
								return
							}
						}

					case <-searchCtx.Done():
						searchCancel()
						log.Printf("Worker %d: Search timeout", workerID)

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

	go func() {
		defer close(tasks)

		shuffles := generateFixedShuffles(startElements)

		targetElem, exists := db.Elements[targetElement]
		targetRecipeCount := 0
		if exists {
			targetRecipeCount = len(targetElem.Recipes)
		}

		for _, shuffle := range shuffles {
			mu.Lock()
			currentCount := len(collectedPaths)
			mu.Unlock()

			if currentCount >= maxPaths {
				break
			}

			task := BFSTask{
				Shuffle:       shuffle,
				BannedRecipes: []int{},
			}

			select {
			case tasks <- task:
			case <-ctx.Done():
				return
			}
		}

		time.Sleep(500 * time.Millisecond)

		bannedRecipes := []int{}
		for recipeIdx := 0; recipeIdx < targetRecipeCount; recipeIdx++ {
			mu.Lock()
			currentCount := len(collectedPaths)
			mu.Unlock()

			if currentCount >= maxPaths {
				break
			}

			bannedRecipes = append(bannedRecipes, recipeIdx)
			log.Printf("Starting phase 2: Banning recipe %d", recipeIdx)

			task := BFSTask{
				Shuffle:       startElements,
				BannedRecipes: append([]int{}, bannedRecipes...),
			}

			select {
			case tasks <- task:
			case <-ctx.Done():
				return
			}
			for j := 0; j < min(3, len(shuffles)); j++ {
				mu.Lock()
				currentCount := len(collectedPaths)
				mu.Unlock()

				if currentCount >= maxPaths {
					break
				}

				task := BFSTask{
					Shuffle:       shuffles[j],
					BannedRecipes: append([]int{}, bannedRecipes...),
				}

				select {
				case tasks <- task:
				case <-ctx.Done():
					return
				}
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

	result <- &BFSResult{
		TargetElement: targetElement,
		Paths:         finalPaths,
		VisitedNodes:  finalVisited,
	}
	close(result)
}

type BFSTask struct {
	Shuffle       []string
	BannedRecipes []int
}

func generateFixedShuffles(elements []string) [][]string {
	if len(elements) != 4 {
		return [][]string{elements}
	}

	perms := [][]string{}
	a, b, c, d := elements[0], elements[1], elements[2], elements[3]

	perms = append(perms, []string{a, b, c, d})
	perms = append(perms, []string{a, b, d, c})
	perms = append(perms, []string{a, c, b, d})
	perms = append(perms, []string{a, c, d, b})
	perms = append(perms, []string{a, d, b, c})
	perms = append(perms, []string{a, d, c, b})
	perms = append(perms, []string{b, a, c, d})
	perms = append(perms, []string{b, a, d, c})
	perms = append(perms, []string{b, c, a, d})
	perms = append(perms, []string{b, c, d, a})
	perms = append(perms, []string{b, d, a, c})
	perms = append(perms, []string{b, d, c, a})
	perms = append(perms, []string{c, a, b, d})
	perms = append(perms, []string{c, a, d, b})
	perms = append(perms, []string{c, b, a, d})
	perms = append(perms, []string{c, b, d, a})
	perms = append(perms, []string{c, d, a, b})
	perms = append(perms, []string{c, d, b, a})
	perms = append(perms, []string{d, a, b, c})
	perms = append(perms, []string{d, a, c, b})
	perms = append(perms, []string{d, b, a, c})
	perms = append(perms, []string{d, b, c, a})
	perms = append(perms, []string{d, c, a, b})
	perms = append(perms, []string{d, c, b, a})

	return perms
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func isDuplicatePath(newPath []model.Recipe, existingPaths [][]model.Recipe) bool {
	if len(existingPaths) == 0 {
		return false
	}

	newRecipes := make(map[string]bool)
	for _, recipe := range newPath {
		e1, e2 := recipe.Element1, recipe.Element2
		if e1 > e2 {
			e1, e2 = e2, e1
		}
		//Use the Result from Recipe
		key := fmt.Sprintf("%s+%s->%s", e1, e2, recipe.Result)
		newRecipes[key] = true
	}

	for _, existingPath := range existingPaths {
		if len(existingPath) != len(newPath) {
			continue
		}

		matches := true
		existingRecipes := make(map[string]bool)
		for _, recipe := range existingPath {
			e1, e2 := recipe.Element1, recipe.Element2
			if e1 > e2 {
				e1, e2 = e2, e1
			}
			//Use the Result from Recipe
			key := fmt.Sprintf("%s+%s->%s", e1, e2, recipe.Result)
			existingRecipes[key] = true
		}

		for key := range newRecipes {
			if !existingRecipes[key] {
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

func Driver(db *model.ElementsDatabase, targetElement string, maxPaths int, step chan<- *SearchProgress) *BFSResult {
	sortedDb := utility.SortByTier(db)
	result := make(chan *BFSResult, 1)
	startElement := []string{"Air", "Water", "Fire", "Earth"}
	//Run BFS in a goroutine
	if maxPaths == 1 {
		go BFSSingle(sortedDb, startElement, targetElement, result, step)
	} else if maxPaths > 1 {
		go BFSMultipleThreaded(sortedDb, startElement, targetElement, maxPaths, 10, result)
	} else {
		result <- &BFSResult{
			TargetElement: targetElement,
			Paths:         [][]model.Recipe{},
			VisitedNodes:  0,
		}
		close(result)
		return nil
	}
	//Wait for result
	results := <-result

	return results
}

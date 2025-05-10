package scrapper

import (
	"backend/src/model"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const DEFAULT_ELEMENTS_PATH = "src/data/elements.json"
const IMAGE_DIR = "src/data/images"

func downloadImage(url, name string) (string, error) {
	resp, err := http.Get(url)

	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	ext := filepath.Ext(url)
	if ext == "" {
		ext = ".png"
	}

	filename := strings.ReplaceAll(strings.ToLower(name), " ", "_") + ext
	path := filepath.Join(IMAGE_DIR, filename)

	out, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", err
	}

	return filename, nil
}

func RunScrapperAndSave() error {
	url := "https://little-alchemy.fandom.com/wiki/Elements_(Little_Alchemy_2)"

	res, err := http.Get(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return fmt.Errorf("status error: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return err
	}

	// Make sure data/images directory exists
	err = os.MkdirAll(IMAGE_DIR, os.ModePerm)
	if err != nil {
		return err
	}

	elements := make(map[string]model.ScrapeElement)

	currentTier := "Unknown"

	// Define all valid tiers for validation
	validTiers := map[string]bool{
		"Starting elements": true,
		"Special element":   true,
	}

	for i := 1; i <= 15; i++ {
		validTiers[fmt.Sprintf("Tier %d Elements", i)] = true
	}

	doc.Find("h3,table.list-table").Each(func(i int, s *goquery.Selection) {
		// get the tier from the h3 tag
		if goquery.NodeName(s) == "h3" {
			header := s.Find(".mw-headline").Text()
			header = strings.TrimSpace(header)

			// Validate if this is a recognized tier header
			if validTiers[header] {
				currentTier = header
				log.Printf("Processing %s", currentTier)
			} else {
				// Handle potential partial matches (just in case HTML structure changes)
				if strings.Contains(header, "Starting") {
					currentTier = "Starting elements"
					log.Printf("Processing %s", currentTier)
				} else if strings.Contains(header, "Special") {
					currentTier = "Special element"
					log.Printf("Processing %s", currentTier)
				} else if strings.Contains(header, "Tier") {
					// Extract tier number if possible
					currentTier = header
					log.Printf("Processing %s", currentTier)
				} else {
					log.Printf("Unknown section: %s (continuing with previous tier)", header)
				}
			}
		} else if goquery.NodeName(s) == "table" {
			s.Find("tr").Each(func(j int, row *goquery.Selection) {
				// Skip header rows
				if j == 0 && row.Find("th").Length() > 0 {
					return
				}

				cells := row.Find("td")
				if cells.Length() >= 2 {
					elementCell := cells.Eq(0)
					recipeCell := cells.Eq(1)

					// Extract element name, clean up whitespace and line breaks
					// The element name is inside the <a> tag after the image
					var name string
					elementLink := elementCell.Find("a").Last()
					if elementLink.Length() > 0 {
						name = elementLink.Text()
					} else {
						// Fallback to the cell text if no link found
						name = elementCell.Text()
					}
					name = strings.TrimSpace(name)

					// Skip empty names or header rows
					if name == "" || name == "Element" {
						return
					}

					imgTag := elementCell.Find("img")
					imgURL, _ := imgTag.Attr("data-src")

					imgURL = strings.TrimSpace(imgURL)
					if imgURL != "" && strings.HasPrefix(imgURL, "//") {
						imgURL = "https:" + imgURL
					}

					var localImage string
					if imgURL != "" {
						localImage, err = downloadImage(imgURL, name)
						if err != nil {
							log.Printf("failed to download image for %s: %v", name, err)
						}
					}

					recipeText := recipeCell.Text()

					// Process recipes
					var combos [][2]string
					recipeCell.Find("li").Each(func(k int, recipe *goquery.Selection) {
						// Extract recipe components from links or plain text
						var ingredients []string
						recipe.Find("a").Each(func(l int, ingredient *goquery.Selection) {
							ingredientName := strings.TrimSpace(ingredient.Text())
							if ingredientName != "" && ingredientName != "+" {
								ingredients = append(ingredients, ingredientName)
							}
						})

						// If we couldn't extract from links, try parsing the text
						if len(ingredients) != 2 {
							comboText := recipe.Text()
							parts := strings.Split(comboText, "+")
							if len(parts) == 2 {
								ingredients = []string{
									strings.TrimSpace(parts[0]),
									strings.TrimSpace(parts[1]),
								}
							}
						}

						// Add valid recipe
						if len(ingredients) == 2 && ingredients[0] != "" && ingredients[1] != "" {
							combos = append(combos, [2]string{ingredients[0], ingredients[1]})
						}
					})

					// Handle special case for starting elements (no combinations)
					if len(combos) == 0 && recipeText != "" && !strings.Contains(recipeText, "+") {
						// This is likely a starting element or special case
						recipeText = strings.TrimSpace(recipeText)
						if recipeText != "" {
							log.Printf("Element %s has description: %s", name, recipeText)
						}
					}

					if name != "" {
						existingElement, exists := elements[name]
						if exists {
							// If the element already exists, append new combinations
							existingElement.Combos = append(existingElement.Combos, combos...)
							// Keep the image if it wasn't downloaded this time
							if localImage != "" {
								existingElement.Image = localImage
							}
							existingElement.Tier = currentTier
							elements[name] = existingElement
						} else {
							// Create new element
							elements[name] = model.ScrapeElement{
								Combos: combos,
								Image:  localImage,
								Tier:   currentTier,
							}
						}
					}
				}
			})
		}

		// table.Find("tr").Each(func(j int, row *goquery.Selection) {
		// 	cells := row.Find("td")
		// 	if cells.Length() >= 2 {
		// 		elementCell := cells.Eq(0)
		// 		recipeCell := cells.Eq(1)

		// 		name := strings.TrimSpace(elementCell.Text())

		// 		imgTag := elementCell.Find("img")
		// 		imgURL, _ := imgTag.Attr("data-src")

		// 		imgURL = strings.TrimSpace(imgURL)
		// 		if imgURL != "" && strings.HasPrefix(imgURL, "//") {
		// 			imgURL = "https:" + imgURL
		// 		}

		// 		var localImage string
		// 		if imgURL != "" {
		// 			localImage, err = downloadImage(imgURL, name)
		// 			if err != nil {
		// 				log.Printf("failed to download image for %s: %v", name, err)
		// 			}
		// 		}

		// 		recipeText := strings.TrimSpace(recipeCell.Text())
		// 		combosText := strings.Split(recipeText, "\n")
		// 		var combos [][2]string

		// 		for _, combo := range combosText {
		// 			parts := strings.Split(combo, "+")
		// 			if len(parts) == 2 {
		// 				left := strings.TrimSpace(parts[0])
		// 				right := strings.TrimSpace(parts[1])
		// 				if left != "" && right != "" {
		// 					combos = append(combos, [2]string{left, right})
		// 				}
		// 			}
		// 		}

		// 		if name != "" {
		// 			elements[name] = model.ScrapeElement{
		// 				Combos: combos,
		// 				Image:  localImage,
		// 			}
		// 		}
		// 	}
		// })
	})

	file, err := os.Create(DEFAULT_ELEMENTS_PATH)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(elements); err != nil {
		return err
	}

	log.Println("Scrape and save successful!")

	tierCounts := make(map[string]int)
	for _, element := range elements {
		tierCounts[element.Tier]++
	}

	log.Printf("Scrape and save successful! Processed %d elements across multiple tiers", len(elements))
	log.Println("Elements by tier:")

	// Track total elements
	totalElements := 0

	// Ordered tiers for nice output
	orderedTiers := []string{"Starting elements", "Special element"}
	for i := 1; i <= 15; i++ {
		orderedTiers = append(orderedTiers, fmt.Sprintf("Tier %d elements", i))
	}

	// Print counts by tier in order
	for _, tier := range orderedTiers {
		count := tierCounts[tier]
		if count > 0 {
			log.Printf("  - %s: %d elements", tier, count)
			totalElements += count
		}
	}

	// Check for any elements in unknown tiers
	if tierCounts["Unknown"] > 0 {
		log.Printf("  - Unknown tier: %d elements", tierCounts["Unknown"])
		totalElements += tierCounts["Unknown"]
	}

	log.Printf("Total elements: %d", totalElements)

	// Save tier information to a separate file
	tiersFile, err := os.Create("src/data/tiers.json")
	if err != nil {
		log.Printf("Warning: Failed to create tiers file: %v", err)
	} else {
		defer tiersFile.Close()

		// Create a map of tier -> element names
		tierElements := make(map[string][]string)
		for name, element := range elements {
			tierElements[element.Tier] = append(tierElements[element.Tier], name)
		}

		encoder := json.NewEncoder(tiersFile)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(tierElements); err != nil {
			log.Printf("Warning: Failed to save tiers data: %v", err)
		} else {
			log.Println("Tiers data saved to src/data/tiers.json")
		}
	}

	// Generate graph data for visualization
	// if err := saveElementsGraph(elements); err != nil {
	// 	log.Printf("Warning: Failed to save elements graph: %v", err)
	// }

	return nil
}

// func saveElementsGraph(elements map[string]model.ScrapeElement) error {
// 	type Node struct {
// 		ID    string `json:"id"`
// 		Label string `json:"label"`
// 		Tier  string `json:"tier"`
// 	}

// 	type Edge struct {
// 		Source      string `json:"source"`
// 		Target      string `json:"target"`
// 		Combination string `json:"combination"`
// 	}

// 	type Graph struct {
// 		Nodes []Node `json:"nodes"`
// 		Edges []Edge `json:"edges"`
// 	}

// 	graph := Graph{
// 		Nodes: []Node{},
// 		Edges: []Edge{},
// 	}

// 	// Create a unique set of nodes
// 	nodesMap := make(map[string]bool)

// 	// First, add all elements as nodes
// 	for name, element := range elements {
// 		if !nodesMap[name] {
// 			graph.Nodes = append(graph.Nodes, Node{
// 				ID:    name,
// 				Label: name,
// 				Tier:  element.Tier,
// 			})
// 			nodesMap[name] = true
// 		}

// 		// Add edges for each combination
// 		for _, combo := range element.Combos {
// 			// Ensure both ingredients exist as nodes
// 			for _, ingredient := range combo {
// 				if !nodesMap[ingredient] {
// 					// If we don't have tier info for an ingredient, mark it as "Unknown"
// 					tier := "Unknown"
// 					if elem, exists := elements[ingredient]; exists {
// 						tier = elem.Tier
// 					}

// 					graph.Nodes = append(graph.Nodes, Node{
// 						ID:    ingredient,
// 						Label: ingredient,
// 						Tier:  tier,
// 					})
// 					nodesMap[ingredient] = true
// 				}
// 			}

// 			// Add the edge
// 			graph.Edges = append(graph.Edges, Edge{
// 				Source:      combo[0],
// 				Target:      combo[1],
// 				Combination: name,
// 			})
// 		}
// 	}

// 	// Save the graph data
// 	graphFile, err := os.Create("src/data/elements_graph.json")
// 	if err != nil {
// 		return err
// 	}
// 	defer graphFile.Close()

// 	encoder := json.NewEncoder(graphFile)
// 	encoder.SetIndent("", "  ")
// 	if err := encoder.Encode(graph); err != nil {
// 		return err
// 	}
// 	log.Println("Elements graph saved to src/data/elements_graph.json")
// 	return nil
// }

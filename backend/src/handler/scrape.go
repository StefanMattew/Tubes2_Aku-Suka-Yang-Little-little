package handler

import (
	"backend/src/scrapper"
	"log"
)

func main() {
	err := scrapper.RunScrapperAndSave()
	if err != nil {
		log.Fatalf("Scraper failed: %v", err)
	}
}

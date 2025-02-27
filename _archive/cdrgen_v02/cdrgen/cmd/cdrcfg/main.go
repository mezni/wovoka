package main

import (
	"log"

	"github.com/mezni/wovoka/cdrgen/application/services"
)

const (
	dbFile       = "./cdrgen.db"
	yamlFilename = "config.yaml"
)

func main() {

	// Create a new LoaderService instance
	loaderService, err := services.NewLoaderService(dbFile)
	if err != nil {
		log.Fatalf("Error initializing loader service: %v", err)
	}
	defer func() {
		// Ensure the database connection is closed when done
		if err := loaderService.Close(); err != nil {
			log.Printf("Error closing loader service: %v", err)
		}
	}()

	// Load data into the database
	if err := loaderService.Load(yamlFilename); err != nil {
		log.Fatalf("Error loading data: %v", err)
	}

	log.Println("Data loading process completed successfully.")
}

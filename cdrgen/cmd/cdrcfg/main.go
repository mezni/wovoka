package main

import (
	"flag"
	"fmt"
	"os"
	"log"
	"github.com/mezni/wovoka/cdrgen/domain/entities"
	"github.com/mezni/wovoka/cdrgen/infrastructure/inmem"
	"github.com/mezni/wovoka/cdrgen/infrastructure/bolt"
	"github.com/mezni/wovoka/cdrgen/domain/services"
)

// Version will be set during build time using -ldflags
var version = "development" // default version

func main() {
	// Define flags
	configFile := flag.String("f", "", "Path to the config file (required)")
	databaseName := flag.String("d", "", "Name of the BoltDB database (required)")
	showVersion := flag.Bool("v", false, "Show the version of the tool")
	showHelp := flag.Bool("h", false, "Show help information")

	// Customize flag usage
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}

	// Parse flags
	flag.Parse()

	// Handle help flag
	if *showHelp {
		flag.Usage()
		os.Exit(0)
	}

	// Handle version flag
	if *showVersion {
		fmt.Printf("%s version %s\n", os.Args[0], version)
		os.Exit(0)
	}

	// Validate required flags
	if *configFile == "" || *databaseName == "" {
		fmt.Fprintln(os.Stderr, "Error: Both -f (config file) and -d (database name) are required.")
		flag.Usage()
		os.Exit(1)
	}

	// Validate config file existence
	if _, err := os.Stat(*configFile); os.IsNotExist(err) {
		log.Fatalf("Error: Config file '%s' does not exist.\n", *configFile)
	}

	var repo repositories.LocationRepository
	useBolt := true // Change to false for in-memory

	if useBolt {
		// Using BoltDB repository
		dbName := "locations.db"
		repo, _ = boltstore.NewBoltDBLocationRepository(dbName)
		defer repo.Close()
	} else {
		// Using in-memory repository
		repo = inmemorystore.NewInMemoryLocationRepository()
	}

	// Load the location configuration file
	configFilePath := "locations.json"
	locationService, err := services.NewLocationService(configFilePath, repo)
	if err != nil {
		log.Fatalf("Error creating location service: %v", err)
	}

	// Generate locations based on the configuration
	locations, err := locationService.GenerateLocations()
	if err != nil {
		log.Fatalf("Error generating locations: %v", err)
	}

	// Output the generated locations
	fmt.Println("Generated Locations:")
	for _, location := range locations {
		fmt.Printf("ID: %d, Name: %s, NetworkType: %s, Lat: %.4f - %.4f, Lon: %.4f - %.4f\n",
			location.LocationID, location.LocationName, location.NetworkType.String(),
			location.LatMin, location.LatMax, location.LonMin, location.LonMax)
	}
}

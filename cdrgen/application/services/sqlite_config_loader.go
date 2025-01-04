package services

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/mezni/wovoka/cdrgen/application/interfaces"
	"github.com/mezni/wovoka/cdrgen/domain/entities"
	"github.com/mezni/wovoka/cdrgen/infrastructure/sqlitestore"
)

type LoaderService struct {
	DB                  *sql.DB
	NetworkTechRepo     *sqlitestore.NetworkTechnologyRepository
	NetworkElementTypeRepo  *sqlitestore.NetworkElementTypeRepository
}

// NewLoaderService initializes the LoaderService with all repositories.
func NewLoaderService(dbFile string) (*LoaderService, error) {
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		return nil, fmt.Errorf("could not open database: %v", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("could not connect to database: %v", err)
	}

	return &LoaderService{
		DB:                  db,
		NetworkTechRepo:     sqlitestore.NewNetworkTechnologyRepository(db),
		NetworkElementTypeRepo:  sqlitestore.NewNetworkElementTypeRepository(db),
	}, nil
}

// SetupDatabase creates all necessary tables.
func (l *LoaderService) SetupDatabase() error {
	if err := l.NetworkTechRepo.CreateTable(); err != nil {
		return fmt.Errorf("failed to create network technology table: %v", err)
	}
	if err := l.NetworkElementTypeRepo.CreateTable(); err != nil {
		return fmt.Errorf("failed to create network element type table: %v", err)
	}
	log.Println("All tables created successfully.")
	return nil
}

// LoadNetworkTechnologies processes and inserts NetworkTechnologies into the database.
func (l *LoaderService) LoadNetworkTechnologies(networkTechnologies []interfaces.NetworkTechnology) error {
	tx, err := l.DB.Begin()
	if err != nil {
		return fmt.Errorf("could not begin transaction: %v", err)
	}

	for _, nt := range networkTechnologies {
		entity := entities.NetworkTechnology{
			Name:        nt.Name,
			Description: nt.Description,
		}
		if err := l.NetworkTechRepo.Insert(entity); err != nil {
			tx.Rollback()
			return fmt.Errorf("error saving network technology %s: %v", entity.Name, err)
		}
		log.Printf("Successfully inserted network technology: %s", entity.Name)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("could not commit transaction: %v", err)
	}

	log.Printf("Successfully inserted %d network technologies", len(networkTechnologies))
	return nil
}

// LoadNetworkElementTypes processes and inserts NetworkElementTypes into the database.
func (l *LoaderService) LoadNetworkElementTypes(networkElementTypes []interfaces.NetworkElementType) error {
	tx, err := l.DB.Begin()
	if err != nil {
		return fmt.Errorf("could not begin transaction: %v", err)
	}

	for _, ne := range networkElementTypes {
		entity := entities.NetworkElementType{
			Name:              ne.Name,
			Description:       ne.Description,
			NetworkTechnology: ne.NetworkTechnology,
		}
		if err := l.NetworkElementTypeRepo.Insert(entity); err != nil {
			tx.Rollback()
			return fmt.Errorf("error saving network element type %s: %v", entity.Name, err)
		}
		log.Printf("Successfully inserted network element type: %s", entity.Name)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("could not commit transaction: %v", err)
	}

	log.Printf("Successfully inserted %d network element types", len(networkElementTypes))
	return nil
}

// Load reads JSON data and loads them into the database.
func (l *LoaderService) Load() error {
	// Step 1: Setup the database (create necessary tables)
	if err := l.SetupDatabase(); err != nil {
		return fmt.Errorf("failed to set up database tables: %v", err)
	}

	// Step 2: Read JSON data from the file
	jsonData, err := interfaces.ReadConfig()
	if err != nil {
		return fmt.Errorf("could not read JSON file: %v", err)
	}

	// Step 3: Load Network Technologies
	log.Printf("Started loading %d network technologies", len(jsonData.NetworkTechnologies))
	if err := l.LoadNetworkTechnologies(jsonData.NetworkTechnologies); err != nil {
		return fmt.Errorf("failed to load network technologies: %v", err)
	}

	// Step 4: Load Network Element Types
	log.Printf("Started loading %d network element types", len(jsonData.NetworkElementTypes))
	if err := l.LoadNetworkElementTypes(jsonData.NetworkElementTypes); err != nil {
		return fmt.Errorf("failed to load network element types: %v", err)
	}

	return nil
}

// Close closes the database connection.
func (l *LoaderService) Close() error {
	if l.DB != nil {
		err := l.DB.Close()
		if err != nil {
			log.Printf("Error closing database connection: %v", err)
			return err
		}
		log.Println("Database connection closed successfully.")
	}
	return nil
}

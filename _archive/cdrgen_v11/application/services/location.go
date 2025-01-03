package services

import (
    "encoding/json"
    "fmt"
    "math/rand"
    "time"

    "go.etcd.io/bbolt"
    "github.com/mezni/wovoka/cdrgen/domain/entities"
    "github.com/mezni/wovoka/cdrgen/application/dtos"
)

// Define a constant slice for valid network technologies
var validNetworkTechnologies = []string{"2G", "3G", "4G", "5G"}

// IsValidNetworkTechnology checks if the given network technology is valid.
func IsValidNetworkTechnology(networkTechnology string) bool {
    for _, validTech := range validNetworkTechnologies {
        if networkTechnology == validTech {
            return true
        }
    }
    return false
}

// LocationService handles logic for generating and managing locations.
type LocationService struct {
    db *bbolt.DB // BoltDB database instance
}

// NewLocationService initializes a LocationService with a provided BoltDB instance.
func NewLocationService(db *bbolt.DB) *LocationService {
    return &LocationService{
        db: db,
    }
}

// GenerateLatitudeRanges generates latitude ranges for each row.
func (s *LocationService) generateLatitudeRanges(minLat, maxLat float64, rows int) [][2]float64 {
    step := (maxLat - minLat) / float64(rows)
    latitudeRanges := make([][2]float64, rows)

    for i := 0; i < rows; i++ {
        latitudeRanges[i][0] = minLat + float64(i)*step     // minLatitude for the row
        latitudeRanges[i][1] = minLat + float64(i+1)*step   // maxLatitude for the row
    }

    return latitudeRanges
}

// GenerateLongitudeRanges generates longitude ranges for each column.
func (s *LocationService) generateLongitudeRanges(minLong, maxLong float64, cols int) [][2]float64 {
    step := (maxLong - minLong) / float64(cols)
    longitudeRanges := make([][2]float64, cols)

    for i := 0; i < cols; i++ {
        longitudeRanges[i][0] = minLong + float64(i)*step     // minLongitude for the column
        longitudeRanges[i][1] = minLong + float64(i+1)*step   // maxLongitude for the column
    }

    return longitudeRanges
}

// GenerateAreaCode generates a 5-digit area code based on the network technology prefix.
func (s *LocationService) generateAreaCode(prefix int) int {
    return prefix*10000 + rand.Intn(10000) // Ensure the AreaCode is always 5 digits
}

// GetNetworkTechnologyPrefix returns the prefix based on the network technology.
func getNetworkTechnologyPrefix(networkTechnology string) int {
    var prefix int
    switch networkTechnology {
    case "2G":
        prefix = 2
    case "3G":
        prefix = 3
    case "4G":
        prefix = 4
    case "5G":
        prefix = 5
    default:
        prefix = 0 // Default for unknown technologies
    }
    return prefix
}

// GenerateLocations processes the entire CfgData struct and generates locations.
func (s *LocationService) GenerateLocations(cfgData dtos.CfgData) ([]entities.Location, error) {
    // Seed the random number generator
    rand.Seed(time.Now().UnixNano())

    var locations []entities.Location
    locationID := 1

    // Process the data from the cfgData struct
    minLat, maxLat := cfgData.Locations.Latitude[0], cfgData.Locations.Latitude[1]
    minLong, maxLong := cfgData.Locations.Longitude[0], cfgData.Locations.Longitude[1]

    // Iterate through the LocationSplits
    for _, locationSplit := range cfgData.Locations.LocationSplit {
        // Validate NetworkTechnology using IsValidNetworkTechnology function
        if !IsValidNetworkTechnology(locationSplit.NetworkTechnology) {
            return nil, fmt.Errorf("invalid NetworkTechnology: %s, must be one of: %v", locationSplit.NetworkTechnology, validNetworkTechnologies)
        }

        fmt.Println("Processing NetworkTechnology:", locationSplit.NetworkTechnology)

        latitudeRanges := s.generateLatitudeRanges(minLat, maxLat, locationSplit.SplitRows)
        longitudeRanges := s.generateLongitudeRanges(minLong, maxLong, locationSplit.SplitColumns)

        // Get the prefix based on the network technology
        prefix := getNetworkTechnologyPrefix(locationSplit.NetworkTechnology)

        // Combine rows and columns to generate Location entries
        for i, latRange := range latitudeRanges {
            for j, longRange := range longitudeRanges {
                // Determine the name based on LocationNames and grid position
                var name string
                if i*locationSplit.SplitColumns+j < len(locationSplit.LocationNames) {
                    name = locationSplit.LocationNames[i*locationSplit.SplitColumns+j]
                } else {
                    name = fmt.Sprintf("Unnamed-%d", locationID)
                }

                // Generate AreaCode
                areaCode := s.generateAreaCode(prefix)

                // Use NewLocation to create a Location entry with original latitudes and longitudes
                location, err := entities.NewLocation(locationID, locationSplit.NetworkTechnology, name, latRange[0], latRange[1], longRange[0], longRange[1], areaCode)
                if err != nil {
                    return nil, fmt.Errorf("failed to create location: %v", err)
                }

                // Append to the locations slice
                locations = append(locations, *location)

                // Increment LocationID
                locationID++
            }
        }
    }

    return locations, nil
}

// WithBoltDB allows actions to be performed with the BoltDB instance in a transaction.
func (s *LocationService) WithBoltDB(action func(tx *bbolt.Tx) error) error {
    // Start a write transaction for BoltDB
    return s.db.Update(func(tx *bbolt.Tx) error {
        return action(tx)
    })
}

// SaveToBoltDB saves the generated locations to BoltDB using a transaction.
func (s *LocationService) SaveToBoltDB(dbName, bucketName string, locations []entities.Location) error {
    return s.WithBoltDB(func(tx *bbolt.Tx) error {
        // Create a bucket for storing locations (if it doesn't exist)
        bucket, err := tx.CreateBucketIfNotExists([]byte(bucketName))
        if err != nil {
            return fmt.Errorf("failed to create bucket: %v", err)
        }

        // Iterate through the locations and save them to the bucket
        for _, location := range locations {
            // Marshal the location into JSON for storage
            locationData, err := json.Marshal(location)
            if err != nil {
                return fmt.Errorf("failed to marshal location: %v", err)
            }

            // Use the Location ID as the key, and the JSON data as the value
            err = bucket.Put([]byte(fmt.Sprintf("%d", location.ID)), locationData)
            if err != nil {
                return fmt.Errorf("failed to save location to bucket: %v", err)
            }
        }
        return nil
    })
}

// ReadFromBoltDB reads locations from the specified BoltDB database and bucket.
func (s *LocationService) ReadFromBoltDB(bucketName string) ([]entities.Location, error) {
    var locations []entities.Location

    err := s.WithBoltDB(func(tx *bbolt.Tx) error {
        // Access the bucket where locations are stored
        bucket := tx.Bucket([]byte(bucketName))
        if bucket == nil {
            return fmt.Errorf("bucket %s not found", bucketName)
        }

        // Iterate through all keys (location IDs) in the bucket
        return bucket.ForEach(func(k, v []byte) error {
            // Unmarshal the location data from JSON
            var location entities.Location
            if err := json.Unmarshal(v, &location); err != nil {
                return fmt.Errorf("failed to unmarshal location: %v", err)
            }

            // Append the location to the locations slice
            locations = append(locations, location)
            return nil
        })
    })

    if err != nil {
        return nil, fmt.Errorf("failed to read locations from BoltDB: %v", err)
    }

    return locations, nil
}

package factories

import (
	"fmt"
	"github.com/mezni/wovoka/cdrgen/application/mappers"
	"github.com/mezni/wovoka/cdrgen/domain/entities"
)

// GenerateLocations generates locations based on the provided configuration.
func GenerateLocations(config *mappers.Config) ([]*entities.Location, error) {
	var locations []*entities.Location
	locationID := 1

	for networkType, networkData := range config.Networks {
		latRange := config.Coordinates.Latitude[1] - config.Coordinates.Latitude[0]
		lonRange := config.Coordinates.Longitude[1] - config.Coordinates.Longitude[0]
		latStep := latRange / float64(networkData.Rows)
		lonStep := lonRange / float64(networkData.Columns)

		if len(networkData.LocationNames) != networkData.Rows*networkData.Columns {
			return nil, fmt.Errorf("mismatch between location names and grid dimensions for network %s", networkType)
		}

		// Generate grid locations for the network
		index := 0
		for row := 0; row < networkData.Rows; row++ {
			for col := 0; col < networkData.Columns; col++ {
				latMin := config.Coordinates.Latitude[0] + latStep*float64(row)
				latMax := latMin + latStep
				lonMin := config.Coordinates.Longitude[0] + lonStep*float64(col)
				lonMax := lonMin + lonStep

				locationName := networkData.LocationNames[index]
				index++

				// Create the location entity with the locationID and formatted AreaCode
				location := &entities.Location{
					ID:                locationID,
					NetworkTechnology: networkType,
					Name:              locationName,
					LatitudeMin:       latMin,
					LatitudeMax:       latMax,
					LongitudeMin:      lonMin,
					LongitudeMax:      lonMax,
					AreaCode:          fmt.Sprintf("%04d", locationID), // Formatting locationID as a 4-digit string
				}

				locations = append(locations, location)
				locationID++
			}
		}
	}

	return locations, nil
}

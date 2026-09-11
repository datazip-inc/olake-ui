package dto

import (
	"fmt"

	"github.com/datazip-inc/olake-ui/server/internal/constants"
)

// ValidateSourceType checks if the provided type is in the list of supported source types
func ValidateSourceType(t string) error {
	for _, allowed := range constants.SupportedSourceTypes {
		if t == allowed {
			return nil
		}
	}
	return fmt.Errorf("invalid source type '%s', supported sources are: %v", t, constants.SupportedSourceTypes)
}

// ValidateDestinationType checks if the provided type is in the list of supported destination types
func ValidateDestinationType(t string) error {
	for _, allowed := range constants.SupportedDestinationTypes {
		if t == allowed {
			return nil
		}
	}
	return fmt.Errorf("invalid destination type '%s', supported destinations are: %v", t, constants.SupportedDestinationTypes)
}

// ValidateJobDriverConfig validates source/destination on job create/update requests.
// When id is omitted, name, version, config, and connector type are required.
func ValidateJobDriverConfig(source, destination *DriverConfig) error {
	if source != nil && source.ID == nil {
		if err := ValidateSourceType(source.Type); err != nil {
			return err
		}
		if source.Name == "" || source.Version == "" || source.Config == "" {
			return fmt.Errorf("source name, version, and config are required when source id is not provided")
		}
	}
	if destination != nil && destination.ID == nil {
		if err := ValidateDestinationType(destination.Type); err != nil {
			return err
		}
		if destination.Name == "" || destination.Version == "" || destination.Config == "" {
			return fmt.Errorf("destination name, version, and config are required when destination id is not provided")
		}
	}
	return nil
}

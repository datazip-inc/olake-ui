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

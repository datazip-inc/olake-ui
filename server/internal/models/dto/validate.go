package dto

import (
	"encoding/json"
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

// UnmarshalAndValidate unmarshals JSON from request body into the provided struct
func UnmarshalAndValidate(requestBody []byte, target interface{}) error {
	if err := json.Unmarshal(requestBody, target); err != nil {
		return err
	}
	return Validate(target)
}

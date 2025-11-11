package dto

import (
	"fmt"

	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"github.com/go-playground/validator/v10"
)

// ValidateStruct validates any struct that has `validate` tags.
func Validate(s interface{}) error {
	validate := validator.New()
	err := validate.Struct(s)
	if err != nil {
		if _, ok := err.(*validator.InvalidValidationError); ok {
			return fmt.Errorf("invalid validation: %s", err)
		}

		// collect all validation errors into a single message
		var errorMessages string
		for _, err := range err.(validator.ValidationErrors) {
			errorMessages += fmt.Sprintf("Field '%s' failed validation rule '%s'; ", err.Field(), err.Tag())
		}
		return fmt.Errorf("validation failed: %s", errorMessages)
	}
	return nil
}

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

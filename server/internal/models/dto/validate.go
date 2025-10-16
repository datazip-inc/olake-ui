package dto

import (
	"fmt"

	"github.com/datazip/olake-ui/server/internal/constants"
	"github.com/go-playground/validator/v10"
)

// ValidateStruct validates any struct that has `validate` tags.
func Validate(s interface{}) error {
	validate := validator.New()
	err := validate.Struct(s)
	if err != nil {
		if _, ok := err.(*validator.InvalidValidationError); ok {
			return fmt.Errorf("invalid validation error: %v", err)
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

// ValidateSourceType checks if the provided type is in the list of supported types
func ValidateDriverType(t string) error {
	for _, allowed := range constants.SupportedDriverTypes {
		if t == allowed {
			return nil
		}
	}
	return fmt.Errorf("invalid source type '%s', supported types are: %v", t, constants.SupportedDriverTypes)
}

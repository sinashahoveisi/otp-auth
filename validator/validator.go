package validator

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Validator wraps the go-playground validator
type Validator struct {
	validator *validator.Validate
}

// New creates a new validator instance
func New() *Validator {
	v := validator.New()

	// Register custom tag name function to use json tags for field names
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	// Register custom phone number validator
	v.RegisterValidation("phone_number", validatePhoneNumber)

	return &Validator{
		validator: v,
	}
}

// ValidateStruct validates a struct and returns formatted errors
func (v *Validator) ValidateStruct(s interface{}) error {
	if s == nil {
		return fmt.Errorf("input cannot be nil")
	}

	if err := v.validator.Struct(s); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			var errors []string
			for _, validationErr := range validationErrors {
				errors = append(errors, v.formatFieldError(validationErr))
			}
			return fmt.Errorf("validation failed: %s", strings.Join(errors, "; "))
		}
		// Handle other validation errors (like InvalidValidationError)
		return fmt.Errorf("validation error: %v", err)
	}
	return nil
}

// formatFieldError formats a single field validation error
func (v *Validator) formatFieldError(err validator.FieldError) string {
	field := err.Field()
	tag := err.Tag()
	param := err.Param()

	switch tag {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "min":
		if err.Kind() == reflect.String {
			return fmt.Sprintf("%s must be at least %s characters long", field, param)
		}
		return fmt.Sprintf("%s must be at least %s", field, param)
	case "max":
		if err.Kind() == reflect.String {
			return fmt.Sprintf("%s must be at most %s characters long", field, param)
		}
		return fmt.Sprintf("%s must be at most %s", field, param)
	case "len":
		return fmt.Sprintf("%s must be exactly %s characters long", field, param)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "phone":
		return fmt.Sprintf("%s must be a valid phone number", field)
	case "phone_number":
		return fmt.Sprintf("%s must be a valid phone number (format: +1234567890)", field)
	default:
		return fmt.Sprintf("%s is invalid", field)
	}
}

// validatePhoneNumber validates phone number format
// Accepts international format starting with + followed by country code and number
// Examples: +1234567890, +12345678901, +123456789012
func validatePhoneNumber(fl validator.FieldLevel) bool {
	phoneNumber := fl.Field().String()

	// Phone number must start with + and have 7-15 digits total
	phoneRegex := regexp.MustCompile(`^\+[1-9]\d{6,14}$`)

	return phoneRegex.MatchString(phoneNumber)
}

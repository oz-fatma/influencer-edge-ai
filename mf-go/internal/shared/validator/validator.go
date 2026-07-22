package validator

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Validate is the shared validator instance.
var Validate *validator.Validate

func init() {
	Validate = validator.New(validator.WithRequiredStructEnabled())
}

// ValidateStruct validates a struct using validator tags.
func ValidateStruct(s interface{}) error {
	return Validate.Struct(s)
}

// FormatValidationErrors formats validator errors into a human-readable string.
func FormatValidationErrors(err error) string {
	if errs, ok := err.(validator.ValidationErrors); ok {
		var msgs []string
		for _, e := range errs {
			msgs = append(msgs, fmt.Sprintf("field '%s' failed on '%s' tag", e.Field(), e.Tag()))
		}
		return strings.Join(msgs, "; ")
	}
	return err.Error()
}

// DecodeAndValidate decodes a JSON request body into the target and validates it.
func DecodeAndValidate(r *http.Request, target interface{}) error {
	if err := json.NewDecoder(r.Body).Decode(target); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}
	if err := ValidateStruct(target); err != nil {
		return fmt.Errorf("%s", FormatValidationErrors(err))
	}
	return nil
}

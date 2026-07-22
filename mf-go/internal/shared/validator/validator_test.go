package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type testStruct struct {
	Email string `validate:"required,email"`
	Name  string `validate:"required,min=2"`
}

func TestValidateStruct_Valid(t *testing.T) {
	s := testStruct{Email: "test@example.com", Name: "John"}
	err := ValidateStruct(s)
	assert.NoError(t, err)
}

func TestValidateStruct_Invalid(t *testing.T) {
	s := testStruct{Email: "not-an-email", Name: ""}
	err := ValidateStruct(s)
	assert.Error(t, err)
}

func TestFormatValidationErrors(t *testing.T) {
	s := testStruct{Email: "", Name: ""}
	err := ValidateStruct(s)
	msg := FormatValidationErrors(err)
	assert.Contains(t, msg, "Email")
	assert.Contains(t, msg, "Name")
}

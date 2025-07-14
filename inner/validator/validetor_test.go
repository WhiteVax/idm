package validator

import (
	"errors"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"idm/inner/employee"
	"testing"
	"time"
)

func AssertValidationField(t *testing.T, err error, expectedField string) {
	t.Helper()
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		for _, fe := range ve {
			if fe.Field() == expectedField {
				return
			}
		}
		t.Errorf("Expected field '%s' to be invalid, but it wasn't", expectedField)
	} else {
		t.Errorf("Expected validator.ValidationErrors, got %v", err)
	}
}

func TestCreateRequest(t *testing.T) {
	v := validator.New()
	validRequest := employee.CreateRequest{
		Name:      "John",
		Surname:   "Sina",
		Age:       18,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	t.Run("Valid request", func(t *testing.T) {
		err := v.Struct(validRequest)
		assert.Nil(t, err)
	})

	t.Run("Empty name", func(t *testing.T) {
		req := validRequest
		req.Name = ""
		err := v.Struct(req)
		assert.NotNil(t, err)
		AssertValidationField(t, err, "Name")
	})

	t.Run("Too short name", func(t *testing.T) {
		req := validRequest
		req.Name = "E"
		err := v.Struct(req)
		assert.NotNil(t, err)
		AssertValidationField(t, err, "Name")
	})

	t.Run("Empty surname", func(t *testing.T) {
		req := validRequest
		req.Surname = ""
		err := v.Struct(req)
		assert.NotNil(t, err)
		AssertValidationField(t, err, "Surname")
	})

	t.Run("Too short surname", func(t *testing.T) {
		req := validRequest
		req.Surname = "J"
		err := v.Struct(req)
		assert.NotNil(t, err)
		AssertValidationField(t, err, "Surname")
	})

	t.Run("Age is too young", func(t *testing.T) {
		req := validRequest
		req.Age = 14
		err := v.Struct(req)
		assert.NotNil(t, err)
		AssertValidationField(t, err, "Age")
	})

	t.Run("Empty created data", func(t *testing.T) {
		req :=
			employee.CreateRequest{
				Name:      "John",
				Surname:   "Sina",
				Age:       18,
				UpdatedAt: time.Now()}
		err := v.Struct(req)
		assert.NotNil(t, err)
		AssertValidationField(t, err, "CreatedAt")
	})

	t.Run("Empty updated data", func(t *testing.T) {
		req :=
			employee.CreateRequest{
				Name:      "John",
				Surname:   "Sina",
				Age:       18,
				CreatedAt: time.Now()}
		err := v.Struct(req)
		assert.NotNil(t, err)
		AssertValidationField(t, err, "UpdatedAt")
	})
}

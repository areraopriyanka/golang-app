package request

import (
	"process-api/pkg/db/dao"
	"process-api/pkg/validators"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

func TestUpdateCustomerAddressRequest(t *testing.T) {
	t.Run("Missing address in struct", func(t *testing.T) {
		err := validators.InitValidationRules()
		if !assert.NoError(t, err) {
			return
		}

		errors := validators.ValidateStruct(dao.UpdateCustomerAddressRequest{
			StreetAddress: "",
			ApartmentNo:   "",
			ZipCode:       "12345",
			City:          "City",
			State:         "RI",
		})

		switch errors := errors.(type) {
		case validator.ValidationErrors:
			if assert.Len(t, errors, 1, "Expected one error") != true {
				return
			}

			assert.Equal(t, "streetAddress", errors[0].Field())
			assert.Equal(t, "required", errors[0].ActualTag())
		default:
			assert.Fail(t, "Expected error type of ValidationErrors")
		}
	})

	t.Run("Invalid address in struct", func(t *testing.T) {
		err := validators.InitValidationRules()
		if !assert.NoError(t, err) {
			return
		}

		errors := validators.ValidateStruct(dao.UpdateCustomerAddressRequest{
			StreetAddress: "!!!!",
			ApartmentNo:   "!!!!",
			ZipCode:       "12345",
			City:          "City",
			State:         "RI",
		})

		switch errors := errors.(type) {
		case validator.ValidationErrors:
			if assert.Len(t, errors, 2, "Expected two errors") != true {
				return
			}

			assert.Equal(t, "streetAddress", errors[0].Field())
			assert.Equal(t, "validateAddress", errors[0].ActualTag())

			assert.Equal(t, "apartmentNo", errors[1].Field())
			assert.Equal(t, "validateAddress", errors[1].ActualTag())
		default:
			assert.Fail(t, "Expected error type of ValidationErrors")
		}
	})

	t.Run("Valid address in struct", func(t *testing.T) {
		err := validators.InitValidationRules()
		if !assert.NoError(t, err) {
			return
		}

		errors := validators.ValidateStruct(dao.UpdateCustomerAddressRequest{
			StreetAddress: "456 Elm Street",
			ApartmentNo:   "Apt #12",
			ZipCode:       "12345",
			City:          "City",
			State:         "RI",
		})

		assert.Nil(t, errors, "Expected no errors for valid struct")
	})
}

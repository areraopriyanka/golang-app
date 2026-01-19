package handler

import (
	"net/http"
	"process-api/pkg/constant"
	"process-api/pkg/db/dao"
	"process-api/pkg/model/response"
	"process-api/pkg/security"

	"github.com/labstack/echo/v4"
)

// PersonalDetailsResponse represents the response for the personal details endpoint
type PersonalDetailsResponse struct {
	FirstName    string                  `json:"firstName" validate:"required"`
	LastName     string                  `json:"lastName" validate:"required"`
	Email        string                  `json:"email" validate:"required,email"`
	MobileNumber string                  `json:"mobileNumber" validate:"required"`
	Address      response.AddressDetails `json:"address" validate:"required"`
}

// GetPersonalDetails handles the GET request to /account/personal-details
// Returns the user's personal details including name, contact info, and address
// @summary Get Personal Details
// @description Returns the user's personal details including name, contact info, and address
// @tags accounts
// @produce json
// @success 200 {object} PersonalDetailsResponse
// @failure 401 {object} response.ErrorResponse
// @failure 404 {object} response.ErrorResponse
// @failure 500 {object} response.ErrorResponse
// @router /account/personal-details [get]
func GetPersonalDetails(c echo.Context) error {
	// Get user ID from the context (using the security middleware)
	cc, ok := c.(*security.LoggedInRegisteredUserContext)
	if !ok {
		return response.UnauthorizedError("Failed to get user Id from custom context")
	}
	userId := cc.UserId

	// Query the database to get the user's personal details
	user, errResponse := dao.RequireUserWithState(userId, constant.ACTIVE)
	if errResponse != nil {
		return errResponse
	}

	// Format the data into the response struct
	personalDetails := PersonalDetailsResponse{
		FirstName:    user.FirstName,
		LastName:     user.LastName,
		Email:        user.Email,
		MobileNumber: user.MobileNo,
		Address: response.AddressDetails{
			AddressLine1: user.StreetAddress,
			AddressLine2: user.ApartmentNo,
			City:         user.City,
			State:        user.State,
			PostalCode:   user.ZipCode,
		},
	}

	// Return the response as JSON
	return c.JSON(http.StatusOK, personalDetails)
}

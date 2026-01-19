package handler

import (
	"fmt"
	"net/http"
	"process-api/pkg/config"
	"process-api/pkg/constant"
	"process-api/pkg/logging"
	"process-api/pkg/model/response"
	"process-api/pkg/utils"

	"braces.dev/errtrace"
	"github.com/labstack/echo/v4"
)

// @summary AddressAutoComplete
// @description Returns a list of addresses matching the search query prefix
// @tags onboardingPrimaryAddress
// @produce json
// @param search query string true "Smarty requires the 'search' parameter in a specific order to return results Format should follow: streetNumber → street → optional suffix → city/state/zip"
// @param Authorization header string true "Bearer token for user authentication"
// @success 200 {object} response.Address
// @header 200 {string} Authorization "Bearer token for user authentication"
// @success 400 {object} response.BadRequestErrors
// @failure 500 {object} response.ErrorResponse
// @router /onboarding/customer/address-autocomplete [get]
func AddressAutoComplete(c echo.Context) error {
	search := c.QueryParam("search")

	if !utils.ValidateSearchFieldLength(search) {
		return response.BadRequestErrors{
			Errors: []response.BadRequestError{
				{
					FieldName: "search",
					Error:     constant.SEARCH_FIELD_TOO_LONG_ERROR_MSG,
				},
			},
		}
	}

	authID := config.Config.SmartyStreets.AuthId
	authToken := config.Config.SmartyStreets.AuthToken
	client := utils.NewAutocompleteClient(authID, authToken)
	filteredSuggestions, err := client.Lookup(search, 10)
	if err != nil {
		logging.Logger.Error("Error from smarty street", "error", err.Error())
		return response.InternalServerError(fmt.Sprintf("Error from smarty street: %s", err.Error()), errtrace.Wrap(err))
	}

	addresses := response.Address{
		Suggestions: filteredSuggestions,
	}

	if len(addresses.Suggestions) == 0 {
		logging.Logger.Error("No address suggestions found for given prefix")
	}

	return c.JSON(http.StatusOK, addresses)
}

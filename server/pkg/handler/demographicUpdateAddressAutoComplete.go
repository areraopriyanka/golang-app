package handler

import (
	"fmt"
	"net/http"
	"process-api/pkg/config"
	"process-api/pkg/constant"
	"process-api/pkg/db/dao"
	"process-api/pkg/logging"
	"process-api/pkg/model/response"
	"process-api/pkg/security"
	"process-api/pkg/utils"

	"braces.dev/errtrace"
	"github.com/labstack/echo/v4"
)

// @summary DemographicUpdateAddressAutoComplete
// @description Returns addresses matching the search query prefix for a demographic update
// @tags DemographicPrimaryAddress
// @produce json
// @param search query string true "Smarty requires the 'search' parameter in a specific order to return results Format should follow: streetNumber → street → optional suffix → city/state/zip"
// @param Authorization header string true "Bearer token for user authentication"
// @success 200 {object} response.Address
// @header 200 {string} Authorization "Bearer token for user authentication"
// @success 400 {object} response.BadRequestErrors
// @Failure 401 {object} response.ErrorResponse
// @failure 500 {object} response.ErrorResponse
// @router /account/customer/demographic-update/address-autocomplete [get]
func DemographicUpdateAddressAutoComplete(c echo.Context) error {
	cc, ok := c.(*security.LoggedInRegisteredUserContext)
	if !ok {
		return response.UnauthorizedError("Failed to get user Id from custom context")
	}
	userId := cc.UserId

	logger := logging.GetEchoContextLogger(c)

	_, errResponse := dao.RequireUserWithState(userId, constant.ACTIVE)
	if errResponse != nil {
		return errResponse
	}

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
		logger.Error("Error from smarty street", "error", err.Error())
		return response.InternalServerError(fmt.Sprintf("Error from smarty street: %s", err.Error()), errtrace.Wrap(err))
	}

	addresses := response.Address{
		Suggestions: filteredSuggestions,
	}

	if len(addresses.Suggestions) == 0 {
		logger.Error("No address suggestions found for given prefix")
	}

	return c.JSON(http.StatusOK, addresses)
}

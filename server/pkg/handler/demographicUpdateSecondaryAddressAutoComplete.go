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

// @summary DemographicUpdateSecondaryAddressAutoComplete
// @description Returns a list of nested addresses, including all matching apartments for the given address for demographic updates
// @tags DemographicSecondaryAddress
// @accept json
// @produce json
// @param Authorization header string true "Bearer token for user authentication"
// @param secondaryAddressAutoCompleteRequest body  response.Suggestion true "SecondaryAddressAutoComplete payload"
// @Success 200 {object} response.Address
// @header 200 {string} Authorization "Bearer token for user authentication"
// @failure 400 {object} response.BadRequestErrors
// @Failure 401 {object} response.ErrorResponse
// @failure 500 {object} response.ErrorResponse
// @router /account/customer/demographic-update/secondary-address-autocomplete [post]
func DemographicUpdateSecondaryAddressAutoComplete(c echo.Context) error {
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

	authID := config.Config.SmartyStreets.AuthId
	authToken := config.Config.SmartyStreets.AuthToken

	var suggestion response.Suggestion

	if err := c.Bind(&suggestion); err != nil {
		logger.Error("Error decoding request data", "error", err.Error())
		return response.BadRequestInvalidBody
	}

	if err := c.Validate(suggestion); err != nil {
		return err
	}

	client := utils.NewAutocompleteClient(authID, authToken)
	suggestions, err := client.FetchNestedSuggestions(suggestion)
	if err != nil {
		logger.Error("Error fetching nested suggestions", "error", err.Error())
		return response.InternalServerError(fmt.Sprintf("Error fetching nested suggestions: %s", err.Error()), errtrace.Wrap(err))
	}

	nestedAddresses := response.Address{
		Suggestions: suggestions,
	}

	if len(nestedAddresses.Suggestions) == 0 {
		logger.Error("No suggestions found for given address")
	}

	return c.JSON(http.StatusOK, nestedAddresses)
}

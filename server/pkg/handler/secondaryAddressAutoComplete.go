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

// @summary SecondaryAddressAutoComplete
// @description Returns a list of nested addresses, including all matching apartments for the given address
// @tags onboardingSecondaryAddress
// @accept json
// @produce json
// @param Authorization header string true "Bearer token for user authentication"
// @param secondaryAddressAutoCompleteRequest body  response.Suggestion true "SecondaryAddressAutoComplete payload"
// @Success 200 {object} response.Address
// @header 200 {string} Authorization "Bearer token for user authentication"
// @failure 400 {object} response.BadRequestErrors
// @failure 500 {object} response.ErrorResponse
// @router /onboarding/customer/secondary-address-autocomplete [post]
func SecondaryAddressAutoComplete(c echo.Context) error {
	authID := config.Config.SmartyStreets.AuthId
	authToken := config.Config.SmartyStreets.AuthToken

	var suggestion response.Suggestion

	logger := logging.GetEchoContextLogger(c)

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
		logging.Logger.Error("Error fetching nested suggestions", "error", err.Error())
		return response.ErrorResponse{ErrorCode: constant.INTERNAL_SERVER_ERROR, StatusCode: http.StatusInternalServerError, LogMessage: fmt.Sprintf("Error fetching nested suggestions: error: %s", err.Error()), MaybeInnerError: errtrace.Wrap(err)}
	}

	nestedAddresses := response.Address{
		Suggestions: suggestions,
	}

	if len(nestedAddresses.Suggestions) == 0 {
		logging.Logger.Error("No suggestions found for given address")
	}

	return c.JSON(http.StatusOK, nestedAddresses)
}

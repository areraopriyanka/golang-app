package handler

import (
	"fmt"
	"net/http"
	"process-api/pkg/config"
	"process-api/pkg/constant"
	"process-api/pkg/db/dao"
	"process-api/pkg/ledger"
	"process-api/pkg/logging"
	"process-api/pkg/model/request"
	"process-api/pkg/model/response"
	"process-api/pkg/security"
	"process-api/pkg/utils"

	"braces.dev/errtrace"
	"github.com/labstack/echo/v4"
)

// @Summary BuildValidateCvvPayload
// @Description Generate ValidateCvv payload.
// @Tags cards
// @Accept json
// @Produce json
// @Param payload body request.BuildValidateCvvRequest true "payload with cvv"
// @Param Authorization header string true "Bearer token for user authentication"
// @Success 200 {object} response.BuildPayloadResponse
// @header 200 {string} Authorization "Bearer token for user authentication"
// @Failure 400 {object} response.BadRequestErrors
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 412 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /account/cards/validate-cvv/build [post]
func BuildValidateCvvPayload(c echo.Context) error {
	cc, ok := c.(*security.LoggedInRegisteredUserContext)
	if !ok {
		return response.UnauthorizedError("Failed to get user Id from custom context")
	}

	userId := cc.UserId
	logger := logging.GetEchoContextLogger(c)

	requestData := new(request.BuildValidateCvvRequest)
	err := c.Bind(requestData)
	if err != nil {
		return response.BadRequestInvalidBody
	}

	// validations
	if err := c.Validate(requestData); err != nil {
		return err
	}

	_, errResponse := dao.RequireUserWithState(userId, constant.ACTIVE)
	if errResponse != nil {
		return errResponse
	}
	userAccountCard, errResponse := dao.RequireActiveCardHolderForUser(userId)
	if errResponse != nil {
		return errResponse
	}

	payloadResponse, errResponse := getValidateCvvPayload(userAccountCard.CardId, requestData.CVV, userId)
	if errResponse != nil {
		return errResponse
	}

	logger.Info("ValidateCvv payload created successfully for user", "userId", userId)
	return c.JSON(http.StatusOK, payloadResponse)
}

func getValidateCvvPayload(cardId string, cvv string, userId string) (*response.BuildPayloadResponse, *response.ErrorResponse) {
	userClient := ledger.NewNetXDCardApiClient(config.Config.Ledger, nil)
	payload, err := userClient.BuildValidateCvvRequest(cardId, cvv, utils.IsValidUnencrypted3DigitCVV(cvv))
	if err != nil {
		return nil, &response.ErrorResponse{ErrorCode: constant.INTERNAL_SERVER_ERROR, StatusCode: http.StatusInternalServerError, LogMessage: fmt.Sprintf("error occurred while creating validate cvv payload: %s", err.Error()), MaybeInnerError: errtrace.Wrap(err)}
	}

	payloadResponse, errResponse := dao.CreateSignablePayloadForUser(userId, payload)
	if errResponse != nil {
		return nil, errResponse
	}
	return payloadResponse, nil
}

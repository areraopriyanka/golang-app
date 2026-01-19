package handler

import (
	"net/http"
	"process-api/pkg/config"
	"process-api/pkg/constant"
	"process-api/pkg/db/dao"
	"process-api/pkg/ledger"
	"process-api/pkg/logging"
	"process-api/pkg/model/response"
	"process-api/pkg/security"

	"github.com/labstack/echo/v4"
)

// @Summary BuildGetCardLimitPayload
// @Description Create GetCardLimitPayload payload.
// @Tags dashboard
// @Produce json
// @Param Authorization header string true "Bearer token for user authentication"
// @Success 200 {object} response.BuildPayloadResponse
// @header 200 {string} Authorization "Bearer token for user authentication"
// @Failure 500 {object} response.ErrorResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 412 {object} response.ErrorResponse
// @Failure 404  {object} response.ErrorResponse
// @Router /account/dashboard/card-limit/build [post]
func BuildGetCardLimitPayload(c echo.Context) error {
	cc, ok := c.(*security.LoggedInRegisteredUserContext)
	if !ok {
		return response.UnauthorizedError("Failed to get user Id from custom context")
	}

	logger := logging.GetEchoContextLogger(c)

	userId := cc.UserId

	user, errResponse := dao.RequireUserWithState(userId, constant.ACTIVE)
	if errResponse != nil {
		return errResponse
	}
	userAccountCard, errResponse := dao.RequireActiveCardHolderForUser(userId)
	if errResponse != nil {
		return errResponse
	}

	apiConfig := ledger.NewNetXDCardApiConfig(config.Config.Ledger)
	payload := apiConfig.BuildGetCardLimitRequest(user.LedgerCustomerNumber, userAccountCard.AccountNumber, userAccountCard.CardId)

	payloadResponse, errResponse := dao.CreateSignablePayloadForUser(userId, payload)
	if errResponse != nil {
		return errResponse
	}

	logger.Debug("GetCardLimit payload created successfully for user", "userId", userId)
	return c.JSON(http.StatusOK, payloadResponse)
}

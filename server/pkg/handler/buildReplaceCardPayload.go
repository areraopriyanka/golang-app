package handler

import (
	"net/http"
	"process-api/pkg/config"
	"process-api/pkg/constant"
	"process-api/pkg/db/dao"
	"process-api/pkg/ledger"
	"process-api/pkg/logging"
	"process-api/pkg/model/request"
	"process-api/pkg/model/response"
	"process-api/pkg/security"

	"github.com/labstack/echo/v4"
)

// @Summary BuildReplaceCardPayload
// @Description Create ReplaceCard payload.
// @Tags cards
// @Produce json
// @Param Authorization header string true "Bearer token for user authentication"
// @Param payload body request.ReplaceCardReason true "request.ReplaceCardReason"
// @Success 200 {object} response.BuildPayloadResponse
// @header 200 {string} Authorization "Bearer token for user authentication"
// @Failure 500 {object} response.ErrorResponse
// @Failure 400 {object} response.BadRequestErrors
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 412 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /account/cards/replace/build [post]
func BuildReplaceCardPayload(c echo.Context) error {
	cc, ok := c.(*security.LoggedInRegisteredUserContext)
	if !ok {
		return response.UnauthorizedError("Failed to get user Id from custom context")
	}

	logger := logging.GetEchoContextLogger(c)

	requestData := new(request.ReplaceCardReason)
	err := c.Bind(requestData)
	if err != nil {
		return response.BadRequestInvalidBody
	}

	if err := c.Validate(requestData); err != nil {
		return err
	}

	userId := cc.UserId

	user, errResponse := dao.RequireUserWithState(userId, constant.ACTIVE)
	if errResponse != nil {
		return errResponse
	}
	userAccountCard, errResponse := dao.RequireActiveCardHolderForUser(userId)
	if errResponse != nil {
		return errResponse
	}

	userClient := ledger.NewNetXDCardApiClient(config.Config.Ledger, nil)

	statusAction := ledger.REPLACE_CARD
	if requestData.Reason == "damaged" {
		statusAction = ledger.REISSUE
	}
	replaceOrReissuePayload := userClient.BuildReplaceOrReissueCardRequest(user.LedgerCustomerNumber, userAccountCard.CardId, userAccountCard.AccountNumber, statusAction)

	payloadResponse, errResponse := dao.CreateSignablePayloadForUser(userId, replaceOrReissuePayload)
	if errResponse != nil {
		return errResponse
	}

	logger.Debug("ReplaceOrReissueCard payload created successfully for user", "userId", userId)

	return c.JSON(http.StatusOK, payloadResponse)
}

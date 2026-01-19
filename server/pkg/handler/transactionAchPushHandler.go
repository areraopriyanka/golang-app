package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"process-api/pkg/constant"
	"process-api/pkg/db/dao"
	"process-api/pkg/ledger"
	"process-api/pkg/logging"
	"process-api/pkg/model/response"
	"process-api/pkg/security"
	"process-api/pkg/utils"

	"braces.dev/errtrace"
	"github.com/labstack/echo/v4"
)

// @summary TransactionAchPush
// @description Generate a transaction to send money from the user to an external account
// @tags Transactions
// @accept json
// @produce json
// @param transactionAchPushRequest body TransactionAchPushRequest true "Request body to request a transaction"
// @success 200 {object} TransactionAchPushResponse
// @failure 400 {object} response.BadRequestErrors
// @failure 401 {object} response.ErrorResponse
// @failure 404 {object} response.ErrorResponse
// @failure 409 {object} response.ErrorResponse
// @failure 410 {object} response.ErrorResponse
// @failure 412 {object} response.ErrorResponse
// @failure 500 {object} response.ErrorResponse
// @router /account/accounts/ach/push [post]
func (h *Handler) TransactionAchPush(c echo.Context) error {
	cc, ok := c.(*security.LoggedInRegisteredUserContext)
	if !ok {
		return response.UnauthorizedError("Failed to get user Id from custom context")
	}
	userId := cc.UserId

	logger := logging.GetEchoContextLogger(c)

	user, errResponse := dao.RequireUserWithState(userId, constant.ACTIVE)
	if errResponse != nil {
		return errResponse
	}

	var requestData TransactionAchPushRequest

	if err := c.Bind(&requestData); err != nil {
		logger.Error("Invalid request", "error", err.Error())
		return response.BadRequestErrors{
			Errors: []response.BadRequestError{
				{Error: err.Error()},
			},
		}
	}

	if err := c.Validate(requestData); err != nil {
		return err
	}

	userPublicKey, errResponse := dao.RequireUserPublicKey(userId, cc.PublicKey)
	if errResponse != nil {
		return errResponse
	}

	payloadRecord, errResponse := dao.ConsumePayload(user.Id, requestData.PayloadId)
	if errResponse != nil {
		return errResponse
	}

	decryptedLedgerPassword, decryptedApiKey, err := utils.DecryptApiKeyAndLedgerPassword(user.LedgerPassword, user.KmsEncryptedLedgerPassword, userPublicKey.ApiKey, userPublicKey.KmsEncryptedApiKey, logger)
	if err != nil {
		logger.Error(err.Error())
		return response.ErrorResponse{ErrorCode: constant.INTERNAL_SERVER_ERROR, StatusCode: http.StatusInternalServerError, LogMessage: err.Error(), MaybeInnerError: errtrace.Wrap(err)}
	}

	userClient := ledger.CreatePaymentApiClient(userPublicKey.PublicKey, requestData.Signature, payloadRecord.Payload, user.Email, decryptedLedgerPassword, userPublicKey.KeyId, decryptedApiKey)

	var request ledger.OutboundAchCreditRequest
	err = json.Unmarshal([]byte(payloadRecord.Payload), &request)
	if err != nil {
		logger.Error("Error while unmarshaling payload", "error", err.Error())
		return response.ErrorResponse{ErrorCode: constant.INTERNAL_SERVER_ERROR, StatusCode: http.StatusInternalServerError, LogMessage: fmt.Sprintf("Error while unmarshaling payload: error: %s", err.Error())}
	}

	responseData, err := userClient.OutboundAchCredit(request)
	if err != nil {
		logger.Error("Error from callLedgerOutboundAchCredit", "error", err.Error())
		return response.ErrorResponse{ErrorCode: constant.INTERNAL_SERVER_ERROR, StatusCode: http.StatusInternalServerError, LogMessage: fmt.Sprintf("Error from callLedgerOutboundAchCredit: error: %s", err.Error()), MaybeInnerError: errtrace.Wrap(err)}
	}

	if responseData.Error != nil && (responseData.Error.Code == "5019" && responseData.Error.Message == "INSUFFICIENT BALANCE") {
		logger.Error("The ledger responded with an error for insufficient balance", "code", responseData.Error.Code, "msg", responseData.Error.Message)
		return response.ErrorResponse{ErrorCode: "INSUFFICIENT_FUNDS", Message: "Insufficient funds", StatusCode: http.StatusConflict, MaybeInnerError: errtrace.New("")}
	}

	if responseData.Error != nil {
		logger.Error("The ledger responded with an error", "code", responseData.Error.Code, "msg", responseData.Error.Message)
		return response.ErrorResponse{ErrorCode: constant.INTERNAL_SERVER_ERROR, StatusCode: http.StatusInternalServerError, LogMessage: fmt.Sprintf("The ledger responded with an error: %s", responseData.Error.Message), MaybeInnerError: errtrace.New("")}
	}

	if responseData.Result == nil {
		logger.Error("The ledger responded with an empty result object", "responseData", responseData)
		return response.ErrorResponse{ErrorCode: constant.INTERNAL_SERVER_ERROR, StatusCode: http.StatusInternalServerError, LogMessage: "The ledger responded with an empty result object", MaybeInnerError: errtrace.New("")}
	}

	if responseData.Result.Api.Type != "ACH_OUT_ACK" {
		logger.Error("Unexpected response type from Ledger API", "apiType", responseData.Result.Api.Type)
		return c.JSON(http.StatusOK, TransactionAchPushResponse{
			Reference:         responseData.Result.Api.Reference,
			Status:            responseData.Result.TransactionStatus,
			Amount:            responseData.Result.TransactionAmountCents,
			TransactionNumber: responseData.Result.TransactionNumber,
		})
	}

	transactionResponse := TransactionAchPushResponse{
		Reference:         responseData.Result.Api.Reference,
		Status:            responseData.Result.TransactionStatus,
		Amount:            responseData.Result.TransactionAmountCents,
		TransactionNumber: responseData.Result.TransactionNumber,
	}

	return c.JSON(http.StatusOK, transactionResponse)
}

type TransactionAchPushRequest struct {
	Signature string `json:"signature" validate:"required"`
	PayloadId string `json:"payloadId" validate:"required"`
}

type TransactionAchPushResponse struct {
	Reference         string `json:"reference" validate:"required"`
	Status            string `json:"status" validate:"required"`
	Amount            int64  `json:"amount" validate:"required"`
	TransactionNumber string `json:"transactionNumber" validate:"required"`
}

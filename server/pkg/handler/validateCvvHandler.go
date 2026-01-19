package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
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

// @Summary ValidateCvv
// @Description Validates cvv.
// @Tags cards
// @Produce json
// @Param payload body request.LedgerApiRequest true "ValidateCvvRequest"
// @Param Authorization header string true "Bearer token for user authentication"
// @Success 200 {object} response.ValidateCvvResponse
// @header 200 {string} Authorization "Bearer token for user authentication"
// @Failure 400 {object} response.BadRequestErrors
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 410 {object} response.ErrorResponse
// @Failure 412 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /account/cards/validate-cvv [post]
func ValidateCvv(c echo.Context) error {
	cc, ok := c.(*security.LoggedInRegisteredUserContext)
	if !ok {
		return response.UnauthorizedError("Failed to get user Id from custom context")
	}

	userId := cc.UserId

	logger := logging.GetEchoContextLogger(c)

	requestData := new(request.LedgerApiRequest)
	err := c.Bind(requestData)
	if err != nil {
		return response.BadRequestInvalidBody
	}

	// validations
	if err := c.Validate(requestData); err != nil {
		return err
	}

	user, errResponse := dao.RequireUserWithState(userId, constant.ACTIVE)
	if errResponse != nil {
		return errResponse
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

	userClient := ledger.CreateCardApiClient(userPublicKey.PublicKey, requestData.Signature, payloadRecord.Payload, user.Email, decryptedLedgerPassword, userPublicKey.KeyId, decryptedApiKey)

	var responseData ledger.NetXDApiResponse[ledger.ValidateCvvResult]

	var request ledger.ValidateCvvRequest
	err = json.Unmarshal([]byte(payloadRecord.Payload), &request)
	if err != nil {
		logger.Error("Error while unmarshaling payload", "error", err.Error())
		return response.ErrorResponse{ErrorCode: constant.INTERNAL_SERVER_ERROR, StatusCode: http.StatusInternalServerError, LogMessage: fmt.Sprintf("Error while unmarshaling payload: error: %s", err.Error()), MaybeInnerError: errtrace.Wrap(err)}
	}

	responseData, err = userClient.ValidateCvv(request)
	if err != nil {
		logger.Error("Error from callLedgerGetCard", "error", err.Error())
		return response.ErrorResponse{ErrorCode: constant.INTERNAL_SERVER_ERROR, StatusCode: http.StatusInternalServerError, LogMessage: fmt.Sprintf("Error from callLedgerGetCard: error: %s", err.Error()), MaybeInnerError: errtrace.Wrap(err)}
	}

	if responseData.Error != nil {
		logger.Error("The ledger responded with an error", "code", responseData.Error.Code, "msg", responseData.Error.Message)
		return response.ErrorResponse{ErrorCode: constant.INTERNAL_SERVER_ERROR, StatusCode: http.StatusInternalServerError, LogMessage: fmt.Sprintf("The ledger responded with an error:%s", responseData.Error.Message), MaybeInnerError: errtrace.New("")}
	}

	isValid := responseData.Result != nil && responseData.Result.Api.Type == ledger.ValidateCvvValidType
	logger.Info("validate cvv", "isValid", isValid)

	return c.JSON(http.StatusOK, response.ValidateCvvResponse{IsValidCvv: isValid})
}

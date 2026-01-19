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

// @summary SetCardPin
// @description Sets a PIN for customer's card
// @tags cards
// @accept json
// @produce json
// @param setCardPinRequest body request.SetCardPinRequest true "Request body to set the card PIN"
// @success 200 {object} response.SetCardPinResponse
// @failure 400 {object} response.BadRequestErrors
// @failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 422 {object} response.ErrorResponse
// @failure 500 {object} response.ErrorResponse
// @router /account/cards/pin [post]
func SetCardPin(c echo.Context) error {
	cc, ok := c.(*security.LoggedInRegisteredUserContext)
	if !ok {
		return response.UnauthorizedError("Failed to get user Id from custom context")
	}
	userId := cc.UserId

	logger := logging.GetEchoContextLogger(c)

	var requestData request.SetCardPinRequest

	err := json.NewDecoder(c.Request().Body).Decode(&requestData)
	if err != nil {
		return response.BadRequestInvalidBody
	}

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
	payloadRecord, errResponse := dao.RequireSignablePayload(requestData.PayloadId)
	if errResponse != nil {
		return errResponse
	}

	apiPath := "/account/cards/pin/otp"
	err = utils.CheckOtpChallengeIsNotExpired(requestData.OtpId, requestData.Otp, apiPath, userId)
	if err != nil {
		logging.Logger.Error("error verifying otp for card reset", "error", err.Error())
		return response.GenerateOTPErrResponse(errtrace.Wrap(err))
	}

	decryptedLedgerPassword, decryptedApiKey, err := utils.DecryptApiKeyAndLedgerPassword(user.LedgerPassword, user.KmsEncryptedLedgerPassword, userPublicKey.ApiKey, userPublicKey.KmsEncryptedApiKey, logger)
	if err != nil {
		logger.Error(err.Error())
		return response.InternalServerError(fmt.Sprintf("error in decryption: %s", err.Error()), errtrace.Wrap(err))
	}

	userClient := ledger.CreateCardApiClient(userPublicKey.PublicKey, requestData.Signature, payloadRecord.Payload, user.Email, decryptedLedgerPassword, userPublicKey.KeyId, decryptedApiKey)

	var request ledger.ChangePinRequest
	err = json.Unmarshal([]byte(payloadRecord.Payload), &request)
	if err != nil {
		logger.Error("Error while unmarshaling payload", "error", err.Error())
		return response.InternalServerError(fmt.Sprintf("Error while unmarshaling payload: %s", err.Error()), errtrace.Wrap(err))
	}
	responseData, err := userClient.ChangePin(request)
	if err != nil {
		logger.Error("Error from callLedgerChangePin", "error", err.Error())
		return response.InternalServerError(fmt.Sprintf("Error from callLedgerChangePin: %s", err.Error()), errtrace.Wrap(err))
	}

	if responseData.Error != nil {
		logger.Error("The ledger responded with an error", "code", responseData.Error.Code, "msg", responseData.Error.Message)
		return response.InternalServerError(fmt.Sprintf("The ledger responded with an error: %s", responseData.Error.Message), errtrace.New(""))
	}

	if responseData.Result == nil {
		logger.Error("The ledger responded with an empty result object", "responseData", responseData)
		return response.InternalServerError("The ledger responded with an empty result object", errtrace.New(""))
	}

	if responseData.Result.Api.Type != "PIN_CHANGE_ACK" {
		logger.Error("Unexpected response type from Ledger API", "apiType", responseData.Result.Api.Type)
		return c.JSON(http.StatusOK, response.SetCardPinResponse{PinSet: false})
	}

	setCardPinResponse := response.SetCardPinResponse{
		PinSet: true,
	}

	err = utils.ExpireOtpChallenge(requestData.OtpId)
	if err != nil {
		logger.Error(err.Error())
	}

	return c.JSON(http.StatusOK, setCardPinResponse)
}

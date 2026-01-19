package handler

import (
	"encoding/base64"
	"encoding/json"
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

// @Summary GetStatement
// @Description Returns a PDF file for the specified statement id
// @Tags statement
// @Accept json
// @Produce application/pdf
// @Param Authorization header string true "Bearer token for user authentication"
// @Param payload body request.LedgerApiRequest true "Get statement request with user signed payload and payloadId"
// @Success 200 {file} statement
// @Header 200 {string} Authorization "Bearer token for user authentication"
// @Failure 400 {object} response.BadRequestErrors
// @Failure 401 {object} response.ErrorResponse
// @Failure 404  {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 410 {object} response.ErrorResponse
// @Failure 412 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /account/get-statement [post]
func GetStatement(c echo.Context) error {
	cc, ok := c.(*security.LoggedInRegisteredUserContext)
	if !ok {
		return response.UnauthorizedError("Failed to get user Id from custom context")
	}

	userId := cc.UserId

	user, errResponse := dao.RequireUserWithState(userId, constant.ACTIVE)
	if errResponse != nil {
		return errResponse
	}

	logger := logging.GetEchoContextLogger(c)

	requestData := new(request.LedgerApiRequest)
	err := c.Bind(requestData)
	if err != nil {
		return response.BadRequestInvalidBody
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
		return response.InternalServerError(fmt.Sprintf("Error while decrypting apiKey and ledgerPassword: %s", err.Error()), errtrace.Wrap(err))
	}

	userParamsBuilder := ledger.NewPreSignedParamsBuilder(userPublicKey.PublicKey, requestData.Signature, payloadRecord.Payload, user.Email, decryptedLedgerPassword, userPublicKey.KeyId, decryptedApiKey)
	userClient := ledger.NewNetXDLedgerApiClient(config.Config.Ledger, userParamsBuilder)

	var request ledger.GetStatementRequest
	err = json.Unmarshal([]byte(payloadRecord.Payload), &request)
	if err != nil {
		logger.Error("Error while unmarshaling payload", "error", err.Error())
		return response.InternalServerError(fmt.Sprintf("Error while unmarshaling payload: %s", err.Error()), errtrace.Wrap(err))
	}

	var responseData ledger.NetXDApiResponse[ledger.GetStatementResult]
	responseData, err = userClient.GetStatement(request)
	if err != nil {
		logger.Error("Error from GetStatement", "error", err.Error())
		return response.InternalServerError(fmt.Sprintf("Error from GetStatement: %s", err.Error()), errtrace.Wrap(err))
	}

	if responseData.Error != nil {
		logger.Error("The ledger responded with an error", "code", responseData.Error.Code, "msg", responseData.Error.Message)
		return response.InternalServerError(fmt.Sprintf("The ledger responded with an error: %s", responseData.Error.Message), errtrace.New(""))
	}

	if responseData.Result == nil {
		logger.Error("The ledger responded with an empty result object", "response", responseData)
		return response.InternalServerError("The ledger responded with an empty result object", errtrace.New(""))
	}

	pdfBase64, err := responseData.Result.PdfBase64()
	if err != nil {
		logger.Error("Error retrieving base64 data from payload", "error", err)
		return response.InternalServerError(fmt.Sprintf("Error retrieving base64 data from payload: %s", err.Error()), errtrace.Wrap(err))
	}

	pdfBytes, err := base64.StdEncoding.DecodeString(pdfBase64)
	if err != nil {
		logger.Error("Error while decoding the pdf data", "error", err)
		return response.InternalServerError(fmt.Sprintf("Error while decoding the pdf data: %s", err.Error()), errtrace.Wrap(err))
	}

	fileName, err := utils.GenerateStatementFileName(responseData.Result.LastDate)
	if err != nil {
		logger.Error("Error while generating statement title", "error", err)
		return response.InternalServerError(fmt.Sprintf("Error while generating statement title: %s", err.Error()), errtrace.Wrap(err))
	}
	c.Response().Header().Set(echo.HeaderContentDisposition, fmt.Sprintf("attachment; filename=%s", fileName))
	return c.Blob(http.StatusOK, "application/pdf", pdfBytes)
}

package handler

import (
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

// @Summary ListStatements
// @Description Retrieves paginated statements for a given user
// @Tags statement
// @Produce json
// @Param Authorization header string true "Bearer token for user authentication"
// @Param payload body ListStatementsRequest true "List statements request with user signed payload, payloadId, and accountId"
// @Success 200 {object}  response.ListStatementResponse
// @Header 200 {string} Authorization "Bearer token for user authentication"
// @Failure 400 {object} response.BadRequestErrors
// @Failure 401 {object} response.ErrorResponse
// @Failure 404  {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 410 {object} response.ErrorResponse
// @Failure 412 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /account/list-statements [post]
func ListStatements(c echo.Context) error {
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

	var requestData ListStatementsRequest
	err := c.Bind(&requestData)
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

	payloadRecord, errResponse := dao.ConsumePayload(user.Id, requestData.LedgerApiRequest.PayloadId)
	if errResponse != nil {
		return errResponse
	}

	decryptedLedgerPassword, decryptedApiKey, err := utils.DecryptApiKeyAndLedgerPassword(user.LedgerPassword, user.KmsEncryptedLedgerPassword, userPublicKey.ApiKey, userPublicKey.KmsEncryptedApiKey, logger)
	if err != nil {
		logger.Error(err.Error())
		return response.ErrorResponse{
			ErrorCode:       constant.INTERNAL_SERVER_ERROR,
			StatusCode:      http.StatusInternalServerError,
			LogMessage:      fmt.Sprintf("Error in decryption: %s", err.Error()),
			MaybeInnerError: errtrace.Wrap(err),
		}
	}

	userParamsBuilder := ledger.NewPreSignedParamsBuilder(userPublicKey.PublicKey, requestData.LedgerApiRequest.Signature, payloadRecord.Payload, user.Email, decryptedLedgerPassword, userPublicKey.KeyId, decryptedApiKey)
	userClient := ledger.NewNetXDLedgerApiClient(config.Config.Ledger, userParamsBuilder)

	var request ledger.ListStatementRequest
	err = json.Unmarshal([]byte(payloadRecord.Payload), &request)
	if err != nil {
		logger.Error("Error while unmarshaling payload", "error", err.Error())
		return response.ErrorResponse{
			ErrorCode:       constant.INTERNAL_SERVER_ERROR,
			StatusCode:      http.StatusInternalServerError,
			LogMessage:      fmt.Sprintf("Error while unmarshaling payload: %s", err.Error()),
			MaybeInnerError: errtrace.Wrap(err),
		}
	}

	var responseData ledger.NetXDApiResponse[ledger.ListStatementResult]
	responseData, err = userClient.ListStatement(request)
	if err != nil {
		logger.Error("Error from ListStatement", "error", err.Error())
		return response.ErrorResponse{
			ErrorCode:       constant.INTERNAL_SERVER_ERROR,
			StatusCode:      http.StatusInternalServerError,
			LogMessage:      fmt.Sprintf("Error from ListStatement: %s", err.Error()),
			MaybeInnerError: errtrace.Wrap(err),
		}
	}

	if responseData.Error != nil {
		logger.Error("The ledger responded with an error", "code", responseData.Error.Code, "msg", responseData.Error.Message)
		return response.ErrorResponse{
			ErrorCode:       constant.INTERNAL_SERVER_ERROR,
			StatusCode:      http.StatusInternalServerError,
			LogMessage:      fmt.Sprintf("The ledger responded with an error: %s", responseData.Error.Message),
			MaybeInnerError: errtrace.New(""),
		}
	}

	if responseData.Result == nil {
		logger.Error("The ledger responded with an empty result object", "response", responseData)
		return response.ErrorResponse{
			ErrorCode:       constant.INTERNAL_SERVER_ERROR,
			StatusCode:      http.StatusInternalServerError,
			LogMessage:      "The ledger responded with an empty result object",
			MaybeInnerError: errtrace.New(""),
		}
	}

	// If no statements available
	if responseData.Result.TotalCounts == 0 {
		logger.Warn("The ledger result has no statements")
		return c.JSON(http.StatusOK, response.ListStatementResponse{
			Statements: []response.Statement{},
			TotalCount: responseData.Result.TotalCounts,
		})
	}

	statements := make([]response.Statement, 0, len(responseData.Result.Accounts))
	for _, statement := range responseData.Result.Accounts {
		if statement.AccountId == requestData.AccountId {
			statements = append(statements, response.Statement{
				Id:    statement.Id,
				Month: statement.Month,
				Year:  statement.Year,
			})
		}
	}

	statementResponse := response.ListStatementResponse{
		Statements: statements,
		TotalCount: responseData.Result.TotalCounts,
	}
	return c.JSON(http.StatusOK, statementResponse)
}

type ListStatementsRequest struct {
	LedgerApiRequest request.LedgerApiRequest `json:"ledgerApiRequest" validate:"required"`
	AccountId        string                   `json:"accountId" validate:"required"`
}

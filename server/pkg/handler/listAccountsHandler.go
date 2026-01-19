package handler

import (
	"errors"
	"fmt"
	"net/http"
	"process-api/pkg/config"
	"process-api/pkg/constant"
	"process-api/pkg/db"
	"process-api/pkg/db/dao"
	"process-api/pkg/ledger"
	"process-api/pkg/logging"
	"process-api/pkg/model/response"
	"process-api/pkg/security"

	"braces.dev/errtrace"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo/v4"
)

type ListAccountAndFirstNameResponse struct {
	FirstName     string           `json:"firstName" validate:"required" mask:"true"`
	Accounts      []ledger.Account `json:"accounts" validate:"required" mask:"true"`
	EncryptionKey string           `json:"encryptionKey" validate:"required,max=255" mask:"true"`
}

// @Summary ListAccounts
// @Description ListAccounts returns the list of accounts and user's firstName
// @Tags dashboard
// @Produce json
// @Param Authorization header string true "Bearer token for user authentication"
// @Success 200 {object} ListAccountAndFirstNameResponse
// @header 200 {string} Authorization "Bearer token for user authentication"
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /account/dashboard/accounts [get]
func ListAccounts(c echo.Context) error {
	cc, ok := c.(*security.LoggedInRegisteredUserContext)
	if !ok {
		return response.UnauthorizedError("Failed to get user Id from custom context")
	}
	userId := cc.UserId

	logger := logging.GetEchoContextLogger(c)

	var user dao.MasterUserRecordDao
	result := db.DB.Where("id=?", userId).Find(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			logger.Error("Could not find the user in db", "error", result.Error.Error())
			return c.NoContent(http.StatusNotFound)
		}
		return response.ErrorResponse{ErrorCode: constant.INTERNAL_SERVER_ERROR, StatusCode: http.StatusInternalServerError, LogMessage: fmt.Sprintf("DB Error: %s", result.Error.Error()), MaybeInnerError: errtrace.Wrap(result.Error)}
	}

	responseData, err := GetLedgerAccountsByCustomerNumber(user.LedgerCustomerNumber)
	if err != nil {
		logger.Error("Error while calling ListAccounts", "error", err.Error())
		return response.ErrorResponse{ErrorCode: constant.INTERNAL_SERVER_ERROR, StatusCode: http.StatusInternalServerError, LogMessage: fmt.Sprintf("Error while calling ListAccounts: error: %s", err.Error()), MaybeInnerError: errtrace.Wrap(err)}
	}

	if responseData.Error != nil {
		logger.Error("Error from ledger ListAccounts", "error", responseData.Error)
		return response.ErrorResponse{ErrorCode: constant.INTERNAL_SERVER_ERROR, StatusCode: http.StatusInternalServerError, LogMessage: fmt.Sprintf("Error from ledger ListAccounts: error: %s", responseData.Error), MaybeInnerError: errtrace.New("")}
	}

	response := ListAccountAndFirstNameResponse{
		FirstName:     responseData.Result.FirstName,
		Accounts:      responseData.Result.Accounts,
		EncryptionKey: config.Config.Ledger.CardsPublicKey,
	}
	logger.Info("Successfully fetched the account list for customer", "ledgerCustomerNo", user.LedgerCustomerNumber)

	return c.JSON(http.StatusOK, response)
}

// TODO: API in https://apidocs.netxd.com/developers/docs/account_apis/Get%20All%20Accounts is currently not functioning
// Therefore, Using "CustomerService.GetCustomer" to fetch account list and user's firstName
func GetLedgerAccountsByCustomerNumber(ledgerCustomerNumber string) (ledger.NetXDApiResponse[ledger.GetCustomerResult], error) {
	ledgerParamsBuilder := ledger.NewLedgerSigningParamsBuilderFromConfig(config.Config.Ledger)
	ledgerClient := ledger.NewNetXDLedgerApiClient(config.Config.Ledger, ledgerParamsBuilder)
	payload := ledger.BuildGetCustomerByCustomerNoPayload(ledgerCustomerNumber)
	return ledgerClient.GetCustomer(*payload)
}

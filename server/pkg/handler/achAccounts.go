package handler

import (
	"fmt"
	"net/http"
	"process-api/pkg/constant"
	"process-api/pkg/db"
	"process-api/pkg/db/dao"
	"process-api/pkg/logging"
	"process-api/pkg/model/response"
	"process-api/pkg/security"

	"braces.dev/errtrace"
	"github.com/labstack/echo/v4"
)

type AchAccount struct {
	ID                    string  `json:"id" validate:"required"`
	Institution           *string `json:"institution"`
	Name                  string  `json:"name" validate:"required"`
	Subtype               string  `json:"subType" validate:"required" enums:"checking,savings"`
	Mask                  *string `json:"mask"`
	AvailableBalanceCents *int64  `json:"availableBalanceCents"`
	// true if the external account needs to be reconnected via Plaid Link's Update mode
	NeedsPlaidLinkUpdate bool `json:"needsPlaidLinkUpdate"`
	// true if the external account is still being verified
	IsPendingVerification bool `json:"isPendingVerification"`
}

type AchAccountsResponse struct {
	DreamFiAccounts  []AchAccount `json:"dreamFiAccounts" validate:"required"`
	ExternalAccounts []AchAccount `json:"externalAccounts" validate:"required"`
}

// @Summary AchAccounts
// @Description Retrieves DreamFi and external account information to facilitate ACH transfers
// @Tags ach
// @Produce json
// @Param Authorization header string true "Bearer token for user authentication"
// @Success 200 {object} AchAccountsResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 412 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /account/ach/accounts [get]
func (h *Handler) AchAccounts(c echo.Context) error {
	cc, ok := c.(*security.LoggedInRegisteredUserContext)
	if !ok {
		return response.UnauthorizedError("Failed to get user Id from custom context")
	}
	userId := cc.UserId
	logger := logging.GetEchoContextLogger(c).WithGroup("AchAccounts").With("userId", userId)
	user, errResponse := dao.RequireUserWithState(userId, constant.ACTIVE)
	if errResponse != nil {
		return errResponse
	}
	// I think it is right to assume that we need an active card holder here, but if
	// this causes issues, adjust as needed.
	_, errResponse = dao.RequireActiveCardHolderForUser(userId)
	if errResponse != nil {
		return errResponse
	}

	responseData, err := GetLedgerAccountsByCustomerNumber(user.LedgerCustomerNumber)
	if err != nil {
		logger.Error("Error while calling CustomerService.GetCustomer", "error", err.Error())
		return response.InternalServerError(fmt.Sprintf("Error while calling CustomerService.GetCustomer: %s", err.Error()), errtrace.Wrap(err))
	}
	if responseData.Error != nil {
		logger.Error("Error from ledger CustomerService.GetCustomer", "error", responseData.Error)
		return response.InternalServerError(fmt.Sprintf("Error from ledger CustomerService.GetCustomer: %s", responseData.Error.Message), errtrace.New(""))
	}
	ledgerAccounts := responseData.Result.Accounts
	dreamFiAccounts := make([]AchAccount, 0, len(ledgerAccounts))
	for _, ledgerAccount := range ledgerAccounts {
		if MapLedgerCardStatus(ledgerAccount.Status, logger) != "active" {
			continue
		}
		mask := ledgerAccount.Number
		if len(mask) > 4 {
			mask = mask[len(mask)-4:]
		}
		// While ledgerAccount has many different "balance" fields, it looks like `Balance` is the correct
		// one to use, provided the docs from GetAccount apply here as well:
		// https://apidocs.netxd.com/developers/docs/account_apis/Get%20Account
		availableBalanceCents := int64(ledgerAccount.Balance)
		dreamFiAccounts = append(dreamFiAccounts, AchAccount{
			ID:                    ledgerAccount.ID,
			Institution:           &ledgerAccount.InstitutionName,
			Name:                  ledgerAccount.Name,
			Subtype:               ledgerAccount.AccountType,
			Mask:                  &mask,
			AvailableBalanceCents: &availableBalanceCents,
		})
	}

	var records []PlaidAccountWithItemDetails
	result := db.DB.
		Table("plaid_accounts").
		Select("plaid_accounts.*, plaid_items.item_error").
		Joins("JOIN plaid_items ON plaid_items.plaid_item_id = plaid_accounts.plaid_item_id").
		Where("plaid_accounts.user_id = ?", userId).
		Find(&records)
	if result.Error != nil {
		logger.Error("query for plaid accounts failed", "error", result.Error.Error())
		return response.InternalServerError(fmt.Sprintf("query for plaid accounts failed: %s", result.Error.Error()), errtrace.Wrap(result.Error))
	}
	externalAccounts := make([]AchAccount, 0, result.RowsAffected)
	for _, record := range records {
		externalAccounts = append(externalAccounts, AchAccount{
			ID:                    record.ID,
			Institution:           record.InstitutionName,
			Name:                  record.Name,
			Subtype:               string(record.Subtype),
			Mask:                  record.Mask,
			AvailableBalanceCents: record.AvailableBalanceCents,
			NeedsPlaidLinkUpdate:  record.ItemError != nil,
			IsPendingVerification: !record.IsVerified(),
		})
	}
	logger.Info("retrieved accounts successfully")
	return c.JSON(http.StatusOK, AchAccountsResponse{ExternalAccounts: externalAccounts, DreamFiAccounts: dreamFiAccounts})
}

type PlaidAccountWithItemDetails struct {
	dao.PlaidAccountDao
	ItemError *string `gorm:"column:item_error"`
}

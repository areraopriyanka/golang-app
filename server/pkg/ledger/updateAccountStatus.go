// Generated from https://apidocs.netxd.com/developers/docs/account_apis/Update%20Account%20Status
package ledger

import (
	"log/slog"

	"braces.dev/errtrace"
	"github.com/go-playground/validator/v10"
)

const (
	// DISABLED - Account is inactive and unusable, either temporarily or permanently.
	DISABLED = "DISABLED"
	// SUSPENDED- Account is Temporarily disabled
	// No transactions are allowed.
	// Admin can update the status of the account to “SUSPENDED” and back to “ACTIVE” based on their compliance policies.
	SUSPENDED = "SUSPENDED"
	// CLOSED - Account is permanently terminated and can no longer be used
	// When the account status is CLOSED, No Activity can be performed in this state.
	// Admin needs to take necessary precautions before changing the status of the Account to CLOSED.
	CLOSED = "CLOSED"
	// DORMANT- Account is inactive for a specific period
	// It can accept only Incoming Payments and Outgoing Payments are restricted.
	// This account status is Automatically updated to ACTIVE as soon as an Inbound Credit is received
	DORMANT = "DORMANT"
	// ACTIVE- Account is currently active and can be used for transactions
	// It can accept both Incoming and Outgoing Payments
	ACTIVE = "ACTIVE"
)

type UpdateAccountStatusRequest struct {
	AccountNumber string `json:"accountNumber" validate:"required"`
	Status        string `json:"status" validate:"required,oneof=CREATED ACTIVE CURTAILED DORMANT SUSPENDED BLOCKED CLOSED DISABLED"`
}

type UpdateAccountStatusResult struct {
	CustomerId    string `json:"CustomerId"`
	AccountNumber string `json:"AccountNumber"`
	InstitutionId string `json:"InstitutionId"`
	Name          string `json:"Name"`
	Status        string `json:"Status"`
}

func BuildUpdateAccountStatusRequest(accountNumber string, status string) UpdateAccountStatusRequest {
	return UpdateAccountStatusRequest{
		AccountNumber: accountNumber,
		Status:        status,
	}
}

func (c *NetXDLedgerApiClient) UpdateAccountStatus(req UpdateAccountStatusRequest) (NetXDApiResponse[UpdateAccountStatusResult], error) {
	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		return NetXDApiResponse[UpdateAccountStatusResult]{}, errtrace.Wrap(err)
	}

	var response NetXDApiResponse[UpdateAccountStatusResult]
	err := c.call("AccountService.UpdateAccountStatus", c.url, req, &response)
	return response, errtrace.Wrap(err)
}

func MapLedgerAccountStatus(s string, logger *slog.Logger) string {
	switch s {
	case SUSPENDED:
		return "temporary_inactive"
	case DISABLED:
		return "disabled"
	case ACTIVE:
		return "active"
	case DORMANT:
		return "dormant"
	case CLOSED:
		return "closed"
	}
	logger.Info("Unknown accountStatus", "status", s)
	return ""
}

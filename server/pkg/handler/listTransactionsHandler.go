package handler

import (
	"fmt"
	"net/http"
	"process-api/pkg/config"
	"process-api/pkg/constant"
	"process-api/pkg/db/dao"
	"process-api/pkg/ledger"
	"process-api/pkg/logging"
	"process-api/pkg/model/response"
	"process-api/pkg/resource/mcc"
	"process-api/pkg/security"
	"process-api/pkg/utils"
	"regexp"
	"slices"
	"strings"
	"time"

	"braces.dev/errtrace"
	"github.com/labstack/echo/v4"
)

var AdministrativeTransactionTypes = []string{
	"OVER_DRAFT_REQ",
	"OVER_DRAFT_RELEASE",
}

// @summary ListTransactions
// @description Get a list of transactions for the user
// @tags Transactions
// @produce json
// @success 200 {object} ListTransactionsResponse
// @failure 401 {object} response.ErrorResponse
// @failure 404 {object} response.ErrorResponse
// @failure 412 {object} response.ErrorResponse
// @failure 500 {object} response.ErrorResponse
// @router /account/accounts/transactions [get]
func ListTransactions(c echo.Context) error {
	cc, ok := c.(*security.LoggedInRegisteredUserContext)
	if !ok {
		return response.UnauthorizedError("Failed to get user Id from custom context")
	}
	userId := cc.UserId

	logger := logging.GetEchoContextLogger(c)
	_, errResponse := dao.RequireUserWithState(userId, ledger.ACTIVE)
	if errResponse != nil {
		return errResponse
	}

	cardHolder, errResponse := dao.RequireCardHolderForUser(userId)
	if errResponse != nil {
		return errResponse
	}

	ledgerParamsBuilder := ledger.NewLedgerSigningParamsBuilderFromConfig(config.Config.Ledger)
	ledgerClient := ledger.NewNetXDLedgerApiClient(config.Config.Ledger, ledgerParamsBuilder)

	request := ledger.BuildListTransactionsByAccountPayload(
		cardHolder.AccountNumber,
	)

	responseData, err := ledgerClient.ListTransactionsByAccount(request)
	if err != nil {
		logger.Error("Error from listTransactionsByAccount", "error", err.Error())
		return response.ErrorResponse{ErrorCode: constant.INTERNAL_SERVER_ERROR, StatusCode: http.StatusInternalServerError, LogMessage: fmt.Sprintf("Error from listTransactionsByAccount: error: %s", err.Error()), MaybeInnerError: errtrace.Wrap(err)}
	}

	if responseData.Error != nil {
		logger.Error("The ledger responded with an error", "code", responseData.Error.Code, "msg", responseData.Error.Message)
		return response.ErrorResponse{
			ErrorCode:       constant.INTERNAL_SERVER_ERROR,
			StatusCode:      http.StatusInternalServerError,
			LogMessage:      fmt.Sprintf("The ledger responded with an error:%s", responseData.Error.Message),
			MaybeInnerError: errtrace.New(""),
		}
	}

	if responseData.Result == nil {
		logger.Error("The ledger responded with an empty result object", "responseData", responseData)
		return response.ErrorResponse{ErrorCode: constant.INTERNAL_SERVER_ERROR, StatusCode: http.StatusInternalServerError, LogMessage: "The ledger responded with an empty result object", MaybeInnerError: errtrace.New("")}
	}

	finalTransactions, completionTimeStamps := MergeTransactions(responseData.Result.AccountTransactions)

	transformedTransactions := make([]Transaction, 0, len(finalTransactions))

	referenceIds := make([]string, 0, len(finalTransactions))
	for _, data := range finalTransactions {
		referenceIds = append(referenceIds, data.ReferenceID)
	}

	disputes, err := dao.TransactionDisputeDao{}.FindByUserIdAndReferenceIds(userId, referenceIds)
	if err != nil {
		logger.Error("Failed to get disputes for transactions", "error", err.Error())
		return errtrace.Wrap(err)
	}

	for _, data := range finalTransactions {
		var disputeStatus, disputeCreatedAt, disputeUpdatedAt *string

		if !slices.Contains(NonDisputableTransactionTypes, data.Type) {
			disputeIndex := slices.IndexFunc(*disputes, func(d dao.TransactionDisputeDao) bool { return d.TransactionIdentifier == data.ReferenceID })
			if disputeIndex == -1 {
				disputeStatus = utils.Pointer("none")
			} else {
				dispute := (*disputes)[disputeIndex]
				disputeStatus = utils.Pointer(dispute.Status)
				disputeCreatedAt = utils.Pointer(dispute.CreatedAt.Format(time.RFC3339)) // Todo: Remove disputeCreatedAt as we can rely on disputeUpdatedAt
				disputeUpdatedAt = utils.Pointer(dispute.UpdatedAt.Format(time.RFC3339))
			}
		}

		var cardAcceptor *string
		if data.CardAcceptor == "" {
			cardAcceptor = nil
		} else {
			cardAcceptor = &data.CardAcceptor
		}

		var merchantAccount ledger.ListTransactionsByAccountResultTransactionAccount
		if data.Credit {
			merchantAccount = data.DebtorAccount
		} else {
			merchantAccount = data.CreditorAccount
		}
		merchantName := ledger.GetTransactionAccountMerchantName(merchantAccount)
		var completionTimestamp *string
		if timeStamp, ok := completionTimeStamps[data.ProcessID]; ok && timeStamp != "" {
			completionTimestamp = &timeStamp
		}

		transactionType := data.TransactionTypeDetails
		if transactionType == "" {
			transactionType = data.Type
		}

		merchantCategory, _ := mcc.GetCategory(data.Mcc)

		transformedTransactions = append(transformedTransactions, Transaction{
			MerchantName:        merchantName,
			MerchantCategory:    merchantCategory,
			Type:                transactionType,
			TypeRaw:             data.Type,
			ReferenceID:         data.ReferenceID,
			TimeStamp:           data.TimeStamp,
			CompletionTimeStamp: completionTimestamp,
			InstructedAmount: InstructedAmount{
				Amount:   data.InstructedAmount.Amount,
				Currency: data.InstructedAmount.Currency,
			},
			Status:           data.Status,
			Credit:           data.Credit,
			CardAcceptor:     parseCardAcceptor(data.CardAcceptor),
			DisputeStatus:    disputeStatus,
			DisputeCreatedAt: disputeCreatedAt,
			DisputeUpdatedAt: disputeUpdatedAt,
			RawCardAcceptor:  cardAcceptor,
		})
	}

	transactionResponse := ListTransactionsResponse{
		Count:        int64(len(transformedTransactions)),
		Transactions: transformedTransactions,
	}

	return c.JSON(http.StatusOK, transactionResponse)
}

type ListTransactionsResponse struct {
	Count        int64         `json:"count"`
	Transactions []Transaction `json:"transactions" validate:"required"`
}
type Transaction struct {
	MerchantName        string           `json:"merchantName" validate:"required"`
	MerchantCategory    string           `json:"merchantCategory" validate:"required"`
	Type                string           `json:"type" validate:"required"`
	TypeRaw             string           `json:"typeRaw" validate:"required"`
	ReferenceID         string           `json:"referenceID" validate:"required"`
	TimeStamp           string           `json:"timeStamp" validate:"required"`
	CompletionTimeStamp *string          `json:"completionTimeStamp,omitempty"`
	InstructedAmount    InstructedAmount `json:"instructedAmount" validate:"required"`
	DisputeStatus       *string          `json:"disputeStatus" enums:"none,pending,credited,voided,rejected"`
	DisputeCreatedAt    *string          `json:"disputeCreatedAt"`
	DisputeUpdatedAt    *string          `json:"disputeUpdatedAt"`
	Status              string           `json:"status" validate:"required"`
	Credit              bool             `json:"credit" validate:"required"`
	CardAcceptor        *CardAcceptor    `json:"cardAcceptor,omitempty"`
	RawCardAcceptor     *string          `json:"rawCardAcceptor,omitempty"`
}
type InstructedAmount struct {
	Amount   int64  `json:"amount" validate:"required"`
	Currency string `json:"currency" validate:"required"`
}

type CardAcceptor struct {
	Merchant string `json:"merchant,omitempty"`
	City     string `json:"city,omitempty"`
	State    string `json:"state,omitempty"`
	Country  string `json:"country"`
	Phone    string `json:"phone,omitempty"`
	Website  string `json:"website,omitempty"`
}

var states = map[string]bool{
	"AL": true, "AK": true, "AZ": true, "AR": true, "CA": true,
	"CO": true, "CT": true, "DE": true, "FL": true, "GA": true,
	"HI": true, "ID": true, "IL": true, "IN": true, "IA": true,
	"KS": true, "KY": true, "LA": true, "ME": true, "MD": true,
	"MA": true, "MI": true, "MN": true, "MS": true, "MO": true,
	"MT": true, "NE": true, "NV": true, "NH": true, "NJ": true,
	"NM": true, "NY": true, "NC": true, "ND": true, "OH": true,
	"OK": true, "OR": true, "PA": true, "RI": true, "SC": true,
	"SD": true, "TN": true, "TX": true, "UT": true, "VT": true,
	"VA": true, "WA": true, "WV": true, "WI": true, "WY": true,
}

var (
	mobileRegex  = regexp.MustCompile(`\d{10}|\d{3}[-.\s]?\d{3}[-.\s]?\d{4}`)
	websiteRegex = regexp.MustCompile(`[a-zA-Z0-9.-]+\.[a-z]{2,}(?:/[A-Za-z0-9._~!$&'()*+,;=:@%-]*)?`)
)

func parseCardAcceptor(raw string) *CardAcceptor {
	if len(raw) != 40 {
		return nil
	}

	ca := CardAcceptor{}

	merchant := strings.TrimSpace(raw[0:23])
	middle := strings.TrimSpace(raw[23:36])
	state := strings.TrimSpace(raw[36:38])
	country := strings.TrimSpace(raw[38:40])

	ca.Merchant = merchant
	ca.Country = country

	switch {
	case mobileRegex.MatchString(middle):
		ca.Phone = mobileRegex.FindString(middle)
	case websiteRegex.MatchString(middle):
		ca.Website = websiteRegex.FindString(middle)
	default:
		ca.City = middle
	}

	if states[state] {
		ca.State = state
	}

	return &ca
}

func MergeTransactions(transactions []ledger.ListTransactionsByAccountResultTransaction) ([]ledger.ListTransactionsByAccountResultTransaction, map[string]string) {
	processIDIndexMap := make(map[string]int)
	var finalTransactions []ledger.ListTransactionsByAccountResultTransaction
	completionTimeStamps := make(map[string]string)
	for _, txn := range transactions {
		if slices.Contains(AdministrativeTransactionTypes, txn.Type) {
			continue
		}
		idx, exists := processIDIndexMap[txn.ProcessID]

		if !exists {
			finalTransactions = append(finalTransactions, txn)
			processIDIndexMap[txn.ProcessID] = len(finalTransactions) - 1
			continue
		}

		existing := finalTransactions[idx]

		switch {
		case txn.Type == "COMPLETION" && existing.Type == "PRE_AUTH":
			txn.TimeStamp = existing.TimeStamp
			completionTimeStamps[txn.ProcessID] = txn.TimeStamp
			finalTransactions[idx] = txn

		case txn.Type == "PRE_AUTH" && existing.Type == "COMPLETION":
			completionTimeStamps[txn.ProcessID] = existing.TimeStamp
			existing.TimeStamp = txn.TimeStamp
			finalTransactions[idx] = existing
		}
	}

	return finalTransactions, completionTimeStamps
}

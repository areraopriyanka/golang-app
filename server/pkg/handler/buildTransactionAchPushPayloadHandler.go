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
	"process-api/pkg/model"
	"process-api/pkg/model/response"
	"process-api/pkg/plaid"
	"process-api/pkg/security"
	"process-api/pkg/utils"

	"braces.dev/errtrace"
	"github.com/labstack/echo/v4"
)

// @summary BuildTransactionAchPushPayload
// @description Builds and stores a signable payload for requesting a push transaction
// @tags Transactions
// @accept json
// @produce json
// @param buildTransactionAchPushPayloadRequest body BuildTransactionAchPushPayloadRequest true "Request body to request a transaction"
// @param Authorization header string true "Bearer token for user authentication"
// @success 200 {object} response.BuildPayloadResponse
// @header 200 {string} Authorization "Bearer token for user authentication"
// @Failure 400 {object} response.BadRequestErrors
// @failure 401 {object} response.ErrorResponse
// @failure 404 {object} response.ErrorResponse
// @Failure 412 {object} response.ErrorResponse
// @failure 422 {object} response.ErrorResponse
// @failure 500 {object} response.ErrorResponse
// @router /account/accounts/ach/push/build [post]
func (h *Handler) BuildTransactionAchPushPayload(c echo.Context) error {
	cc, ok := c.(*security.LoggedInRegisteredUserContext)
	if !ok {
		return response.UnauthorizedError("Failed to get user Id from custom context")
	}
	userId := cc.UserId

	logger := logging.GetEchoContextLogger(c)

	var requestData BuildTransactionAchPushPayloadRequest

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

	userRecord, errResponse := dao.RequireUserWithState(userId, constant.ACTIVE)
	if errResponse != nil {
		return errResponse
	}

	cardHolder, errResponse := dao.RequireActiveCardHolderForUser(userId)
	if errResponse != nil {
		return errResponse
	}

	ledgerParamsBuilder := ledger.NewLedgerSigningParamsBuilderFromConfig(config.Config.Ledger)
	ledgerClient := ledger.NewNetXDLedgerApiClient(config.Config.Ledger, ledgerParamsBuilder)
	ledgerAccountResp, err := ledgerClient.GetAccount(requestData.DreamFiAccountID)
	if err != nil {
		logger.Error("error while calling ledger's GetAccount", "error", err.Error())
		return response.InternalServerError(fmt.Sprintf("error while calling ledger's GetAccount: %s", err.Error()), errtrace.Wrap(err))
	}
	if ledgerAccountResp.Error != nil {
		logger.Error("error from ledger's GetAccount", "error", ledgerAccountResp.Error)
		return response.InternalServerError(fmt.Sprintf("error from ledger's GetAccount: %s", ledgerAccountResp.Error.Message), errtrace.New(""))
	}
	if ledgerAccountResp.Result == nil {
		logger.Error("no error was reported from ledger GetAccount, but Result is missing from response")
		return response.InternalServerError("no error was reported from ledger GetAccount, but Result is missing from response", errtrace.New(""))
	}
	dreamFiAccount := ledgerAccountResp.Result.Account
	if requestData.DreamFiAccountID != dreamFiAccount.Id {
		logger.Error("request's DreamFiAccountID didn't match the dreamFi account's id", "DreamFiAccountID", requestData.DreamFiAccountID, "dreamFiAccount", dreamFiAccount.Id)
		return response.InternalServerError("request's DreamFiAccountID didn't match the dreamFi account's id", errtrace.New(""))
	}
	if cardHolder.AccountNumber != dreamFiAccount.Number {
		logger.Error("the user's primary card account number doesn't match the requested account's number", "cardHolder.AccountNumber", cardHolder.AccountNumber, "dreamFiAccount.Number", dreamFiAccount.Number)
		return response.InternalServerError("the user's primary card account number doesn't match the requested account's number", errtrace.New(""))
	}
	hasSufficientFunds := dreamFiAccount.Balance-int64(requestData.AmountCents) >= 0
	if !hasSufficientFunds {
		logger.Error("dreamFi account does not have sufficient funds for transfer", "transfer amount cents", requestData.AmountCents, "ledgerAccount.Result.Account.Balance", dreamFiAccount.Balance)
		return response.GenerateErrResponse(constant.INSUFFICIENT_FUNDS, "", "dreamFi account does not have sufficient funds for transfer", http.StatusUnprocessableEntity, errtrace.New(""))
	}

	externalAccountID := requestData.ExternalAccountID
	externalAccountRecord, err := dao.PlaidAccountDao{}.GetAccountForUserByID(userId, externalAccountID)
	if err != nil {
		logger.Error("could not retrieve external account record", "error", err.Error())
		return response.InternalServerError(fmt.Sprintf("could not retrieve external account record: %s", err.Error()), errtrace.Wrap(err))
	}
	if externalAccountRecord == nil {
		return response.BadRequestErrors{
			Errors: []response.BadRequestError{
				{
					FieldName: "externalAccountId",
					Error:     "No matching account found.",
				},
			},
		}
	}

	var plaidItemRecord dao.PlaidItemDao
	err = db.DB.Model(dao.PlaidItemDao{}).Where("user_id=? AND plaid_item_id=?", userId, externalAccountRecord.PlaidItemID).First(&plaidItemRecord).Error
	if err != nil {
		logger.Error("could not retrieve PlaidItem record", "error", err.Error())
		return response.InternalServerError(fmt.Sprintf("could not retrieve PlaidItem record: %s", err.Error()), errtrace.Wrap(err))
	}

	accessToken, err := utils.DecryptPlaidAccessToken(plaidItemRecord.EncryptedAccessToken, plaidItemRecord.KmsEncryptedAccessToken)
	if err != nil {
		logger.Error("Could not decrypt Plaid access token", "error", err.Error())
		return response.InternalServerError(fmt.Sprintf("Could not decrypt Plaid access token: %s", err.Error()), errtrace.Wrap(err))
	}

	ps := plaid.PlaidService{Logger: logger, Plaid: h.Plaid, DB: db.DB, WebhookURL: h.Config.Plaid.WebhookURL}

	var primaryOwnerName string
	if externalAccountRecord.PrimaryOwnerName == nil {
		err = ps.GetIdentity(userId, plaidItemRecord.PlaidItemID, accessToken)
		if err != nil {
			logger.Error("could not get identity information", "plaid_item_id", plaidItemRecord.PlaidItemID, "error", err.Error())
			// It's okay to fail here - we'll fallback to the user's first name (it arguably doesn't matter DT-1144)
		} else {
			err = db.DB.Model(dao.PlaidAccountDao{}).Where("user_id=? AND id=?", userId, externalAccountID).First(&externalAccountRecord).Error
			if err != nil {
				logger.Error("could not re-retrieve external account record after identity update", "error", err.Error())
			}
		}
	}
	if externalAccountRecord.PrimaryOwnerName != nil {
		primaryOwnerName = *externalAccountRecord.PrimaryOwnerName
	} else {
		primaryOwnerName = userRecord.FirstName
		logger.Warn("using user's first name as fallback for ACH transaction", "plaid account id", externalAccountRecord.PlaidAccountID)
	}

	achDetails, err := ps.GetACHDetails(accessToken, externalAccountRecord.PlaidAccountID)
	if err != nil {
		if errors.Is(err, plaid.ErrProductNotReady) {
			return response.GenerateErrResponse("PRODUCT_NOT_READY", "", err.Error(), http.StatusUnprocessableEntity, errtrace.Wrap(err))
		}
		logger.Error("could not get ACH details", "plaid_account_id", externalAccountRecord.PlaidAccountID, "error", err.Error())
		return response.InternalServerError(fmt.Sprintf("could not get ACH details, plaid_account_id: %s, error: %s", externalAccountRecord.PlaidAccountID, err.Error()), errtrace.Wrap(err))
	}

	externalAccountType, err := plaid.PlaidSubtypeToLedgerIdentificationType2(externalAccountRecord.Subtype)
	if err != nil {
		logger.Error("Could not convert plaid account subtype to ledger identificationType2", "error", err.Error(), "plaid subtype", externalAccountType)
		return response.InternalServerError(fmt.Sprintf("Could not convert plaid account subtype to ledger identificationType2, error: %s, plaid subtype: %v", err.Error(), externalAccountType), errtrace.Wrap(err))
	}

	payload := ledger.BuildOutboundAchCreditRequest(
		userRecord,
		cardHolder.AccountNumber,
		ledgerAccountResp.Result.Account.InstitutionID,
		requestData.AmountCents.String(),
		primaryOwnerName,
		achDetails.Account,
		achDetails.Routing,
		externalAccountType,
		requestData.Note,
		nil,
	)

	payloadResponse, errResponse := dao.CreateSignablePayloadForUser(userRecord.Id, payload)
	if errResponse != nil {
		return errResponse
	}

	logger.Info("created ach push payload", "userID", userRecord.Id)

	return c.JSON(http.StatusOK, payloadResponse)
}

type BuildTransactionAchPushPayloadRequest struct {
	AmountCents       model.TransferCents `json:"amountCents" validate:"required"`
	Note              *string             `json:"note,omitempty" validate:"omitnil,validateTransactionNote"`
	DreamFiAccountID  string              `json:"dreamFiAccountId" validate:"required"`
	ExternalAccountID string              `json:"externalAccountId" validate:"required"`
}

package handler

import (
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
	"slices"
	"time"

	"braces.dev/errtrace"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lib/pq"
)

var NonDisputableTransactionTypes = []string{
	"PROVISIONAL_CREDIT",
	"VOID",
}

// @summary SubmitTransactionDispute
// @description Initiates a transaction dispute and stores the record in the middleware database.
// @tags Transaction Disputes
// @accept json
// @param referenceId path string true "reference Id"
// @param payload body dao.SubmitTransactionDisputeRequest true "Transaction dispute payload"
// @Param Authorization header string true "Bearer token for user authentication"
// @success 201 {object} SubmitTransactionDisputeResponse
// @header 201 {string} Authorization "Bearer token for user authentication"
// @Failure 400 {object} response.BadRequestErrors
// @failure 401 {object} response.ErrorResponse
// @failure 409 {object} response.ErrorResponse
// @failure 404 {object} response.ErrorResponse
// @failure 412 {object} response.ErrorResponse
// @failure 500 {object} response.ErrorResponse
// @router /account/customer/transaction/{referenceId}/dispute [post]
func SubmitTransactionDispute(c echo.Context) error {
	cc, ok := c.(*security.LoggedInRegisteredUserContext)
	if !ok {
		return response.UnauthorizedError("Failed to get user Id from custom context")
	}
	userId := cc.UserId

	referenceId := c.Param("referenceId")

	logger := logging.GetEchoContextLogger(c)

	user, errResponse := dao.RequireUserWithState(userId, constant.ACTIVE)
	if errResponse != nil {
		return errResponse
	}

	requestData := new(dao.SubmitTransactionDisputeRequest)
	if err := c.Bind(requestData); err != nil {
		return response.BadRequestInvalidBody
	}

	if err := c.Validate(requestData); err != nil {
		return err
	}

	ledgerParamsBuilder := ledger.NewLedgerSigningParamsBuilderFromConfig(config.Config.Ledger)
	ledgerClient := ledger.NewNetXDLedgerApiClient(config.Config.Ledger, ledgerParamsBuilder)

	payload := ledger.BuildGetTransactionByReferenceNumberRequest(referenceId)

	responseData, err := ledgerClient.GetTransactionByReferenceNumber(payload)
	if err != nil {
		logger.Error("Error from getTransactionByReferenceNumber", "error", err.Error())
		return response.ErrorResponse{
			ErrorCode:       constant.INTERNAL_SERVER_ERROR,
			StatusCode:      http.StatusInternalServerError,
			LogMessage:      fmt.Sprintf("Error from getTransactionByReferenceNumber: %s", err.Error()),
			MaybeInnerError: errtrace.Wrap(err),
		}
	}

	if responseData.Error != nil {
		if responseData.Error.Code == "NOT_FOUND_TRANSACTION" {
			return response.ErrorResponse{ErrorCode: constant.TRANSACTION_DOES_NOT_EXIST, Message: constant.TRANSACTION_DOES_NOT_EXIST_MSG, StatusCode: http.StatusNotFound, MaybeInnerError: errtrace.New("")}
		}
		logger.Error("The ledger responded with an error", "code", responseData.Error.Code, "msg", responseData.Error.Message)
		return response.ErrorResponse{
			ErrorCode:       constant.INTERNAL_SERVER_ERROR,
			StatusCode:      http.StatusInternalServerError,
			LogMessage:      fmt.Sprintf("The ledger responded with an error: %s", responseData.Error.Message),
			MaybeInnerError: errtrace.New(""),
		}
	}

	if responseData.Result == nil {
		logger.Error("The ledger responded with an empty result object", "responseData", responseData)
		return response.ErrorResponse{
			ErrorCode:       constant.INTERNAL_SERVER_ERROR,
			StatusCode:      http.StatusInternalServerError,
			LogMessage:      "The ledger responded with an empty result object",
			MaybeInnerError: errtrace.New(""),
		}
	}

	if responseData.Result.CustomerID != user.LedgerCustomerNumber {
		logger.Error("user requested transaction that does not belong to them")
		return response.ErrorResponse{
			ErrorCode:       constant.INTERNAL_SERVER_ERROR,
			StatusCode:      http.StatusInternalServerError,
			LogMessage:      "user requested transaction that does not belong to them",
			MaybeInnerError: errtrace.New(""),
		}
	}

	if slices.Contains(NonDisputableTransactionTypes, responseData.Result.Type) || slices.Contains(AdministrativeTransactionTypes, responseData.Result.Type) {
		logger.Error("Cannot request a dispute for this transaction type", "type", responseData.Result.Type)
		return response.ErrorResponse{ErrorCode: constant.DISPUTE_REQUEST_INVALID_TRANSACTION_TYPE, Message: constant.DISPUTE_REQUEST_INVALID_TRANSACTION_TYPE_MSG, StatusCode: http.StatusUnprocessableEntity, MaybeInnerError: errtrace.New("")}
	}

	id := uuid.New().String()

	transactionDisputes := dao.TransactionDisputeDao{
		Id:                    id,
		Status:                PENDING,
		TransactionIdentifier: referenceId,
		Reason:                requestData.Reason,
		Details:               requestData.Details,
		UserId:                userId,
	}

	err = db.DB.Select("id", "status", "transaction_identifier", "reason", "details", "user_id").Create(&transactionDisputes).Error
	if err != nil {
		// check if dispute already exists
		if pgErr, ok := err.(*pq.Error); ok {
			if pgErr.Code == "23505" {
				logger.Error("Dispute request already exists for this transaction", "transaction_identifier", referenceId)
				return response.ErrorResponse{ErrorCode: constant.DISPUTE_REQUEST_ALREADY_EXISTS, Message: constant.DISPUTE_REQUEST_ALREADY_EXISTS_MSG, StatusCode: http.StatusConflict, MaybeInnerError: errtrace.Wrap(err)}
			}
		}
		logger.Error("Error while creating transaction dispute record", "error", err.Error())
		return response.ErrorResponse{
			ErrorCode:       constant.INTERNAL_SERVER_ERROR,
			StatusCode:      http.StatusInternalServerError,
			LogMessage:      fmt.Sprintf("Error while creating transaction dispute record: %s", err.Error()),
			MaybeInnerError: errtrace.Wrap(err),
		}
	}

	return c.JSON(http.StatusCreated, SubmitTransactionDisputeResponse{
		Status:    transactionDisputes.Status,
		CreatedAt: transactionDisputes.CreatedAt.Format(time.RFC3339),
	})
}

type SubmitTransactionDisputeResponse struct {
	Status    string `json:"status" validate:"required" enums:"pending"`
	CreatedAt string `json:"createdAt" validate:"required"`
}

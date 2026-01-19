package handler

import (
	"fmt"
	"net/http"
	"process-api/pkg/config"
	"process-api/pkg/db/dao"
	"process-api/pkg/ledger"
	"process-api/pkg/logging"
	"process-api/pkg/model/request"
	"process-api/pkg/model/response"

	"braces.dev/errtrace"
	"github.com/labstack/echo/v4"
)

// @summary IsEmailDuplicate
// @description Checks if an email is already registered
// @tags onboarding
// @accept json
// @produce json
// @param emailDuplicateRequest body request.EmailDuplicateRequest true "EmailDuplicate payload"
// @Success 200 {object} response.EmailDuplicateResponse
// @failure 500 {object} response.ErrorResponse
// @failure 400 {object} response.BadRequestErrors
// @router /onboarding/emailDuplicate [post]
func IsEmailDuplicate(c echo.Context) error {
	// check if email already exists in middleware
	var requestData request.EmailDuplicateRequest
	if err := c.Bind(&requestData); err != nil {
		return response.BadRequestInvalidBody
	}

	// validations
	if err := c.Validate(requestData); err != nil {
		return err
	}

	user, err := dao.MasterUserRecordDao{}.FindUserByEmail(requestData.Email)
	if err != nil {
		// Handle Other DB error's
		logging.Logger.Error("Error while fetching user record", "error", err.Error())
		return response.InternalServerError(fmt.Sprintf("Error while fetching user record :%s", err.Error()), errtrace.Wrap(err))
	}

	IsEmailDuplicate := user != nil

	if !IsEmailDuplicate {
		ledgerClient := ledger.CreateLedgerApiClient(config.Config.Ledger)

		req := ledger.BuildGetCustomerByContactPayload(requestData.Email, "")
		resp, err := ledgerClient.GetCustomer(req)
		if err != nil {
			return response.InternalServerError(fmt.Sprintf("Error while calling GetCustomer: %s", err.Error()), errtrace.Wrap(err))
		}
		if resp.Error != nil && resp.Error.Code != "NOT_FOUND_CUSTOMER" {
			return response.InternalServerError(fmt.Sprintf("Ledger GetCustomer responded with error: %s %s", resp.Error.Code, resp.Error.Message), errtrace.New(""))
		} else {
			IsEmailDuplicate = resp.Result != nil && resp.Result.Id != ""
		}

	}

	response := response.EmailDuplicateResponse{
		IsEmailDuplicate: IsEmailDuplicate,
	}
	return c.JSON(http.StatusOK, response)
}

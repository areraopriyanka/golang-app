package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"process-api/pkg/config"
	"process-api/pkg/constant"
	"process-api/pkg/db"
	"process-api/pkg/db/dao"
	"process-api/pkg/debtwise"
	"process-api/pkg/logging"
	"process-api/pkg/model/response"
	"process-api/pkg/security"
	"strings"

	"braces.dev/errtrace"
	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// @Summary CreateDebtwiseUser
// @Description Creates a debtwise user and then stores the ID on the user record
// @Tags credit score
// @Accept json
// @Produce json
// @Success 200 "OK"
// @Success 201 "Created"
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /account/debtwise/user [post]
func (h *Handler) CreateDebtwiseUser(c echo.Context) error {
	cc, ok := c.(*security.LoggedInRegisteredUserContext)
	if !ok {
		return response.UnauthorizedError("Failed to get user Id from custom context")
	}

	userId := cc.UserId

	logger := logging.GetEchoContextLogger(cc)

	debtwiseClient, err := debtwise.NewDebtwiseClient(config.Config.Debtwise, logger)
	if err != nil {
		logger.Error("Error occurred while creating debtwise client", "error", err.Error())
		return response.InternalServerError(fmt.Sprintf("Error occurred while creating debtwise client: %s", err.Error()), errtrace.Wrap(err))
	}

	user, errResponse := dao.RequireUserWithState(userId, constant.ACTIVE)
	if errResponse != nil {
		return errResponse
	}

	// NOTE: Do not attempt to create debtwise user if a customer number already exists
	if user.DebtwiseCustomerNumber != nil {
		logger.Info("Debtwise customer number already exists for user", "userId", userId)
		return cc.NoContent(http.StatusOK)
	}

	var firstName string
	var lastName string
	var ssn string
	if h.Env != constant.PROD {
		firstName = "Abby"
		lastName = "CaineenUAT"
		ssn = "00000000"
		logger.Info("Overriding values for create debtwise user call", "firstName", firstName, "lastName", lastName, "ssn", ssn)
	} else {
		firstName = user.Email
		lastName = user.LastName
		ssn = ""
	}

	dateOfBirth := openapi_types.Date{Time: user.DOB}
	phoneNumber := strings.ReplaceAll(user.MobileNo, "-", "")

	debtwiseRequest := debtwise.CreateUserJSONRequestBody{
		DateOfBirth: dateOfBirth,
		PhoneNumber: phoneNumber,
		Email:       user.Email,
		FirstName:   firstName,
		LastName:    lastName,
		Ssn:         ssn,
	}

	jsonBytes, err := json.Marshal(debtwiseRequest)
	if err != nil {
		logger.Error("Error occurred while marshalling request body: " + err.Error())
		return response.InternalServerError(fmt.Sprintf("Error occurred while marshalling request body: %s", err), errtrace.Wrap(err))
	}

	debtwiseResponse, err := debtwiseClient.CreateUserWithBodyWithResponse(
		context.Background(),
		nil,
		"application/json",
		bytes.NewReader(jsonBytes),
	)
	if err != nil {
		logger.Error("Error occurred while creating debtwise user", "debtwiseError", err.Error())
		return response.InternalServerError(fmt.Sprintf("Error occurred while creating debtwise user: debtwiseError: %s", err.Error()), errtrace.Wrap(err))
	}

	if debtwiseResponse.JSON201 != nil {
		debtwiseUser := debtwiseResponse.JSON201

		updateResult := db.DB.Model(user).Update("debtwise_customer_number", debtwiseUser.Id)
		if updateResult.Error != nil {
			return response.InternalServerError(updateResult.Error.Error(), errtrace.Wrap(err))
		}

		if updateResult.RowsAffected == 0 {
			return response.InternalServerError("Could not update debtwise customer number", errtrace.Wrap(err))
		}
	} else if debtwiseResponse.JSON422 != nil {
		logger.Error("Received a 422 response from debtwise during user creation", "errorMessage", debtwiseResponse.JSON422.Error.Message)
		return response.InternalServerError("Received unprocessable entity error from Debtwise", errtrace.Wrap(err))
	} else {
		logger.Error("Unexpected Debtwise response", "status", debtwiseResponse.HTTPResponse.StatusCode)
		return response.InternalServerError("Unexpected response from Debtwise", errtrace.Wrap(err))
	}

	logger.Info("Successfully created debtwise user", "userId", userId)
	return cc.NoContent(http.StatusCreated)
}

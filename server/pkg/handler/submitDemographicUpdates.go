package handler

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"process-api/pkg/constant"
	"process-api/pkg/db"
	"process-api/pkg/db/dao"
	"process-api/pkg/logging"
	"process-api/pkg/model/response"
	"process-api/pkg/security"

	"braces.dev/errtrace"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

const (
	PENDING  = "pending"
	ACCEPTED = "accepted"
	REJECTED = "rejected"
)

const (
	FULL_NAME = "full_name"
	ADDRESS   = "address"
)

// @summary SubmitFullNameDemographicUpdates
// @description Initiates a full name demographic update request and stores the update data in the middleware database
// @tags Demographic Updates
// @accept json
// @param payload body dao.UpdateFullNameRequest true "Payload with fullName, lastName and suffix."
// @Param Authorization header string true "Bearer token for user authentication"
// @success 201 "Created"
// @header 201 {string} Authorization "Bearer token for user authentication"
// @Failure 400 {object} response.BadRequestErrors
// @failure 401 {object} response.ErrorResponse
// @failure 409 {object} response.ErrorResponse
// @failure 404 {object} response.ErrorResponse
// @failure 412 {object} response.ErrorResponse
// @failure 500 {object} response.ErrorResponse
// @router /account/customer/demographic-update/full-name [post]
func SubmitFullNameDemographicUpdates(c echo.Context) error {
	cc, ok := c.(*security.LoggedInRegisteredUserContext)
	if !ok {
		return response.UnauthorizedError("Failed to get user Id from custom context")
	}
	userId := cc.UserId

	logger := logging.GetEchoContextLogger(c)

	_, errResponse := dao.RequireUserWithState(userId, constant.ACTIVE)
	if errResponse != nil {
		return errResponse
	}

	requestData := new(dao.UpdateFullNameRequest)
	if err := c.Bind(requestData); err != nil {
		return response.BadRequestInvalidBody
	}

	if err := c.Validate(requestData); err != nil {
		return err
	}

	rawValue, err := json.Marshal(requestData)
	if err != nil {
		logger.Error("Error while marshaling value", "error", err.Error())
	}

	err = createDemographicUpdateData(json.RawMessage(rawValue), FULL_NAME, userId, logger)
	if err != nil {
		return err
	}

	logger.Info("Demographic update request submitted successfully", "field", FULL_NAME, "userId", userId)
	return c.NoContent(http.StatusCreated)
}

// @summary SubmitAddressDemographicUpdates
// @description Initiates a address demographic update request and stores the update data in the middleware database
// @tags Demographic Updates
// @accept json
// @param payload body dao.UpdateCustomerAddressRequest true "Payload with address fields"
// @Param Authorization header string true "Bearer token for user authentication"
// @success 201 "Created"
// @header 201 {string} Authorization "Bearer token for user authentication"
// @Failure 400 {object} response.BadRequestErrors
// @failure 401 {object} response.ErrorResponse
// @failure 404 {object} response.ErrorResponse
// @failure 409 {object} response.ErrorResponse
// @failure 412 {object} response.ErrorResponse
// @failure 500 {object} response.ErrorResponse
// @router /account/customer/demographic-update/address [post]
func SubmitAddressDemographicUpdates(c echo.Context) error {
	cc, ok := c.(*security.LoggedInRegisteredUserContext)
	if !ok {
		return response.UnauthorizedError("Failed to get user Id from custom context")
	}
	userId := cc.UserId

	logger := logging.GetEchoContextLogger(c)

	_, errResponse := dao.RequireUserWithState(userId, constant.ACTIVE)
	if errResponse != nil {
		return errResponse
	}

	requestData := new(dao.UpdateCustomerAddressRequest)
	if err := c.Bind(requestData); err != nil {
		return response.BadRequestInvalidBody
	}

	if err := c.Validate(requestData); err != nil {
		return err
	}

	rawValue, err := json.Marshal(requestData)
	if err != nil {
		logger.Error("Error while marshaling value", "error", err.Error())
	}

	err = createDemographicUpdateData(json.RawMessage(rawValue), ADDRESS, userId, logger)
	if err != nil {
		return err
	}

	logger.Info("Demographic update request submitted successfully", "field", ADDRESS, "userId", userId)
	return c.NoContent(http.StatusCreated)
}

func createDemographicUpdateData(value json.RawMessage, demographicUpdateType, userId string, logger *slog.Logger) error {
	id := uuid.New().String()
	demographicUpdate := dao.DemographicUpdatesDao{
		Id:           id,
		Type:         demographicUpdateType,
		Status:       PENDING,
		UpdatedValue: value,
		UserId:       userId,
	}

	// Check if a pending request of the same `type` already exists for the given user ID
	demographicUpdateRecord, err := dao.DemographicUpdatesDao{}.FindByTypeAndUserId(userId, demographicUpdateType)
	if err != nil {
		logger.Error("Error while fetching demographic update record", "error", err.Error())
		return response.ErrorResponse{
			ErrorCode:       constant.INTERNAL_SERVER_ERROR,
			StatusCode:      http.StatusInternalServerError,
			LogMessage:      fmt.Sprintf("Error while fetching demographic update record: %s", err.Error()),
			MaybeInnerError: errtrace.Wrap(err),
		}
	}

	// If record already exists
	if demographicUpdateRecord != nil {
		return response.ErrorResponse{ErrorCode: constant.DEMOGRAPHIC_UPDATE_REQUEST_ALREADY_EXISTS, Message: constant.DEMOGRAPHIC_UPDATE_REQUEST_ALREADY_EXISTS_MSG, StatusCode: http.StatusConflict, MaybeInnerError: errtrace.New("")}
	}

	result := db.DB.Select("id", "type", "status", "updated_value", "user_id").Create(&demographicUpdate)
	if result.Error != nil {
		logger.Error("Error while creating demographic update record", "error", result.Error.Error())
		return response.ErrorResponse{
			ErrorCode:       constant.INTERNAL_SERVER_ERROR,
			StatusCode:      http.StatusInternalServerError,
			LogMessage:      fmt.Sprintf("Error while creating demographic update record: %s", result.Error.Error()),
			MaybeInnerError: errtrace.Wrap(err),
		}
	}
	return nil
}

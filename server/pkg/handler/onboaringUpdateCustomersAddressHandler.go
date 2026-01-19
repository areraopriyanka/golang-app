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

// @summary UpdateCustomerAddress
// @description Updating customer's address and setting customer's status
// @tags onboarding
// @accept json
// @produce json
// @Param userId path string true "User ID"
// @Param Authorization header string true "Bearer token for user authentication"
// @param updateCustomerAddressRequest body dao.UpdateCustomerAddressRequest true "UpdateCustomerAddressRequest payload"
// @Success 200 "OK"
// @header 200 {string} Authorization "Bearer token for user authentication"
// @failure 400 {object} response.BadRequestErrors
// @failure 401 {object} response.ErrorResponse
// @failure 404 {object} response.ErrorResponse
// @failure 412 {object} response.ErrorResponse
// @failure 500 {object} response.ErrorResponse
// @router /onboarding/customer/{userId}/address [post]
// @router /onboarding/customer/address [post]
func UpdateCustomerAddress(c echo.Context) error {
	// TODO: Remove path param logic later and use userId from JWT only.
	userId := c.Param("userId")
	if userId == "" {
		cc, ok := c.(*security.OnboardingUserContext)
		if !ok {
			return response.ErrorResponse{ErrorCode: constant.UNAUTHORIZED_ACCESS_ERROR, Message: constant.UNAUTHORIZED_ACCESS_ERROR_MSG, StatusCode: http.StatusUnauthorized, LogMessage: "Invalid type of custom context", MaybeInnerError: errtrace.New("")}
		}

		userId = cc.UserId
	}

	logger := logging.GetEchoContextLogger(c)

	request := new(dao.UpdateCustomerAddressRequest)
	err := c.Bind(request)
	if err != nil {
		return response.BadRequestInvalidBody
	}

	if err := c.Validate(request); err != nil {
		return err
	}

	user, errResponse := dao.RequireUserWithState(
		userId, constant.PHONE_NUMBER_VERIFIED, constant.ADDRESS_CONFIRMED,
	)
	if errResponse != nil {
		return errResponse
	}

	// Update userState only in case of positive flow without back navigation
	updateData := map[string]interface{}{
		"street_address": request.StreetAddress,
		"apartment_no":   request.ApartmentNo,
		"zip_code":       request.ZipCode,
		"city":           request.City,
		"state":          request.State,
	}

	if user.UserStatus == constant.PHONE_NUMBER_VERIFIED {
		updateData["user_status"] = constant.ADDRESS_CONFIRMED
	}

	result := db.DB.Model(&dao.MasterUserRecordDao{}).Where("id=?", userId).Updates(updateData)
	// Handle DB errors
	if result.Error != nil {
		return response.ErrorResponse{ErrorCode: constant.INTERNAL_SERVER_ERROR, Message: constant.INTERNAL_SERVER_ERROR_MSG, StatusCode: http.StatusInternalServerError, LogMessage: fmt.Sprintf("Error in updating user's state and address: %s", result.Error.Error()), MaybeInnerError: errtrace.Wrap(result.Error)}
	}

	logger.Info("Address updated successfully for user", "userId", userId)
	return c.NoContent(http.StatusOK)
}

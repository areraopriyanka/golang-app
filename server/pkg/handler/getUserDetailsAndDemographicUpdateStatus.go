package handler

import (
	"fmt"
	"net/http"
	"process-api/pkg/constant"
	"process-api/pkg/db/dao"
	"process-api/pkg/logging"
	"process-api/pkg/model/response"
	"process-api/pkg/security"

	"braces.dev/errtrace"
	"github.com/labstack/echo/v4"
)

// @summary GetUserDetailsAndDemographicUpdateStatus
// @description Returns the user's personal details along with the latest demographic update status for each type
// @tags Demographic Updates
// @produce json
// @param Authorization header string true "Bearer token for user authentication"
// @success 200 {object} response.GetUserDetailsAndUpdateStatus
// @header 200 {string} Authorization "Bearer token for user authentication"
// @failure 401 {object} response.ErrorResponse
// @failure 404 {object} response.ErrorResponse
// @failure 412 {object} response.ErrorResponse
// @failure 500 {object} response.ErrorResponse
// @router /account/customer/demographic-update [get]
func GetUserDetailsAndDemographicUpdateStatus(c echo.Context) error {
	cc, ok := c.(*security.LoggedInRegisteredUserContext)
	if !ok {
		return response.UnauthorizedError("Failed to get user Id from custom context")
	}
	userId := cc.UserId

	logger := logging.GetEchoContextLogger(c)

	user, errResponse := dao.RequireUserWithState(userId, constant.ACTIVE)
	if errResponse != nil {
		return errResponse
	}

	demographicUpdates, err := dao.DemographicUpdatesDao{}.FindDemographicUpdatesByUserId(user.Id)
	if err != nil {
		logger.Error("Received error while querying demographic updates", "err", err.Error())
		return response.InternalServerError(fmt.Sprintf("Received error while querying demographic updates: %s", err.Error()), errtrace.Wrap(err))
	}

	if demographicUpdates == nil {
		logger.Info("No record found in demographic updates table")
	}

	statusMap := make(map[string]string, len(demographicUpdates))
	for _, demographicUpdate := range demographicUpdates {
		statusMap[demographicUpdate.Type] = demographicUpdate.Status
	}

	userDetailsAndUpdateStatus := response.GetUserDetailsAndUpdateStatus{
		FirstName:    user.FirstName,
		LastName:     user.LastName,
		Suffix:       user.Suffix,
		Email:        user.Email,
		MobileNumber: user.MobileNo,
		Address: response.AddressDetails{
			AddressLine1: user.StreetAddress,
			AddressLine2: user.ApartmentNo,
			City:         user.City,
			State:        user.State,
			PostalCode:   user.ZipCode,
		},
		FullNameStatus: statusMap["full_name"],
		AddressStatus:  statusMap["address"],
	}

	return c.JSON(http.StatusOK, userDetailsAndUpdateStatus)
}

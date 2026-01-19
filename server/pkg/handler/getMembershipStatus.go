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

// @Summary GetMemberShipStatus
// @Description Returns user's membership status.
// @Tags membership
// @Produce json
// @Param Authorization header string true "Bearer token for user authentication"
// @Success 200 {object} response.GetMemberShipStatus
// @header 200 {string} Authorization "Bearer token for user authentication"
// @Failure 500 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 412 {object} response.ErrorResponse
// @Router /account/membership/status [get]
func GetMemberShipStatus(c echo.Context) error {
	cc, ok := c.(*security.LoggedInRegisteredUserContext)
	if !ok {
		return response.ErrorResponse{ErrorCode: constant.UNAUTHORIZED_ACCESS_ERROR, Message: constant.UNAUTHORIZED_ACCESS_ERROR_MSG, StatusCode: http.StatusUnauthorized, LogMessage: "Failed to get user Id from custom context"}
	}

	userId := cc.UserId

	logger := logging.GetEchoContextLogger(c)

	_, errResponse := dao.RequireUserWithState(
		userId, constant.ACTIVE,
	)
	if errResponse != nil {
		return errResponse
	}

	userMembershipRecord, err := dao.UserMembershipDao{}.FindOneByUserId(userId)
	if err != nil {
		logger.Error("could not retrieve user membership record", "error", err.Error())
		return response.InternalServerError(fmt.Sprintf("Error occurred while finding user record: error: %s", err.Error()), errtrace.Wrap(err))
	}

	if userMembershipRecord == nil {
		logger.Error("No record found")
		return response.InternalServerError("Record not found: error", errtrace.New(""))
	}

	return c.JSON(http.StatusOK, response.GetMemberShipStatus{
		Status: userMembershipRecord.MembershipStatus,
	})
}

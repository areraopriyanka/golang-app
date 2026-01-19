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

// @Summary AcceptMembership
// @Description Accept membership.
// @Tags membership
// @Param Authorization header string true "Bearer token for user authentication"
// @Success 200 "OK"
// @header 200 {string} Authorization "Bearer token for user authentication"
// @Failure 500 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 412 {object} response.ErrorResponse
// @Router /onboarding/customer/membership [post]
func AcceptMembership(c echo.Context) error {
	cc, ok := c.(*security.OnboardingUserContext)
	if !ok {
		return response.UnauthorizedError("Failed to get user Id from custom context")
	}
	userId := cc.UserId

	logger := logging.GetEchoContextLogger(c)

	_, errResponse := dao.RequireUserWithState(
		userId, constant.KYC_PASS,
	)
	if errResponse != nil {
		return errResponse
	}

	err := updateUserStatus(userId, constant.MEMBERSHIP_ACCEPTED)
	if err != nil {
		logger.Error("Error while updating user's status", "error", err.Error())
		return response.InternalServerError(fmt.Sprintf("Error while updating user's status: error: %s", err.Error()), errtrace.Wrap(err))
	}
	return c.NoContent(http.StatusOK)
}

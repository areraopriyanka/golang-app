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

type UpdateMembershipStatusRequest struct {
	Action string `json:"action" validate:"required,oneof=subscribed unsubscribed"`
}

// @Summary UpdateMembershipStatus
// @Description Update membership status['subscribed', unsubscribed].
// @Tags membership
// @Param Authorization header string true "Bearer token for user authentication"
// @Success 200 "OK"
// @header 200 {string} Authorization "Bearer token for user authentication"
// @failure 400 {object} response.BadRequestErrors
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 412 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /account/membership/status [put]
func UpdateMembershipStatus(c echo.Context) error {
	cc, ok := c.(*security.LoggedInRegisteredUserContext)
	if !ok {
		return response.UnauthorizedError("Failed to get user Id from custom context")
	}
	userId := cc.UserId

	logger := logging.GetEchoContextLogger(c)

	request := new(UpdateMembershipStatusRequest)

	_, errResponse := dao.RequireUserWithState(userId, constant.ACTIVE)
	if errResponse != nil {
		return errResponse
	}

	err := c.Bind(request)
	if err != nil {
		return response.BadRequestInvalidBody
	}

	err = c.Validate(request)
	if err != nil {
		return err
	}

	err = updateMembershipStatus(userId, request.Action)
	if err != nil {
		logger.Error("Error while updating membership status", "error", err)
		return response.InternalServerError(fmt.Sprintf("Error while updating membership status: %s", err.Error()), errtrace.Wrap(err))
	}

	logger.Debug("membership status updated to", "status", request.Action)
	return c.NoContent(http.StatusOK)
}

func updateMembershipStatus(userId, status string) error {
	result := db.DB.Model(dao.UserMembershipDao{}).Where("user_id=?", userId).Update("membership_status", status)
	if result.Error != nil {
		return errtrace.Wrap(result.Error)
	}
	if result.RowsAffected == 0 {
		return errtrace.Wrap(fmt.Errorf("could not update membership status"))
	}
	return nil
}

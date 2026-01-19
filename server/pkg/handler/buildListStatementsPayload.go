package handler

import (
	"net/http"
	"process-api/pkg/constant"
	"process-api/pkg/db/dao"
	"process-api/pkg/ledger"
	"process-api/pkg/logging"
	"process-api/pkg/model/request"
	"process-api/pkg/model/response"
	"process-api/pkg/security"

	"github.com/labstack/echo/v4"
)

// @Summary BuildListStatementsPayload
// @Description Create BuildListStatements payload
// @Tags statement
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token for user authentication"
// @Param payload body request.BuildListStatementsPayloadRequest true "payload with pageSize and pageNumber"
// @Success 200 {object} response.BuildPayloadResponse
// @Header 200 {string} Authorization "Bearer token for user authentication"
// @Failure 400 {object} response.BadRequestErrors
// @Failure 401 {object} response.ErrorResponse
// @Failure 404  {object} response.ErrorResponse
// @Failure 412 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /account/list-statements/build [post]
func BuildListStatementsPayload(c echo.Context) error {
	cc, ok := c.(*security.LoggedInRegisteredUserContext)
	if !ok {
		return response.UnauthorizedError("Failed to get user Id from custom context")
	}

	userId := cc.UserId

	_, errResponse := dao.RequireUserWithState(userId, constant.ACTIVE)
	if errResponse != nil {
		return errResponse
	}

	logger := logging.GetEchoContextLogger(c)

	requestData := new(request.BuildListStatementsPayloadRequest)
	err := c.Bind(requestData)
	if err != nil {
		return response.BadRequestInvalidBody
	}

	if err := c.Validate(requestData); err != nil {
		return err
	}

	payload := ledger.ListStatementRequest{
		PageNumber: requestData.PageNumber,
		PageSize:   requestData.PageSize,
	}

	payloadResponse, errResponse := dao.CreateSignablePayloadForUser(userId, payload)
	if errResponse != nil {
		return errResponse
	}

	logger.Debug("ListStatements payload created successfully for user", "userId", userId)
	return c.JSON(http.StatusOK, payloadResponse)
}

package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"process-api/pkg/constant"
	"process-api/pkg/logging"
	"process-api/pkg/model/response"
	"process-api/pkg/security"
	"strconv"

	"braces.dev/errtrace"
	"github.com/labstack/echo/v4"
)

// @Summary Get Job Status By Job ID
// @Description Returns status of river job given job id
// @Tags jobs
// @Accept json
// @Produce json
// @param Authorization header string true "Bearer token for user authentication"
// @Param jobId path int true "Job Id"
// @Success 200 {object} JobStatusResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /account/status/jobs/{jobId} [get]
func (h *Handler) JobStatusHandler(c echo.Context) error {
	cc, ok := c.(*security.LoggedInRegisteredUserContext)
	if !ok {
		return response.UnauthorizedError("Failed to get user Id from custom context")
	}

	userId := cc.UserId

	logger := logging.GetEchoContextLogger(c)

	jobIdStr := c.Param("jobId")
	jobId, err := strconv.Atoi(jobIdStr)
	if err != nil {
		return response.BadRequestInvalidBody
	}

	ctx := context.Background()

	job, err := h.RiverClient.JobGet(ctx, int64(jobId))
	if err != nil {
		logger.Error("Failed to fetchjob status", "jobId", jobId, "error", err.Error())
	}

	var jobArgs struct {
		UserId string `json:"user_id"`
	}
	if err := json.Unmarshal(job.EncodedArgs, &jobArgs); err != nil {
		logger.Error("Failed to decode job args", "error", err.Error())
		return response.ErrorResponse{ErrorCode: constant.INTERNAL_SERVER_ERROR, Message: "Failed to decode job args", StatusCode: http.StatusInternalServerError, MaybeInnerError: errtrace.Wrap(err)}
	}

	// NOTE: The 404 status code is a little ambiguous here but we don't want to return a 401 and potentially provide insight into
	// the sequential id incremenetation of job IDs
	if userId != jobArgs.UserId {
		return response.ErrorResponse{ErrorCode: constant.NO_DATA_FOUND, Message: constant.NO_DATA_FOUND_MSG, StatusCode: http.StatusNotFound, MaybeInnerError: errtrace.New("")}
	}

	jobState := string(job.State)

	response := JobStatusResponse{
		State: jobState,
	}

	return cc.JSON(http.StatusOK, response)
}

type JobStatusResponse struct {
	State string `json:"state" validate:"required,oneof=available cancelled completed discarded pending retryable running scheduled"`
}

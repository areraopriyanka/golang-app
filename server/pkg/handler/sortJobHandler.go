package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"process-api/pkg/constant"
	"process-api/pkg/logging"
	"process-api/pkg/model/request"
	"process-api/pkg/model/response"
	"sort"
	"time"

	"braces.dev/errtrace"
	"github.com/labstack/echo/v4"
	"github.com/riverqueue/river"
)

type SortArgs struct {
	// Strings is a slice of strings to sort.
	Strings []string `json:"strings"`
}

func (SortArgs) Kind() string { return "sort" }

type SortWorker struct {
	// An embedded WorkerDefaults sets up default methods to fulfill the rest of
	// the Worker interface:
	river.WorkerDefaults[SortArgs]
}

func (w *SortWorker) Work(ctx context.Context, job *river.Job[SortArgs]) error {
	sort.Strings(job.Args.Strings)

	output := map[string]interface{}{
		"sortedStrings": job.Args.Strings,
	}
	if err := river.RecordOutput(ctx, output); err != nil {
		logging.Logger.Error("Error recording sort job output", "err", err)
		return fmt.Errorf("failed to record job output: %w", err)
	}
	logging.Logger.Info("Successfully sorted strings", "sortedStrings", job.Args.Strings)
	return nil
}

func RegisterSortWorker(workers *river.Workers) {
	river.AddWorker(workers, &SortWorker{})
}

func (h *Handler) StartSortJob(c echo.Context) error {
	logger := logging.GetEchoContextLogger(c)

	var requestData SortJobStartRequest

	if err := c.Bind(&requestData); err != nil {
		logger.Error("Invalid request", "error", err.Error())
		return response.BadRequestErrors{
			Errors: []response.BadRequestError{
				{Error: err.Error()},
			},
		}
	}

	if err := c.Validate(requestData); err != nil {
		return err
	}

	ctx := context.Background()
	// Enqueue a job to sort strings
	jobInsertResult, err := h.RiverClient.Insert(ctx, SortArgs{
		Strings: requestData.StringsToSort,
	}, nil)
	if err != nil {
		logger.Error("Failed to enqueue sort job", "error", err)
		return response.ErrorResponse{ErrorCode: constant.INTERNAL_SERVER_ERROR, StatusCode: http.StatusInternalServerError, LogMessage: fmt.Sprintf("Failed to enqueue sort job: error: %s", err.Error()), MaybeInnerError: errtrace.Wrap(err)}
	} else {
		logger.Info("Enqueued sort job successfully")
	}

	var payload *SortJobOutput = nil
	jobId := jobInsertResult.Job.ID
	var jobState string

	// Poll job status up to 3 times, every 500ms, include payload in response if job is complete
	for i := 0; i < 3; i++ {
		job, err := h.RiverClient.JobGet(ctx, jobId)
		if err != nil {
			logger.Error("Failed to fetch sort job status", "jobId", jobId, "error", err.Error())
			break
		}
		jobState = string(job.State)
		if job.State == "completed" && job.Output() != nil {
			var output SortJobOutput
			if err := json.Unmarshal(job.Output(), &output); err == nil {
				payload = &output
			}
			break
		}
		time.Sleep(500 * time.Millisecond)
	}

	response := response.MaybeRiverJobResponse[SortJobOutput]{
		JobId:   int(jobInsertResult.Job.ID),
		State:   jobState,
		Payload: payload,
	}

	return c.JSON(http.StatusOK, response)
}

func (h *Handler) GetSortJobStatus(c echo.Context) error {
	logger := logging.GetEchoContextLogger(c)

	var requestData request.JobStatusRequest

	if err := c.Bind(&requestData); err != nil {
		logger.Error("Invalid request", "error", err.Error())
		return response.BadRequestErrors{
			Errors: []response.BadRequestError{
				{Error: err.Error()},
			},
		}
	}

	if err := c.Validate(requestData); err != nil {
		return err
	}

	ctx := context.Background()
	job, err := h.RiverClient.JobGet(ctx, int64(requestData.JobId))
	if err != nil {
		logger.Error("Failed to fetch job status", "jobId", requestData.JobId, "error", err.Error())
		return response.ErrorResponse{ErrorCode: constant.INTERNAL_SERVER_ERROR, StatusCode: http.StatusInternalServerError, LogMessage: fmt.Sprintf("Failed to fetch job status: jobId: %d, error: %s", requestData.JobId, err.Error()), MaybeInnerError: errtrace.Wrap(err)}
	}

	var output SortJobOutput
	if job.Output() != nil {
		if err := json.Unmarshal(job.Output(), &output); err != nil {
			logger.Error("Failed to decode job output", "jobId", requestData.JobId, "error", err.Error())
			return response.ErrorResponse{ErrorCode: constant.INTERNAL_SERVER_ERROR, StatusCode: http.StatusInternalServerError, LogMessage: fmt.Sprintf("Failed to decode job output: jobId: %d, error: %s", requestData.JobId, err.Error()), MaybeInnerError: errtrace.Wrap(err)}
		}
	}

	sortJobResponse := response.MaybeRiverJobResponse[SortJobOutput]{
		JobId:   int(job.ID),
		State:   string(job.State),
		Payload: &output,
	}

	return c.JSON(http.StatusOK, sortJobResponse)
}

type SortJobStartRequest struct {
	StringsToSort []string `json:"stringsToSort" validate:"required"`
}

type SortJobOutput struct {
	SortedStrings []string `json:"sortedStrings"`
}

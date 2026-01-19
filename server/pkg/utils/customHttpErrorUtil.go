package utils

import (
	"fmt"
	"net/http"
	"process-api/pkg/logging"
	"process-api/pkg/model/response"
	"strconv"

	"braces.dev/errtrace"
	"github.com/labstack/echo/v4"
)

func CustomHTTPErrorHandler(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}

	logger := logging.GetEchoContextLogger(c)

	var (
		errorResponse   error
		statusCode      int
		logMessage      string
		stackTrace      string
		skipPosthogCall bool
	)

	switch e := err.(type) {
	case *response.ErrorResponse:
		errorResponse = *e
		statusCode = e.StatusCode
		logMessage = e.LogMessage
		if e.MaybeInnerError != nil {
			stackTrace = errtrace.FormatString(e.MaybeInnerError)
		}
	case response.ErrorResponse:
		errorResponse = e
		statusCode = e.StatusCode
		logMessage = e.LogMessage
		if e.MaybeInnerError != nil {
			stackTrace = errtrace.FormatString(e.MaybeInnerError)
		}
	case response.BadRequestErrors:
		errorResponse = e
		statusCode = http.StatusBadRequest
		logMessage = fmt.Sprintf("Validation failure: %s", e.Error())
	case *echo.HTTPError:
		errorResponse = response.ErrorResponse{
			ErrorCode: strconv.Itoa(e.Code),
			Message:   fmt.Sprintf("%v", e.Message),
		}
		statusCode = e.Code
		logMessage = err.Error()
		skipPosthogCall = true
	default:
		// Default error response for unexpected errors
		errorResponse = response.ErrorResponse{
			ErrorCode: "INTERNAL_SERVER_ERROR",
			Message:   "An unexpected error occurred",
		}
		statusCode = http.StatusInternalServerError
		logMessage = err.Error()
		if err != nil {
			stackTrace = errtrace.FormatString(err)
		}
	}

	// Log the error
	if logger != nil {
		logger.Error(logMessage)
	}

	if !skipPosthogCall {
		PosthogClient.SendErrorToPosthog(c, err, stackTrace)
	}

	if err := c.JSON(statusCode, errorResponse); err != nil {
		logger.Error("Failed to send JSON response", "error", err)
	}
}

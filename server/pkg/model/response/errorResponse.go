package response

import (
	"errors"
	"fmt"
	"net/http"
	"process-api/pkg/constant"
	"process-api/pkg/model"
	"strings"

	"braces.dev/errtrace"
	"github.com/jinzhu/gorm"
)

// TODO: Some APIs still directly use ErrorResponse{}.
// We've added proper error messages and wrapped errors which will help in debugging.
// These can later be updated to use GenerateErrResponse().

// NOTE: Going forward, we should avoid directly using ErrorResponse{}.
// Instead, use GenerateErrResponse() and ensure the actual error message is passed in either the Message or LogMessage field for debugging purposes in posthog.
type ErrorResponse struct {
	ErrorCode       string `json:"code"`
	Message         string `json:"msg"`
	StatusCode      int    `json:"-"`
	LogMessage      string `json:"-"`
	MaybeInnerError error  `json:"-"`
}

func (e ErrorResponse) Error() string {
	return fmt.Sprintf("%d: (%s) %s - %s", e.StatusCode, e.ErrorCode, e.Message, e.LogMessage)
}

func GenerateErrResponse(code string, message string, logMessage string, statusCode int, err error) ErrorResponse {
	response := ErrorResponse{
		ErrorCode:       code,
		Message:         message,
		StatusCode:      statusCode,
		LogMessage:      logMessage,
		MaybeInnerError: err,
	}
	return response
}

type BadRequestErrors struct {
	Errors []BadRequestError `json:"errors" validate:"required"`
}

var BadRequestInvalidBody = BadRequestErrors{
	Errors: []BadRequestError{
		{
			Error: "invalid_body",
		},
	},
}

type BadRequestError struct {
	FieldName string `json:"fieldName"`
	Error     string `json:"error" validate:"required"`
}

func (e BadRequestErrors) Error() string {
	var builder strings.Builder

	builder.WriteString("Bad request.")
	for _, error := range e.Errors {
		if error.FieldName != "" {
			builder.WriteString(fmt.Sprintf(" %s: %s.", error.FieldName, error.Error))
		} else {
			builder.WriteString(fmt.Sprintf(" %s.", error.Error))
		}
	}

	return builder.String()
}

func GenerateOTPErrResponse(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return BadRequestErrors{
			Errors: []BadRequestError{
				{
					FieldName: "invalid_otp",
					Error:     constant.INVALID_OTP_MSG,
				},
			},
		}
	} else if errors.Is(err, model.ErrOtpExpired) {
		return BadRequestErrors{
			Errors: []BadRequestError{
				{
					FieldName: "invalid_otp",
					Error:     constant.OTP_EXPIRED_ERROR_MSG,
				},
			},
		}
	}
	return InternalServerError(fmt.Sprintf("Error in GenerateOTPErrResponse: %s", err.Error()), errtrace.Wrap(err))
}

func InternalServerError(logMessage string, err error) ErrorResponse {
	return ErrorResponse{
		ErrorCode:       constant.INTERNAL_SERVER_ERROR,
		Message:         constant.INTERNAL_SERVER_ERROR_MSG,
		StatusCode:      http.StatusInternalServerError,
		LogMessage:      logMessage,
		MaybeInnerError: err,
	}
}

func NotFoundError(logMessage string, err error) ErrorResponse {
	return ErrorResponse{
		ErrorCode:       constant.NO_DATA_FOUND,
		Message:         constant.NO_DATA_FOUND_MSG,
		StatusCode:      http.StatusNotFound,
		LogMessage:      logMessage,
		MaybeInnerError: err,
	}
}

func UnauthorizedError(logMessage string) ErrorResponse {
	return ErrorResponse{
		ErrorCode:  constant.UNAUTHORIZED_ACCESS_ERROR,
		Message:    constant.UNAUTHORIZED_ACCESS_ERROR_MSG,
		StatusCode: http.StatusUnauthorized,
		LogMessage: logMessage,
	}
}

func ForbiddenError(logMessage string, err error) ErrorResponse {
	return ErrorResponse{
		ErrorCode:       constant.FORBIDDEN,
		Message:         constant.FORBIDDEN_MSG,
		StatusCode:      http.StatusForbidden,
		LogMessage:      logMessage,
		MaybeInnerError: err,
	}
}

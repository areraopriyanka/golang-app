// NOTE: This handler is only intended for testing VisaDPS cards and will never be used in production

package handler

import (
	"fmt"
	"net/http"
	"process-api/pkg/config"
	"process-api/pkg/logging"
	"process-api/pkg/model/response"
	"process-api/pkg/security"
	"process-api/pkg/utils"
	"process-api/pkg/visadps"

	"braces.dev/errtrace"
	"github.com/labstack/echo/v4"
)

type GetCardCvvFromVisaRequest struct {
	ExpirationMM              string `json:"expirationMM" validate:"required"`
	ExpirationYY              string `json:"expirationYY" validate:"required"`
	Last4PrimaryAccountNumber string `json:"last4PrimaryAccountNumber" validate:"required"`
	ExternalId                string `json:"externalId" validate:"required"`
}

type GetCardCvvFromVisaResponse struct {
	Cvv2 string `json:"cvv2"`
}

// @Summary GetCardCvvFromVisa
// @Description Returns a cvv for a card. This is only intended for testing and will not be available in production environments
// @Tags cards
// @Produce json
// @Param payload body GetCardCvvFromVisaRequest true "GetCardDetailsRequest payload"
// @Param Authorization header string true "Bearer token for user authentication"
// @Success 200 {object} GetCardCvvFromVisaResponse
// @header 200 {string} Authorization "Bearer token for user authentication"
// @Failure 400 {object} response.BadRequestErrors
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /account/cards/cvv [post]
func GetCardCvvFromVisa(c echo.Context) error {
	_, ok := c.(*security.LoggedInRegisteredUserContext)
	if !ok {
		return response.UnauthorizedError("Failed to get user Id from custom context")
	}

	logger := logging.GetEchoContextLogger(c)

	var request GetCardCvvFromVisaRequest

	if err := c.Bind(&request); err != nil {
		logger.Error("Invalid request", "error", err.Error())
		return response.BadRequestErrors{
			Errors: []response.BadRequestError{
				{Error: err.Error()},
			},
		}
	}

	if err := c.Validate(request); err != nil {
		return err
	}

	cvv, err := generateCvv2(request.ExternalId, request.Last4PrimaryAccountNumber, request.ExpirationMM, request.ExpirationYY)
	if err != nil {
		message := fmt.Sprintf("failed to generateCvv2: %s", err.Error())
		return response.ErrorResponse{
			StatusCode:      http.StatusInternalServerError,
			LogMessage:      message,
			ErrorCode:       "INTERNAL_SERVER_ERROR",
			Message:         message,
			MaybeInnerError: errtrace.Wrap(err),
		}
	}

	return c.JSON(http.StatusOK, GetCardCvvFromVisaResponse{
		Cvv2: cvv,
	})
}

func generateCvv2(cardExternalId, last4PrimaryAccountNumber, expirationMM, expirationYY string) (string, error) {
	visaDpsSecret, err := utils.UnmarshalSecret[visadps.VisaDpsSecret](config.Config.AwsSecretManager.Region, "visadpssecret")
	if err != nil {
		return "", errtrace.Wrap(fmt.Errorf("failed to get visadpssecret: %w", err))
	}
	client, err := visadps.CreateClient(*visaDpsSecret)
	if err != nil {
		return "", errtrace.Wrap(fmt.Errorf("failed to create visadps client: %v", err))
	}
	return client.GenerateCvv2(cardExternalId, last4PrimaryAccountNumber, expirationMM, expirationYY)
}

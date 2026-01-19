package handler

import (
	"fmt"
	"net/http"
	"path"
	"process-api/pkg/constant"
	"process-api/pkg/model/response"
	"process-api/pkg/resource/agreements"
	"slices"

	"braces.dev/errtrace"
	"github.com/labstack/echo/v4"
)

// @Summary GetAgreement
// @Description Gets the latest agreement by its name
// @Produce json
// @Param agreementName path string true "Agreement Name" Enums(e-sign, dreamfi-ach-authorization, card-and-deposit, privacy-notice, terms-of-service, debtwise)
// @Success 200 {object} AgreementJSONResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /agreements/{agreementName} [get]
func GetAgreement(c echo.Context) error {
	agreementName := c.Param("agreementName")

	agreement, err := agreements.GetAgreementStore(agreementName)
	if err != nil {
		return response.ErrorResponse{
			ErrorCode:       constant.NO_DATA_FOUND,
			StatusCode:      http.StatusNotFound,
			LogMessage:      fmt.Sprintf("Could not find agreement: %s", err.Error()),
			MaybeInnerError: errtrace.Wrap(err),
		}
	}

	return c.JSON(200, AgreementJSONResponse{
		Name: agreement.Name,
		Data: agreement.JSON,
		Hash: agreement.Hash,
	})
}

type AgreementJSONResponse struct {
	Name string `json:"name" validate:"required"`
	Data string `json:"data" mask:"true" validate:"required"`
	Hash string `json:"hash" validate:"required"`
}

// @Summary GetAgreementPDF
// @Description Gets a PDF of an agreement by its name
// @Produce application/pdf
// @Param agreementName path string true "Agreement Name" Enums(e-sign, dreamfi-ach-authorization, card-and-deposit, privacy-notice, terms-of-service)
// @Success 200 {file} file
// @Failure 404 {object} response.ErrorResponse
// @Router /agreements/pdf/{agreementName} [get]
func GetAgreementPDF(c echo.Context) error {
	agreementName := c.Param("agreementName")

	validNames := []string{
		"e-sign",
		"dreamfi-ach-authorization",
		"card-and-deposit",
		"privacy-notice",
		"terms-of-service",
		"debtwise",
	}

	if !slices.Contains(validNames, agreementName) {
		return response.ErrorResponse{
			ErrorCode:       constant.NO_DATA_FOUND,
			StatusCode:      http.StatusNotFound,
			LogMessage:      "Invalid agreement name",
			MaybeInnerError: errtrace.New(""),
		}
	}

	fileName := fmt.Sprintf("%s.pdf", agreementName)
	filePath := path.Join("./pkg/resource/agreements/", fileName)

	return c.Attachment(filePath, fileName)
}

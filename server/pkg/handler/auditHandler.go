package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"process-api/pkg/audit"
	"process-api/pkg/logging"
	"process-api/pkg/model/response"

	"braces.dev/errtrace"
	"github.com/labstack/echo/v4"
)

func HandleAudit(c echo.Context) error {
	ctx := context.Background()
	client, err := audit.New(ctx)
	if err != nil {
		logging.Logger.With("error", err).Error("Audit client error")
	}

	request := new(audit.AuditRequest)

	if err := c.Bind(request); err != nil {
		logging.Logger.Error("An error occurred while binding unsigned mobile request from echo context to object", "error", err.Error())
		return response.InternalServerError(fmt.Sprintf("An error occurred while binding unsigned mobile request from echo context to object: %s", err.Error()), errtrace.Wrap(err))
	}

	ledgerResp, err := client.GetAudits(ctx, *request)
	if err != nil {
		logging.Logger.Error("An error occurred during calling Audit API", "error", err.Error())
		return errtrace.Wrap(errors.New("An error occurred during calling Audit API: " + err.Error()))
	}

	ledgerRespBodyJson := json.RawMessage(ledgerResp)

	return c.JSON(http.StatusOK, ledgerRespBodyJson)
}

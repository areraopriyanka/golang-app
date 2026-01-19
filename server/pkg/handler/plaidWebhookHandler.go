package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"process-api/pkg/db"
	"process-api/pkg/logging"
	"process-api/pkg/plaid"

	"github.com/labstack/echo/v4"
)

func (h *Handler) PlaidWebhookHandler(c echo.Context) error {
	logger := logging.GetEchoContextLogger(c).WithGroup("PlaidWebhookHandler")

	// We can't use `c.Bind` here like we do for other endponts, since we need the
	// raw string for Plaid webhook validation.
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		logger.Error("Failed to read request body", "error", err.Error())
		return c.NoContent(http.StatusBadRequest)
	}
	webhookBody := string(body)

	var payload plaid.WebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		logger.Error("Invalid webhook payload", "error", err.Error())
		return c.NoContent(http.StatusBadRequest)
	}

	ps := plaid.PlaidService{Logger: logger, Plaid: h.Plaid, DB: db.DB, WebhookURL: h.Config.Plaid.WebhookURL}

	if !ps.VerifyWebhook(webhookBody, c.Request().Header.Get("Plaid-Verification")) {
		logger.Error("Webhook signature verification failed")
		return c.NoContent(http.StatusUnauthorized)
	}

	webhookType := payload.WebhookType
	webhookCode := payload.WebhookCode
	webhookError := payload.Error
	itemID := payload.ItemID
	logger.Debug("Received Plaid webhook", "webhookType", webhookType, "webhookCode", webhookCode, "itemID", itemID)

	// NOTE: webhooks are not sent for items linked through the "Instant Micro-deposits" flow (DT-1480)
	switch payload.WebhookType {
	case "ITEM":
		if err := ps.HandleItemWebhook(webhookCode, itemID, webhookError); err != nil {
			logger.Error("Failed handle item webhook", "error", err.Error(), "webhookCode", webhookCode, "itemID", itemID)
		}
	case "AUTH":
		if payload.AccountID != nil {
			if err := ps.HandleAuthWebhook(webhookCode, *payload.AccountID); err != nil {
				logger.Error("Failed handle auth webhook", "error", err.Error(), "webhookCode", webhookCode, "itemID", itemID)
			}
		} else {
			logger.Error("Failed handle auth webhook", "error", "payload missing account_id", "webhookCode", webhookCode, "itemID", itemID)
		}
	}

	return c.NoContent(http.StatusOK)
}

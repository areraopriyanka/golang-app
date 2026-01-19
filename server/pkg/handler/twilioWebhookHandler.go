package handler

import (
	"net/http"
	"process-api/pkg/db"
	"process-api/pkg/db/dao"
	"process-api/pkg/logging"

	"github.com/labstack/echo/v4"
)

func TwilioWebhookHandler(c echo.Context) error {
	logger := logging.GetEchoContextLogger(c)

	logger.Info("Inside TwilioWebhookHandler")

	status := c.FormValue("CallStatus")
	to := c.FormValue("To")
	callSID := c.FormValue("CallSid")

	logger.Info("Call status:", "status", status, "to mobileNumber", to)
	callStatus := dao.MasterUserCallRecordDao{
		CallStatus: status,
		To:         to,
		CallSid:    callSID,
	}

	result := db.DB.Where("call_sid=?", callSID).Assign(callStatus).FirstOrCreate(&callStatus)
	if result.Error != nil {
		logger.Error("DB Error:", "error", result.Error)
		return result.Error
	}

	return c.NoContent(http.StatusOK)
}

package handler

import (
	"net/http"
	"process-api/pkg/logging"
	"process-api/pkg/version"

	"github.com/labstack/echo/v4"
)

func GetApplicationVersion(c echo.Context) error {
	versionResponse := map[string]string{
		"appVersion": version.ApplicationVersion,
	}
	logging.Logger.Info("Response of GetApplicationVersion", "versionResponse", versionResponse)
	return c.JSON(http.StatusOK, versionResponse)
}

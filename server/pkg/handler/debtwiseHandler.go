package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"process-api/pkg/config"
	"process-api/pkg/debtwise"
	"process-api/pkg/logging"
	"process-api/pkg/model/response"

	"braces.dev/errtrace"
	"github.com/labstack/echo/v4"
)

func DebtwiseHandler(c echo.Context) error {
	logger := logging.GetEchoContextLogger(c)

	debtwiseClient, err := debtwise.NewDebtwiseClient(config.Config.Debtwise, logger)
	if err != nil {
		logging.Logger.Error("Error occurred while creating debtwise client: " + err.Error())
		return response.InternalServerError(fmt.Sprintf("Error occurred while creating debtwise client: %s", err.Error()), errtrace.Wrap(err))
	}

	var requestBody debtwise.CreateUserJSONRequestBody
	err = json.NewDecoder(c.Request().Body).Decode(&requestBody)
	if err != nil {
		logging.Logger.Error("Error occurred while decoding request body: " + err.Error())
		return response.InternalServerError(fmt.Sprintf("Error occurred while decoding request body: %s", err.Error()), errtrace.Wrap(err))
	}

	jsonBytes, err := json.Marshal(requestBody)
	if err != nil {
		logging.Logger.Error("Error occurred while marshalling request body: " + err.Error())
		return response.InternalServerError(fmt.Sprintf("Error occurred while marshalling request body: %s", err.Error()), errtrace.Wrap(err))
	}

	debtwiseResponse, err := debtwiseClient.CreateUserWithBodyWithResponse(
		context.Background(),
		nil,
		"application/json",
		bytes.NewReader(jsonBytes),
	)
	if err != nil {
		logging.Logger.Error("Error occurred while creating user: " + err.Error())
		return response.InternalServerError(fmt.Sprintf("Error occurred while creating user: %s", err.Error()), errtrace.Wrap(err))
	}
	userId := debtwiseResponse.JSON201.Id

	resp, err := debtwiseClient.RetrieveCreditScore(context.Background(), userId, nil)
	if err != nil {
		logging.Logger.Error("Error occurred while fetching credit score: " + err.Error())
		return response.InternalServerError(fmt.Sprintf("Error occurred while fetching credit score: %s", err.Error()), errtrace.Wrap(err))
	}

	defer resp.Body.Close()

	var creditScore debtwise.CreditScore
	if err := json.NewDecoder(resp.Body).Decode(&creditScore); err != nil {
		logger.Error("Failed to decode CreditScore response", "error", err.Error())
		return response.InternalServerError(fmt.Sprintf("Failed to decode CreditScore response: error: %s", err.Error()), errtrace.Wrap(err))
	}

	return c.JSON(http.StatusOK, creditScore)
}

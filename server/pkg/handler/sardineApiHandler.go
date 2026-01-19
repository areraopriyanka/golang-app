package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"process-api/pkg/config"
	"process-api/pkg/constant"
	"process-api/pkg/logging"
	"process-api/pkg/model/response"
	"process-api/pkg/sardine"
	"process-api/pkg/utils"

	"braces.dev/errtrace"
	"github.com/labstack/echo/v4"
)

func ManageSardineAPICall(c echo.Context) error {
	client, err := utils.NewSardineClient(config.Config.Sardine)
	if err != nil {
		logging.Logger.Error("Error occurred while creating sardine client: " + err.Error())
		return response.ErrorResponse{ErrorCode: constant.INTERNAL_SERVER_ERROR, StatusCode: http.StatusInternalServerError, LogMessage: fmt.Sprintf("Error occurred while creating sardine client: %s", err.Error()), MaybeInnerError: errtrace.Wrap(err)}
	}
	var requestBody sardine.PostCustomerInformationJSONRequestBody
	err = json.NewDecoder(c.Request().Body).Decode(&requestBody)
	if err != nil {
		logging.Logger.Error("Error occurred while decoding request body: " + err.Error())
		return response.ErrorResponse{ErrorCode: constant.INTERNAL_SERVER_ERROR, StatusCode: http.StatusInternalServerError, LogMessage: fmt.Sprintf("Error occurred while decoding request body: %s", err.Error()), MaybeInnerError: errtrace.Wrap(err)}
	}

	sardineResponse, err := client.PostCustomerInformationWithResponse(context.Background(), requestBody, func(ctx context.Context, req *http.Request) error {
		req.Header.Set("Content-Type", "application/json")
		return nil
	})
	if err != nil {
		logging.Logger.Error("Error occurred while calling sardine API: " + err.Error())
		return response.ErrorResponse{ErrorCode: constant.INTERNAL_SERVER_ERROR, StatusCode: http.StatusInternalServerError, LogMessage: fmt.Sprintf("Error occurred while calling sardine API: %s", err.Error()), MaybeInnerError: errtrace.Wrap(err)}
	}

	var res bytes.Buffer
	if err := json.Indent(&res, sardineResponse.Body, "", "  "); err != nil {
		res = *bytes.NewBuffer(sardineResponse.Body)
	}

	logging.Logger.Info("Received response from Sardine", "statusCode", sardineResponse.StatusCode(), "response", res.String())

	switch {
	case sardineResponse.JSON200 != nil:
		return c.JSON(200, sardineResponse.JSON200)
	case sardineResponse.JSON400 != nil:
		return c.JSON(400, sardineResponse.JSON400)
	case sardineResponse.JSON401 != nil:
		return c.JSON(401, sardineResponse.JSON401)
	case sardineResponse.JSON422 != nil:
		return c.JSON(422, sardineResponse.JSON422)
	default:
		logging.Logger.Error("Received error response from sardine API")
		return response.ErrorResponse{ErrorCode: constant.INTERNAL_SERVER_ERROR, StatusCode: sardineResponse.StatusCode(), LogMessage: "Received error response from sardine API", MaybeInnerError: errtrace.New("")}
	}
}

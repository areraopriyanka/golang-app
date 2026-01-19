package ledger

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"process-api/pkg/config"
	"process-api/pkg/logging"
	"process-api/pkg/model/response"

	"braces.dev/errtrace"
	"github.com/labstack/echo/v4"
)

func CallLedgerAPIAndGetRawResponse(c echo.Context, request *Request) (int, json.RawMessage, error) {
	return CallLedgerAPIWithUrlAndGetRawResponse(c, request, "")
}

func CallLedgerAPIWithUrlAndGetRawResponse(c echo.Context, request *Request, apiUrl string) (int, json.RawMessage, error) {
	ledgerUrl := config.Config.Ledger.Endpoint
	if apiUrl != "" {
		ledgerUrl = apiUrl
	}

	logger := logging.Logger.With("ledger", ledgerUrl)

	logger.Info("Calling Ledger API from Middleware", "requestMethod", request.Method)

	reqByteArr, err := json.Marshal(&request)
	if err != nil {
		logger.Error("An error occurred during json marshal mobile request to ledger request byte array", "error", err.Error())
		return 0, nil, errtrace.Wrap(errors.New("An error occurred during json marshal mobile request to ledger request byte array: " + err.Error()))
	}

	var reqIndented bytes.Buffer
	if err := json.Indent(&reqIndented, reqByteArr, "", "  "); err != nil {
		reqIndented = *bytes.NewBuffer(reqByteArr)
	}
	logger.Info("ledger caller request", "request", reqIndented.String())

	logger.Info("Final API URL", "ledgerUrl", ledgerUrl)
	req, err := http.NewRequest("POST", ledgerUrl, bytes.NewBuffer((reqByteArr)))
	if err != nil {
		logger.Error("Error creating request", "error", err)
		return 0, nil, errtrace.Wrap(errors.New("An error occurred during calling ledger API using http post: " + err.Error()))
	}
	if c != nil {
		for key, values := range c.Request().Header {
			for _, value := range values {
				req.Header.Set(key, value)
			}
		}
	}
	client := &http.Client{}
	ledgerResp, err := client.Do(req)
	if err != nil {
		logger.Error("Error while making request", "error", err)
		return 0, nil, errtrace.Wrap(errors.New("An error occurred during calling ledger API using http post: " + err.Error()))
	}

	defer ledgerResp.Body.Close()
	// Read the response body
	ledgerRespBodyByteArr, err := io.ReadAll(ledgerResp.Body)
	if err != nil {
		logger.Error("An error occurred during converting ledgerResp body to byte array", "error", err)
		return 0, nil, errtrace.Wrap(errors.New("An error occurred during converting ledgerResp body to byte array: " + err.Error()))
	}

	ledgerRespBodyJson := json.RawMessage(ledgerRespBodyByteArr)
	logger.Info("Received response from Ledger", "method", request.Method, "statusCode", ledgerResp.StatusCode)

	var res bytes.Buffer
	if err := json.Indent(&res, ledgerRespBodyByteArr, "", "  "); err != nil {
		res = *bytes.NewBuffer(ledgerRespBodyByteArr)
	}
	logger.Info("ledger caller response", "response", res.String())

	return ledgerResp.StatusCode, ledgerRespBodyJson, nil
}

func CallLedgerAPI(c echo.Context, request *Request) error {
	return CallLedgerAPIWithUrl(c, request, "")
}

func CallLedgerAPIWithUrl(c echo.Context, request *Request, apiUrl string) error {
	statusCode, respBody, err := CallLedgerAPIWithUrlAndGetRawResponse(c, request, apiUrl)
	if err != nil {
		return response.InternalServerError(err.Error(), errtrace.Wrap(err))
	}
	return c.JSON(statusCode, respBody)
}

func CallLedgerAPIWithUrlAndGetTypedResponse(request *Request, url string, response interface{}, logger *slog.Logger,
) error {
	statusCode, respBody, ledgerErr := CallLedgerAPIWithUrlAndGetRawResponse(nil, request, url)
	if ledgerErr != nil {
		logger.Error("Ledger returned an error response", "error", ledgerErr.Error())
		return errtrace.Wrap(ledgerErr)
	}

	if statusCode != http.StatusOK {
		logger.Error("Ledger returned non-success status code", "statusCode", statusCode, "respBody", string(respBody))
		return errtrace.Wrap(errors.New("ledger returned non-success statusCode"))
	}

	unmarshalErr := json.Unmarshal(respBody, response)
	if unmarshalErr != nil {
		logger.Error("An error occurred while unmarshaling ledger response", "error", unmarshalErr.Error())
		return errtrace.Wrap(unmarshalErr)
	}

	return nil
}

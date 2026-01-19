package debtwise

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"process-api/pkg/config"

	"braces.dev/errtrace"
)

type LoggingClient struct {
	client HttpRequestDoer
	logger *slog.Logger
}

func (lc *LoggingClient) Do(req *http.Request) (*http.Response, error) {
	if lc.logger.Handler().Enabled(context.Background(), slog.LevelDebug.Level()) {
		reqBody := []byte{}
		if req.Body != nil {
			reqBody, _ = io.ReadAll(req.Body)
		}
		req.Body = io.NopCloser(bytes.NewBuffer(reqBody))

		lc.logger.Debug("Debtwise Request", "url", req.URL.String(), "request", string(reqBody))
	}

	resp, err := lc.client.Do(req)

	if lc.logger.Handler().Enabled(context.Background(), slog.LevelDebug.Level()) {
		respBody := []byte{}
		if resp.Body != nil {
			respBody, _ = io.ReadAll(resp.Body)
		}
		resp.Body = io.NopCloser(bytes.NewBuffer(respBody))

		lc.logger.Debug("Debtwise Response", "response", string(respBody))
	}

	return resp, err
}

func NewDebtwiseClient(debtwiseConfig config.DebtwiseConfigs, logger *slog.Logger) (*ClientWithResponses, error) {
	debtwiseClient, err := NewClientWithResponses(debtwiseConfig.ApiBase, func(c *Client) error {
		c.RequestEditors = append(c.RequestEditors, func(ctx context.Context, req *http.Request) error {
			req.Header.Set("Authorization", debtwiseConfig.Credential)
			req.Header.Set("Content-Type", "application/json")
			return nil
		})
		loggingClient := LoggingClient{
			client: &http.Client{},
			logger: logger,
		}
		c.Client = &loggingClient
		return nil
	})
	if err != nil {
		return nil, errtrace.Wrap(err)
	}
	return debtwiseClient, nil
}

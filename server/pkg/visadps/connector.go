package visadps

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"process-api/pkg/config"
	"process-api/pkg/logging"
)

type ConnectorClient struct {
	baseUrl string
	client  *http.Client
}

func NewConnectorClient(cfg config.VisaSimulatorConfigs) *ConnectorClient {
	return &ConnectorClient{
		baseUrl: cfg.BaseUrl,
		client:  &http.Client{},
	}
}

func (c *ConnectorClient) SendTransaction(endpoint string, payload interface{}) (map[string]interface{}, error) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	url := c.baseUrl + endpoint
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	logging.Logger.Info("Sending transaction", "url", url, "payload", string(jsonData))

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	logging.Logger.Info("Received response", "status", resp.StatusCode, "body", string(body))

	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return result, nil
}

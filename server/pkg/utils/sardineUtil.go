package utils

import (
	"context"
	"net/http"
	"process-api/pkg/config"
	"process-api/pkg/sardine"

	"braces.dev/errtrace"
)

func NewSardineClient(sardineConfig config.SardineConfigs) (*sardine.ClientWithResponses, error) {
	sardineClient, err := sardine.NewClientWithResponses(sardineConfig.ApiBase, func(c *sardine.Client) error {
		c.RequestEditors = append(c.RequestEditors, func(ctx context.Context, req *http.Request) error {
			req.Header.Set("Authorization", sardineConfig.Credential)
			req.Header.Set("Content-Type", "application/json")
			return nil
		})
		return nil
	})
	// If there is an error while creating the client
	if err != nil {
		return nil, errtrace.Wrap(err)
	}
	return sardineClient, nil
}

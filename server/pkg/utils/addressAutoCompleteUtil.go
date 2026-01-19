package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"process-api/pkg/config"
	"process-api/pkg/logging"
	"process-api/pkg/model/response"
	"regexp"

	"braces.dev/errtrace"
)

type SmartyStreetClient struct {
	AuthID    string
	AuthToken string
	BaseURL   string
}

func NewAutocompleteClient(id, token string) *SmartyStreetClient {
	return &SmartyStreetClient{
		AuthID:    id,
		AuthToken: token,
		BaseURL:   config.Config.SmartyStreets.LookupUrl,
	}
}

func (c *SmartyStreetClient) Lookup(search string, max int) ([]response.Suggestion, error) {
	params := url.Values{}
	params.Add("auth-id", c.AuthID)
	params.Add("auth-token", c.AuthToken)
	params.Add("search", search)
	// Request up to 10 address suggestions (allowed range: 1–10).
	// Docs: https://www.smarty.com/docs/cloud/us-autocomplete-pro-api#pro-http-request-headers
	params.Add("max_results", fmt.Sprintf("%d", max))

	suggestions, err := c.LookupAddress(params)
	return suggestions, errtrace.Wrap(err)
}

func (c *SmartyStreetClient) FetchNestedSuggestions(suggestion response.Suggestion) ([]response.Suggestion, error) {
	params := url.Values{}
	params.Add("auth-id", c.AuthID)
	params.Add("auth-token", c.AuthToken)
	params.Add("search", suggestion.Street)
	// The `selected` query parameter must contain the address in a specific format,
	// e.g., 123 1/2 S Lafayette St Apt (3) Greenville, MI 48838
	// If entries > 1, it indicates multiple apartment numbers for the same address.
	// If an address has more than 100 entries, only the first 100 will be returned in case smarty-address-autocomplete-api with selected param.
	// Doc: https://www.smarty.com/docs/cloud/us-autocomplete-pro-api
	params.Add("selected", buildAddress(suggestion))
	nestedSuggestions, err := c.LookupAddress(params)
	return nestedSuggestions, errtrace.Wrap(err)
}

func (c *SmartyStreetClient) LookupAddress(params url.Values) ([]response.Suggestion, error) {
	resp, err := http.Get(c.BaseURL + "?" + params.Encode())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errtrace.Wrap(fmt.Errorf("error reading response body: %w", err))
	}

	logging.Logger.Debug("Response from smarty street API", "Response", string(body))

	if resp.StatusCode != http.StatusOK {
		return nil, errtrace.Wrap(fmt.Errorf("received non-OK HTTP status: %s", resp.Status))
	}

	if len(body) == 0 {
		return nil, errtrace.Wrap(fmt.Errorf("empty response from API"))
	}

	var apiResp response.Address

	if err := json.Unmarshal(body, &apiResp); err != nil {
		logging.Logger.Error("Error while unmarshaling the response", "error", err.Error())
		return nil, errtrace.Wrap(fmt.Errorf("error while unmarshaling the response: %w", err))
	}

	filteredSuggestions := make([]response.Suggestion, 0, len(apiResp.Suggestions))

	for _, suggestion := range apiResp.Suggestions {
		if !IsPOBox(suggestion.Street) {
			filteredSuggestions = append(filteredSuggestions, suggestion)
		}
	}

	return filteredSuggestions, nil
}

func IsPOBox(street string) bool {
	poBoxPattern := regexp.MustCompile(`(?i)\b(p[\.\s]*o[\.\s]*box|box\s+\d+)\b`)
	return poBoxPattern.MatchString(street)
}

func buildAddress(suggestion response.Suggestion) string {
	address := suggestion.Street + " " + suggestion.Secondary + fmt.Sprintf(" (%d)", suggestion.Entries) + " " + suggestion.City + ", " + suggestion.State + " " + suggestion.ZipCode
	return address
}

func ValidateSearchFieldLength(search string) bool {
	length := len([]byte(search))
	if length >= 1 && length <= 127 {
		return true
	}
	return false
}

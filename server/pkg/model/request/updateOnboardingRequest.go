package request

import "encoding/json"

type UpdateOnboardingDetailsRequest struct {
	LastScreenSubmitted string          `json:"lastScreenSubmitted"`
	UserDetails         json.RawMessage `json:"userDetails"`
	Extras              json.RawMessage `json:"extras"`
}

package ledger

import "encoding/json"

type WebhookEventResponse struct {
	EventId   string          `json:"eventId"`
	EventName string          `json:"eventName"`
	Payload   json.RawMessage `json:"payload"`
}

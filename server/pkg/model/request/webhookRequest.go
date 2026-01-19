package request

import "encoding/json"

type LedgerEventPayload struct {
	Source    string          `json:"source"`
	EventId   string          `json:"eventId"`
	EventName string          `json:"eventName"`
	Payload   json.RawMessage `json:"payload" mask:"true"`
	Signature string          `json:"signature" mask:"true"`
	Timestamp *string         `json:"timestamp" mask:"true"`
}

type GetDeviceDetailsRequest struct {
	CustomerNo     string   `json:"customerNo" mask:"true"`
	AccountNumbers []string `json:"accountNo" mask:"true"`
	DeviceToken    string   `json:"deviceToken" mask:"true"`
	DeviceId       string   `json:"deviceId" mask:"true"`
}

type ClearNotificationRequest struct {
	NotificationId string `query:"notificationId"`
	CustomerNo     string `query:"customerNo" mask:"true"`
	DeleteAll      bool   `query:"deleteAll"`
}

type MarkNotificationAsReadRequest struct {
	NotificationId string `query:"notificationId"`
	CustomerNo     string `query:"customerNo" mask:"true"`
}

// NotificationPayload represents the payload for an FCM notification
type NotificationPayload struct {
	Title          string          `json:"title"`
	Data           string          `json:"body"`
	NotificationId string          `json:"notificationId"`
	Category       string          `json:"category"`
	Payload        json.RawMessage `json:"payload" mask:"true"`
}

package response

import "encoding/json"

type CreateSubscriptionResponse struct {
	SubscriptionId string `json:"subscriptionId" mask:"true"`
}

type NotificationResponse struct {
	NotificationId string          `json:"notificationId"`
	Payload        json.RawMessage `json:"payload"`
	IsRead         bool            `json:"isRead"`
}

type CategoryCount struct {
	Category string `json:"category"`
	Count    int    `json:"count"`
}

type ResponseData struct {
	TotalCount      int            `json:"totalCount"`
	CountByCategory map[string]int `json:"countByCategory"`
}

type NotificationPayloadResponse struct {
	Payload json.RawMessage `json:"payload" mask:"true"`
}

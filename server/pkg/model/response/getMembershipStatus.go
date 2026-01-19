package response

type GetMemberShipStatus struct {
	Status string `json:"status" validate:"required" enums:"subscribed,unsubscribed"`
}

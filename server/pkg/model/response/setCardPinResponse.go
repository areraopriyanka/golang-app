package response

type SetCardPinResponse struct {
	PinSet bool `json:"pinSet" validate:"required"`
}

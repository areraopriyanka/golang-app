package request

type ReplaceCardReason struct {
	Reason string `json:"reason" validate:"required,oneof=lost stolen damaged"`
}

type ReplaceCardRequest struct {
	LedgerApiRequest
}

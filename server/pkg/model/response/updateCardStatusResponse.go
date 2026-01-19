package response

type UpdateCardStatusResponse struct {
	// Maps known ledger card statuses to known values. Returns empty status for unknown values.
	UpdatedCardStatus string `json:"updatedCardStatus" validate:"required" enums:"active,inactive,frozen,cancelled"`
}

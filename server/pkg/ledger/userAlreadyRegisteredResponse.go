package ledger

type UserAlreadyRegisteredResponse struct {
	Result UserAlreadyRegisteredResult
}
type UserAlreadyRegisteredResult struct {
	ID string `json:"ID"`
}

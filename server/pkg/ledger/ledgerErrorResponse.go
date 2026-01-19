package ledger

type LedgerErrorResponse struct {
	Id    string          `json:"id"`
	Error MaybeInnerError `json:"error"`
}

type MaybeInnerError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

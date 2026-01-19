package ledger

type AddCardHolderLedgerResponse struct {
	Id     string               `json:"id"`
	Result *AddCardHolderResult `json:"result"`
	Error  *MaybeInnerError     `json:"error"`
}

type AddCardHolderResult struct {
	CardHolderId string `json:"cardHolderId"`
}

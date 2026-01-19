package ledger

type AddCardHolderPayload struct {
	Reference       string `json:"reference" validate:"required"`
	TransactionType string `json:"transactionType" validate:"required"`
	CustomerId      string `json:"customerId" validate:"required"`
	AccountNumber   string `json:"accountNumber" validate:"required"`
	Product         string `json:"product" validate:"required"`
	Channel         string `json:"channel" validate:"required"`
	Program         string `json:"program" validate:"required"`
}

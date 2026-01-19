package request

type BuildValidateCvvRequest struct {
	CVV string `json:"cvv" validate:"required" mask:"true"`
}

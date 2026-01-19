package response

type GetCardResponse struct {
	Card CardData `json:"card" validate:"required"`
}

type CardData struct {
	CardId         string  `json:"cardId" validate:"required" mask:"true"`
	PreviousCardId *string `json:"previousCardId,omitempty" mask:"true"`
	// static cast of ledger card status for known statuses. Undefined if unknown status
	CardStatus string `json:"cardStatus" enums:"active,inactive,frozen,cancelled"`
	// raw ledger card status. For use when cardStatus is undefined.
	CardStatusRaw          string  `json:"cardStatusRaw" validate:"required"`
	OrderStatus            string  `json:"orderStatus" validate:"required"`
	IsReIssue              bool    `json:"isReIssue" validate:"required"`
	IsReplace              bool    `json:"isReplace" validate:"required"`
	IsReplaceLocked        bool    `json:"isReplaceLocked" validate:"required"`
	CardMaskNumber         string  `json:"cardMaskNumber" validate:"required"`
	PreviousCardMaskNumber *string `json:"previousCardMaskNumber,omitempty" mask:"true"`
	CardExpiryDate         string  `json:"cardExpiryDate" validate:"required" mask:"true"`
	IsPreviousCardFrozen   bool    `json:"isPreviousCardFrozen" validate:"required" mask:"true"`
	ExternalCardId         string  `json:"externalCardId" validate:"required" mask:"true"`
}

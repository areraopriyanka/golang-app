package response

type UpdateAccountStatusResponse struct {
	UpdatedAccountStatus string `json:"updatedAccountStatus" validate:"required" enums:"active,temporary_inactive,dormant,closed,disabled"`
}

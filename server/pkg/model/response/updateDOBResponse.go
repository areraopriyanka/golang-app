package response

type UpdateDOBResponse struct {
	UserStatus string `json:"userStatus" validate:"required"`
}

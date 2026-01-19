package request

type UpdateCustomerRequest struct {
	FirstName string `json:"firstName"  gorm:"column:first_name"`
	LastName  string `json:"lastName"  gorm:"column:last_name"`
	Suffix    string `json:"suffix"`
	Email     string `json:"email" validate:"omitempty,validateEmail" mask:"true"`
}

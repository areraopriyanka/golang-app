package response

type GetUserDetailsAndUpdateStatus struct {
	FirstName      string         `json:"firstName" validate:"required"`
	LastName       string         `json:"lastName" validate:"required"`
	Suffix         string         `json:"suffix,omitempty"`
	Email          string         `json:"email" validate:"required,email"`
	MobileNumber   string         `json:"mobileNumber" validate:"required"`
	Address        AddressDetails `json:"address" validate:"required"`
	FullNameStatus string         `json:"fullNameStatus,omitempty" enums:"pending,accepted,rejected"`
	EmailStatus    string         `json:"emailStatus,omitempty" enums:"pending,accepted,rejected"`
	AddressStatus  string         `json:"addressStatus,omitempty" enums:"pending,accepted,rejected"`
}

type AddressDetails struct {
	AddressLine1 string `json:"addressLine1" validate:"required"`
	AddressLine2 string `json:"addressLine2"`
	City         string `json:"city" validate:"required"`
	State        string `json:"state" validate:"required"`
	PostalCode   string `json:"postalCode" validate:"required"`
}

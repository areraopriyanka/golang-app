package request

type AddPersonalDetailsRequest struct {
	FirstName string `json:"firstName" validate:"required,validateName"`
	LastName  string `json:"lastName" validate:"required,validateName"`
	Suffix    string `json:"suffix"`
	DOB       string `json:"dob" validate:"required,validateDOB" mask:"true"`
}

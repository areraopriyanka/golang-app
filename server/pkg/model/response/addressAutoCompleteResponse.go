package response

type Address struct {
	Suggestions []Suggestion `json:"suggestions"`
}

type Suggestion struct {
	Street    string `json:"street_line" validate:"required"`
	Secondary string `json:"secondary"`
	City      string `json:"city" validate:"required"`
	State     string `json:"state" validate:"required"`
	ZipCode   string `json:"zipcode" validate:"required"`
	Entries   int    `json:"entries" validate:"required"`
}

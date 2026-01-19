package response

type Statement struct {
	Id    string `json:"id" validate:"required"`
	Month string `json:"month" validate:"required"`
	Year  int32  `json:"year" validate:"required"`
}
type ListStatementResponse struct {
	Statements []Statement `json:"statements" validate:"required"`
	TotalCount int32       `json:"totalCount" validate:"required"`
}

package request

type BuildListStatementsPayloadRequest struct {
	PageNumber int `json:"pageNumber" validate:"required,gt=0"`
	PageSize   int `json:"pageSize" validate:"required,gt=0"`
}

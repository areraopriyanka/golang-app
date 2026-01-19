package request

type BuildGetStatementPayloadRequest struct {
	StatementId string `json:"statementId" validate:"required"`
}

package request

type JobStatusRequest struct {
	JobId int `json:"jobId" validate:"required"`
}

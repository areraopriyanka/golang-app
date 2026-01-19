package response

// NOTE: Enum values for State derived from rivertype docs here: https://pkg.go.dev/github.com/riverqueue/river/rivertype#JobState
type MaybeRiverJobResponse[T any] struct {
	JobId   int    `json:"jobId" validate:"required"`
	State   string `json:"state" validate:"required,oneof=available cancelled completed discarded pending retryable running scheduled"`
	Payload *T     `json:"payload,omitempty"`
}

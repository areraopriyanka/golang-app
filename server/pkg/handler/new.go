package handler

import (
	"database/sql"
	"process-api/pkg/config"

	"github.com/plaid/plaid-go/v34/plaid"
	"github.com/riverqueue/river"
)

type Handler struct {
	Config      *config.Configs
	Plaid       *plaid.APIClient
	Env         string
	RiverClient *river.Client[*sql.Tx]
}

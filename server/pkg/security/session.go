package security

import (
	"database/sql"
	"net/http"
	"process-api/pkg/config"
	"time"

	"github.com/antonlindstrom/pgstore"
	"github.com/gorilla/sessions"
)

func InitSessionStore(path string, sqlDB *sql.DB) (*pgstore.PGStore, error) {
	sessionStore, err := pgstore.NewPGStoreFromPool(sqlDB, []byte(config.Config.Admin.SessionSecretKey))
	if err != nil {
		return nil, err
	}

	sessionStore.Options = &sessions.Options{
		Path:     path,
		MaxAge:   int(15 * time.Minute / time.Second), // matches configured auth0 jwt timeout
		HttpOnly: true,
		Secure:   config.Config.Server.BaseUrl[:5] == "https",
		SameSite: http.SameSiteStrictMode,
	}

	return sessionStore, nil
}

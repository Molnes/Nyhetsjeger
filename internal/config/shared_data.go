package config

import (
	"database/sql"

	"github.com/antonlindstrom/pgstore"
)

type SharedData struct {
	DatabaseConn *sql.DB
	SessionStore *pgstore.PGStore
	CryptoKey   []byte
}

package config

import (
	"database/sql"

	"github.com/antonlindstrom/pgstore"
)

type SharedData struct {
	DB           *sql.DB
	SessionStore *pgstore.PGStore
	CryptoKey    []byte
}

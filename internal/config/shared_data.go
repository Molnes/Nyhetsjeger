package config

import (
	"database/sql"

	"github.com/antonlindstrom/pgstore"
	"github.com/minio/minio-go/v7"
)

type SharedData struct {
	DB           *sql.DB
	SessionStore *pgstore.PGStore
	CryptoKey    []byte
	Bucket       *minio.Client
	OpenAIKey    string
}

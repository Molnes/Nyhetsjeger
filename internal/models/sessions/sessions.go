package sessions

// Package sessions provides a session store for the application

import (
	"database/sql"
	"time"

	"github.com/antonlindstrom/pgstore"
)

// var (
// 	// Session store
// 	Store *pgstore.PGStore
// )

// Constants related to the session store
const (
	// SESSION_NAME is the name of the session
	SESSION_NAME = "session"
	// USER_DATA_VALUE is the key for the user data in the session
	USER_DATA_VALUE = "user"
)

// Creates and sets up a new session store
func NewSessionStore(databaseConn *sql.DB, sessionKey []byte) (*pgstore.PGStore, error) {

	pgStore, err := pgstore.NewPGStoreFromPool(databaseConn, sessionKey)
	if err != nil {
		return nil, err
	}

	pgStore.Options.Secure = true
	pgStore.Options.HttpOnly = true
	pgStore.Options.MaxAge = 60 * 60 * 24 * 30 // 30 days

	pgStore.Cleanup(time.Hour * 24)

	return pgStore, nil
}

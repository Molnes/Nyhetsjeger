package sessions

// Package sessions provides a session store for the application

import (
	"database/sql"
	"log"
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
	SESSION_NAME    = "session"
	// USER_DATA_VALUE is the key for the user data in the session
	USER_DATA_VALUE = "user"
)

// func init() {
// 	sessionKey, ok := os.LookupEnv("SESSION_SECRET")
// 	if !ok {
// 		log.Fatal("No session secret provided. Expected SESSION_SECRET")
// 	}

// 	pgStore, err := pgstore.NewPGStoreFromPool(database.DB, []byte(sessionKey))
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	pgStore.Options.Secure = true
// 	pgStore.Options.HttpOnly = true
// 	pgStore.Options.MaxAge = 60 * 60 * 24 * 30 // 30 days

// 	pgStore.Cleanup(time.Hour * 24)

// 	Store = pgStore
// }

func NewSessionStore(databaseConn *sql.DB, sessionKey []byte) *pgstore.PGStore {

	pgStore, err := pgstore.NewPGStoreFromPool(databaseConn, sessionKey)
	if err != nil {
		log.Fatal(err)
	}

	pgStore.Options.Secure = true
	pgStore.Options.HttpOnly = true
	pgStore.Options.MaxAge = 60 * 60 * 24 * 30 // 30 days

	pgStore.Cleanup(time.Hour * 24)

	return pgStore
}

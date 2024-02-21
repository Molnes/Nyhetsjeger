package sessions

// Package sessions provides a session store for the application

import (
	"log"
	"os"
	"time"

	"github.com/Molnes/Nyhetsjeger/internal/database"
	"github.com/antonlindstrom/pgstore"
)

var (
	// Session store
	Store *pgstore.PGStore
)

const (
	// SessionName is the name of the session
	SessionName = "session"
)

func init() {
	sessionKey, ok := os.LookupEnv("SESSION_SECRET")
	if !ok {
		log.Fatal("No session secret provided. Expected SESSION_SECRET")
	}

	pgStore, err := pgstore.NewPGStoreFromPool(database.DB, []byte(sessionKey))
	if err != nil {
		log.Fatal(err)
	}

	pgStore.Options.Secure = true
	pgStore.Options.HttpOnly = true
	pgStore.Options.MaxAge = 60 * 60 * 24 * 30 // 30 days

	pgStore.Cleanup(time.Hour * 24)

	Store = pgStore
}

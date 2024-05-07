package database

import (
	"database/sql"

	_ "github.com/lib/pq"
)

// NewDatabaseConnection creates a new connection to a database
func NewDatabaseConnection(dbURL string) (*sql.DB, error) {
	sql_db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, err
	}

	if err = sql_db.Ping(); err != nil {
		return nil, err
	}

	return sql_db, nil
}

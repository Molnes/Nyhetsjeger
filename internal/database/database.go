package database

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"
)

var (
	// Globally available database connection
	DB *sql.DB
)

// initializes the database connection by loading the environment variables from the .env file,
//
// Sets globally available DB to the initialized database connection.
func init() {
	db_url, ok := os.LookupEnv("POSTGRESQL_URL_APP")
	if !ok {
		log.Fatal("No database url provided. Expected POSTGRESQL_URL_APP")
	}

	sql_db, err := sql.Open("postgres", db_url)
	if err != nil {
		log.Fatal(err)
	}

	if err = sql_db.Ping(); err != nil {
		log.Fatal(err)
	}

	DB = sql_db

	log.Println("Database connection successful")
}

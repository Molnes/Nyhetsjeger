package database

import (
	"database/sql"

	_ "github.com/lib/pq"
)

// var (
// Globally available database connection
// DB *sql.DB
// )

// initializes the database connection by loading the environment variables from the .env file,
//
// Sets globally available DB to the initialized database connection.
// func init() {
// 	err := godotenv.Load()
// 	if err != nil {
// 		log.Fatal("Error loading .env")
// 	}

// 	db_url, ok := os.LookupEnv("POSTGRESQL_URL_APP")
// 	if !ok {
// 		log.Fatal("No database url provided. Expected POSTGRESQL_URL_APP")
// 	}

// 	sql_db, err := sql.Open("postgres", db_url)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	if err = sql_db.Ping(); err != nil {
// 		log.Fatal(err)
// 	}

// 	DB = sql_db

// 	log.Println("Database connection successful")
// }

func NewDatabaseConnection(db_url string) (*sql.DB, error) {
	sql_db, err := sql.Open("postgres", db_url)
	if err != nil {
		return nil, err
	}

	if err = sql_db.Ping(); err != nil {
		return nil, err
	}

	return sql_db, nil
}

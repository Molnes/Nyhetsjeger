package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
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

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env")
	}

	creds := get_db_credentials()

	connection_string :=
		fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			creds.db_host, creds.db_port, creds.user, creds.password, creds.db_name)

	sql_db, err := sql.Open("postgres", connection_string)
	if err != nil {
		log.Fatal(err)
	}

	if err = sql_db.Ping(); err != nil {
		log.Fatal(err)
	}

	DB = sql_db

	log.Println("Database connection successful")
}

// Represents database credentials
type db_credentials struct {
	db_host  string
	db_port  string
	user     string
	password string
	db_name  string
}

// Retrieves database credentials from the environment variables
func get_db_credentials() db_credentials {
	db_host, host_ok := os.LookupEnv("DB_HOST")
	if !host_ok {
		log.Fatal("No database host provided")
	}
	db_port, port_ok := os.LookupEnv("DB_PORT")
	if !port_ok {
		log.Fatal("No database port provided")
	}
	user, usr_ok := os.LookupEnv("DB_USR_APP")
	if !usr_ok {
		log.Fatal("No database user provided")
	}
	password, pass_ok := os.LookupEnv("DB_PASSWORD_APP")
	if !pass_ok {
		log.Fatal("No database password provided")
	}
	db_name, name_ok := os.LookupEnv("DB_NAME")
	if !name_ok {
		log.Fatal("No database name provided")
	}
	return db_credentials{
		db_host,
		db_port,
		user,
		password,
		db_name,
	}
}

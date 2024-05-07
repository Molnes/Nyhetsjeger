package main

import (
	"log"
	"os"

	"github.com/Molnes/Nyhetsjeger/db/db_populator"
	"github.com/Molnes/Nyhetsjeger/internal/database"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()

	if err != nil {
		log.Default().Println("Error loading .env file")
	}

	dburl, ok := os.LookupEnv("POSTGRESQL_URL_DEV")
	if !ok {
		log.Fatal("DB Populator: No database url provided. Expected POSTGRESQL_URL_DEV")
	}

	db, err := database.NewDatabaseConnection(dburl)
	if err != nil {
		log.Fatal("DB Populator: Error connecting to database: ", err)
	}

	defer db.Close()

	log.Println("----- Populating database -----")
	defer log.Println("----- Database populated -----")

	db_populator.PopulateDbWithTestData(db)
}

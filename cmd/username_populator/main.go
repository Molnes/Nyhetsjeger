package main

import (
	"context"
	"database/sql"
	"log"
	"os"

	"github.com/Molnes/Nyhetsjeger/internal/database"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("DB Populator: Error loading .env")
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

	populate_adjectives(db)
}

func populate_adjectives(db *sql.DB) {
	// Populate adjectives with raud, fin and brennande.
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
	}

	tx.Exec("INSERT INTO adjectives VALUES ('raud')")
	tx.Exec("INSERT INTO adjectives VALUES ('fin')")
	tx.Exec("INSERT INTO adjectives VALUES ('brennande')")

	if err := tx.Commit(); err != nil {
		log.Println(err)
	}
}
}

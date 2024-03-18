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
	populate_nouns(db)
	/* populate_usernames(db) */
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

func populate_nouns(db *sql.DB) {
	// Populate nouns with lefse, taco, and, stol and appelsin.
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
	}

	tx.Exec("INSERT INTO nouns VALUES ('lefse')")
	tx.Exec("INSERT INTO nouns VALUES ('taco')")
	tx.Exec("INSERT INTO nouns VALUES ('and')")
	tx.Exec("INSERT INTO nouns VALUES ('stol')")
	tx.Exec("INSERT INTO nouns VALUES ('appelsin')")

	if err := tx.Commit(); err != nil {
		log.Println(err)
	}
}

func populate_usernames(db *sql.DB) {
	// Populate usernames with raudlefse, fintaco, brennandeand, raudstol, and finappelsin.
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
	}

	tx.Exec("INSERT INTO usernames VALUES ('raud', 'lefse')")

	if err := tx.Commit(); err != nil {
		log.Println(err)
	}
}

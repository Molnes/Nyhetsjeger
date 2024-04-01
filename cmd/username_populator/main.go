package main

import (
	"context"
	"database/sql"
	"encoding/csv"
	"log"
	"os"

	"github.com/Molnes/Nyhetsjeger/internal/database"
	"github.com/joho/godotenv"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
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

	// populate_adjectives(db)
	// populate_nouns(db)

	loadDataFromCSV(db)
}

// Populates the adjectives table
func populateAdjectives(db *sql.DB) {

	db.Exec("INSERT INTO adjectives VALUES ('Raud'), ('Fin'), ('Brennande')")
}

// Populates the nouns table
func populateNouns(db *sql.DB) {

	db.Exec("INSERT INTO nouns VALUES ('Lefse'), ('Taco'), ('And'), ('Stol'), ('Appelsin')")

}

// Loads data from a csv file into the adjectives and nouns tables.
// Hardcoded to be used with the whitelist-words.csv file.
func loadDataFromCSV(db *sql.DB) {
	file, err := os.Open("data/whitelist-words.csv")
	if err != nil {
		log.Println(err)
	}
	defer file.Close()

	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
	}

	r := csv.NewReader(file)
	r.Comma = ';'
	r.Read() // Skip header
	records, err := r.ReadAll()
	if err != nil {
		log.Println(err)
	}

	c := cases.Title(language.Norwegian)


	for _, record := range records {

		//In case of an unbalanced csv file
		if len(record[0]) > 0 {
			tx.Exec("INSERT INTO adjectives VALUES ($1)", c.String(record[0]))
		}
		if len(record[1]) > 0 {
			tx.Exec("INSERT INTO nouns VALUES ($1)", c.String(record[1]))
		}
	}

	if err := tx.Commit(); err != nil {
		log.Println(err)
	}
}

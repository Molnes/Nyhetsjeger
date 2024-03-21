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

	populate_adjectives(db)
	populate_nouns(db)
	/* populate_usernames(db) */

	/* loadDataFromCSV(db) */
}

// Populates the adjectives table
func populate_adjectives(db *sql.DB) {
	// Populate adjectives with raud, fin and brennande.
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
	}

	tx.Exec("INSERT INTO adjectives VALUES ('Raud')")
	tx.Exec("INSERT INTO adjectives VALUES ('Fin')")
	tx.Exec("INSERT INTO adjectives VALUES ('Brennande')")

	if err := tx.Commit(); err != nil {
		log.Println(err)
	}
}

// Populates the nouns table
func populate_nouns(db *sql.DB) {
	// Populate nouns with lefse, taco, and, stol and appelsin.
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
	}

	tx.Exec("INSERT INTO nouns VALUES ('Lefse')")
	tx.Exec("INSERT INTO nouns VALUES ('Taco')")
	tx.Exec("INSERT INTO nouns VALUES ('And')")
	tx.Exec("INSERT INTO nouns VALUES ('Stol')")
	tx.Exec("INSERT INTO nouns VALUES ('Appelsin')")

	if err := tx.Commit(); err != nil {
		log.Println(err)
	}
}

// Populates the usernames table
// This solution will not work at this moment.
// Function will be used later to debug and test.
func populate_usernames(db *sql.DB) {
	// Populate usernames with raudlefse, fintaco, brennandeand, raudstol, and finappelsin.
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
	}

	tx.Exec("INSERT INTO usernames VALUES ('Raud', 'Lefse')")

	if err := tx.Commit(); err != nil {
		log.Println(err)
	}
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

package main

import (
	"log"
	"time"

	"github.com/Molnes/Nyhetsjeger/internal/database"
	"github.com/google/uuid"
)

func main() {
	defer database.DB.Close()

	log.Println("----- Populating database -----")
	defer log.Println("----- Database populated -----")

	now := time.Now()
	println(now.String())
	to := now.AddDate(0, 0, 1)
	var id uuid.UUID
	err := database.DB.QueryRow(
		`INSERT INTO quizzes (title, available_from, available_to)
		values ($1, $2, $3)
		RETURNING id;`,
		"Test quiz", now, to).Scan(&id)

	if err != nil {
		log.Println(err)
	}

	println(id.String())


	runQuery("INSERT INTO questions (question, arrangement) VALUES ('What is the capital of Norway?',1)")

	runQuery("INSERT INTO users (email, phone, opt_in_ranking) VALUES ('email@example.com', '00000000', true)")

}

func runQuery(query string) {
	_, e := database.DB.Query(query)
	if e != nil {
		log.Println(e)
	}
}

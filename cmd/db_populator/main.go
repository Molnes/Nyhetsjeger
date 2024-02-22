package main

import (
	"log"

	"github.com/Molnes/Nyhetsjeger/internal/database"
)

func main() {
	defer database.DB.Close()

	log.Println("----- Populating database -----")
	defer log.Println("----- Database populated -----")

	runQuery("INSERT INTO questions (question, arrangement) VALUES ('What is the capital of Norway?',1)")

	runQuery("INSERT INTO users (email, phone, opt_in_ranking) VALUES ('email@example.com', '00000000', true)")

}

func runQuery(query string) {
	_, e := database.DB.Query(query)
	if e != nil {
		log.Println(e)
	}
}

package main

import (
	"log"

	"github.com/Molnes/Nyhetsjeger/internal/database"
)

func main() {
	defer database.DB.Close()

	runQuery("INSERT INTO questions (question, article_id) VALUES ('What is the capital of Norway?', 1)")

	runQuery("INSERT INTO users (email, phone) VALUES ('email@example.com', '00000000')")

}

func runQuery(query string) {
	_, e := database.DB.Query(query)
	if e != nil {
		log.Println(e)
	}
}

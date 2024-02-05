package main

import (
	"log"

	"github.com/Molnes/Nyhetsjeger/internal/database"
)

func main(){
	defer database.DB.Close()

	database.DB.Query("INSERT INTO questions (question, article_id) VALUES ('What is the capital of Norway?', 1)")

	log.Println("Database populated!")
}
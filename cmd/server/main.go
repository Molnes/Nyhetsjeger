package main

import (
	"log"

	api "github.com/Molnes/Nyhetsjeger/internal/web_server"
	"github.com/joho/godotenv"

	_ "github.com/lib/pq"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Default().Println("Error loading .env file")
	}

	api.Api()
}

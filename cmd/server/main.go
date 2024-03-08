package main

import (
	"log"

	"github.com/Molnes/Nyhetsjeger/internal/api"
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

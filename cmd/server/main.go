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
		log.Fatal("Error loading .env")
	}

	api.Api()
}

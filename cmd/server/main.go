package main

import (
	"log"

	"github.com/Molnes/Nyhetsjeger/internal/api"
	"github.com/Molnes/Nyhetsjeger/internal/database"
	"github.com/joho/godotenv"

	_ "github.com/lib/pq"
)

// Init runs before main, loads the environment variables from the .env file
func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env")
	}
}

func main() {
	// close db connection when main exits
	defer database.DB.Close()

	api.Api()
}

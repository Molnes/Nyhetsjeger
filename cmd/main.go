package main

import (
	"log"

	"github.com/Molnes/Nyhetsjeger/api"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
	  log.Fatal("Error loading .env")
	}
	
	api.Api()
}

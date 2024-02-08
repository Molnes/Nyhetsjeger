package api

import (
	"fmt"
	"log"
	"os"

	router "github.com/Molnes/Nyhetsjeger/internal/api/web"
	"github.com/labstack/echo/v4"
)

// Sets up the web server and starts it.
//
// Tries reading the PORT environment variable, and falls back to 8080 if it's not found.
func Api() {
	e := echo.New()

	router.SetupRouter(e)

	port, ok := os.LookupEnv("PORT")
	if !ok {
		log.Println("No PORT found, using 8080")
		port = "8080"
	}
	address := fmt.Sprint(":", port)

	e.Logger.Fatal(e.Start(address))
}

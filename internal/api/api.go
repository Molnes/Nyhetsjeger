package api

import (
	"fmt"
	"log"
	"os"

	router "github.com/Molnes/Nyhetsjeger/internal/api/web"
	"github.com/labstack/echo/v4"
)

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

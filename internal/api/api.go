package api

import (
	"fmt"
	"log"
	"os"

	"github.com/Molnes/Nyhetsjeger/internal/api/web/handlers"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func Api() {
	e := echo.New()
	e.Pre(middleware.RemoveTrailingSlash())

	handlers.RegisterQuizHandlers(e);
	handlers.RegisterDashboardHandlers(e);
	handlers.RegisterResourcesHandlers(e);


	port, ok := os.LookupEnv("PORT")
	if !ok {
		log.Println("No PORT found, using 8080")
		port = "8080"
	}
	address := fmt.Sprint(":", port)

	e.Logger.Fatal(e.Start(address))
}

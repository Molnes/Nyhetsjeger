package router

import (
	"github.com/Molnes/Nyhetsjeger/internal/api/web/handlers"
	"github.com/Molnes/Nyhetsjeger/internal/api/web/handlers/api"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

// Sets up the router for the web server
// Takes care of grouping routes, setting up middleware and registering handlers.
func SetupRouter(e *echo.Echo) {
	e.Logger.SetLevel(log.DEBUG)
	e.Pre(middleware.RemoveTrailingSlash())
	e.Use(middleware.Logger())

	handlers.RegisterQuizHandlers(e)
	handlers.RegisterDashboardHandlers(e)

	api_group := e.Group("/api/v1")

	quiz_api_group := api_group.Group("/quiz")
	api.RegisterQuizApiHandlers(quiz_api_group)

	amin_api_group := api_group.Group("/admin")
	api.RegisterAdminApiHandlers(amin_api_group)

	e.Static("/static", "assets")
}

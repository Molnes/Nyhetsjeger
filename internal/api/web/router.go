package router

import (
	"github.com/Molnes/Nyhetsjeger/internal/api/middlewares"
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

	handlers.RegisterPublicPages(e)

	// authentication routes
	auth_group := e.Group("/auth")
	handlers.RegisterAuthHandlers(auth_group)

	// TODO add homepage with login button if not authenticated

	quizGroup := e.Group("/quiz")
	quizGroup.Use(middlewares.IsAuthenticated)
	handlers.RegisterQuizHandlers(quizGroup)

	dashboardGroup := e.Group("/dashboard")
	dashboardGroup.Use(middlewares.IsAdmin)

	handlers.RegisterDashboardHandlers(dashboardGroup)

	api_group := e.Group("/api/v1")
	api_group.Use(middlewares.IsAuthenticated)

	quiz_api_group := api_group.Group("/quiz")
	api.RegisterQuizApiHandlers(quiz_api_group)

	admin_api_group := api_group.Group("/admin")
	admin_api_group.Use(middlewares.IsAdmin)
	api.RegisterAdminApiHandlers(admin_api_group)

	e.Static("/static", "assets")

	// websocket for live updates
	e.GET("/ws", handlers.WebsocketHandler)

}

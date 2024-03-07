package router

import (
	"fmt"
	"os"

	"github.com/Molnes/Nyhetsjeger/internal/api/middlewares"
	"github.com/Molnes/Nyhetsjeger/internal/api/web/handlers"
	"github.com/Molnes/Nyhetsjeger/internal/api/web/handlers/api"
	"github.com/Molnes/Nyhetsjeger/internal/config"
	"github.com/Molnes/Nyhetsjeger/internal/data/users/user_roles"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"golang.org/x/oauth2"
)

// Sets up the router for the web server
// Takes care of grouping routes, setting up middleware and registering handlers.
func SetupRouter(e *echo.Echo, sharedData *config.SharedData, oauthConfig *oauth2.Config) {

	e.Logger.SetLevel(log.DEBUG)
	e.Pre(middleware.RemoveTrailingSlash())
	e.Use(middleware.Logger())

	secureConfig := middleware.DefaultSecureConfig

	currentCsp := secureConfig.ContentSecurityPolicy
	allowedFrameAncestors := os.Getenv("ALLOWED_FRAME_ANCESTORS")
	secureConfig.ContentSecurityPolicy = fmt.Sprintf("frame-ancestors 'self' %s; %s", allowedFrameAncestors, currentCsp)

	e.Use(middleware.SecureWithConfig(secureConfig))

	// pages nor requiring authentication
	publicPagesHandler := handlers.NewPublicPagesHandler(sharedData)
	publicPagesHandler.RegisterPublicPages(e)

	// authentication routes, no authentication required
	authGroup := e.Group("/auth")
	authHandlers := handlers.NewAuthHandler(sharedData, oauthConfig)
	authHandlers.RegisterAuthHandlers(authGroup)

	// routes requiring authentication
	quizGroup := e.Group("/quiz")
	authForceWithRedirect := middlewares.NewAuthenticationMiddleware(sharedData, true)
	quizGroup.Use(authForceWithRedirect.EncofreAuthentication)

	// quiz pages
	quizPagesHandler := handlers.NewQuizPagesHandler(sharedData)
	quizPagesHandler.RegisterQuizHandlers(quizGroup)

	// routes requiring admin
	enforceAdminMiddlewareRedirect :=
		middlewares.NewAuthorizationMiddleware(
			sharedData,
			[]user_roles.Role{
				user_roles.QuizAdmin,
				user_roles.OrganizationAdmin,
			}, true)
	dashboardGroup := e.Group("/dashboard")
	dashboardGroup.Use(authForceWithRedirect.EncofreAuthentication)
	dashboardGroup.Use(enforceAdminMiddlewareRedirect.EnforceRole)

	// dashboard pages
	dashboardPagesHandler := handlers.NewDashboardPagesHandler(sharedData)
	dashboardPagesHandler.RegisterDashboardHandlers(dashboardGroup)

	// api routes, requiring authentication
	apiGroup := e.Group("/api/v1")
	authForce := middlewares.NewAuthenticationMiddleware(sharedData, false)
	apiGroup.Use(authForce.EncofreAuthentication)

	quizApiGroup := apiGroup.Group("/quiz")
	quizApiHandler := api.NewQuizApiHandler(sharedData)
	quizApiHandler.RegisterQuizApiHandlers(quizApiGroup)

	// admin api routes, requiring admin
	adminApiGroup := apiGroup.Group("/admin")
	enforceAdminMiddleware :=
		middlewares.NewAuthorizationMiddleware(
			sharedData,
			[]user_roles.Role{
				user_roles.QuizAdmin,
				user_roles.OrganizationAdmin,
			}, false)
	adminApiGroup.Use(enforceAdminMiddleware.EnforceRole)

	adminApiHandler := api.NewAdminApiHandler(sharedData)
	adminApiHandler.RegisterAdminApiHandlers(adminApiGroup)

	// static files
	e.Static("/static", "assets")

	// websocket for live updates
	e.GET("/ws", handlers.WebsocketHandler)

}

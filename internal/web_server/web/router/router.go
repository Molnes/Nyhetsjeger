package router

import (
	"fmt"
	"os"

	"github.com/Molnes/Nyhetsjeger/internal/config"
	"github.com/Molnes/Nyhetsjeger/internal/models/users/user_roles"
	"github.com/Molnes/Nyhetsjeger/internal/web_server/middlewares"
	"github.com/Molnes/Nyhetsjeger/internal/web_server/web/handlers"
	"github.com/Molnes/Nyhetsjeger/internal/web_server/web/handlers/api"
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

	forceAcceptedTermsWithRedirect := middlewares.NewAcceptedTerms(sharedData, true)
	quizGroup.Use(forceAcceptedTermsWithRedirect.EncofreAcceptedTerms)

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
			})
	dashboardGroup := e.Group("/dashboard")
	dashboardGroup.Use(authForceWithRedirect.EncofreAuthentication)
	dashboardGroup.Use(enforceAdminMiddlewareRedirect.EnforceRole)

	// dashboard pages
	dashboardPagesHandler := handlers.NewDashboardPagesHandler(sharedData)
	dashboardPagesHandler.RegisterDashboardHandlers(dashboardGroup)

	// api routes
	apiGroup := e.Group("/api/v1")
	apiGroup.Use(handlers.SetApiErrorDisplay)

	guestGroup := apiGroup.Group("/guest")
	guestApiHandler := api.NewPublicApiHandler(sharedData)
	guestApiHandler.RegisterPublicApiHandlers(guestGroup)

	authForce := middlewares.NewAuthenticationMiddleware(sharedData, false)

	quizApiGroup := apiGroup.Group("/quiz")
	quizApiGroup.Use(authForce.EncofreAuthentication)

	forceAcceptedTermsNoRedirect := middlewares.NewAcceptedTerms(sharedData, false)
	quizGroup.Use(forceAcceptedTermsNoRedirect.EncofreAcceptedTerms)

	quizApiHandler := api.NewQuizApiHandler(sharedData)
	quizApiHandler.RegisterQuizApiHandlers(quizApiGroup)

	// admin api routes, requiring admin
	adminApiGroup := apiGroup.Group("/admin")
	adminApiGroup.Use(authForce.EncofreAuthentication)

	enforceAdminMiddleware :=
		middlewares.NewAuthorizationMiddleware(
			sharedData,
			[]user_roles.Role{
				user_roles.QuizAdmin,
				user_roles.OrganizationAdmin,
			})
	adminApiGroup.Use(enforceAdminMiddleware.EnforceRole)

	adminApiHandler := api.NewAdminApiHandler(sharedData)
	adminApiHandler.RegisterAdminApiHandlers(adminApiGroup)

	organizationAdminApiGroup := apiGroup.Group("/organization-admin")
	enforceOrganizationAdminMiddleware :=
		middlewares.NewAuthorizationMiddleware(
			sharedData,
			[]user_roles.Role{user_roles.OrganizationAdmin})
	organizationAdminApiGroup.Use(enforceOrganizationAdminMiddleware.EnforceRole)

	organizationAdminApiHandler := api.NewOrganizationAdminApiHandler(sharedData)
	organizationAdminApiHandler.RegisterOrganizationAdminHandlers(organizationAdminApiGroup)

	// static files
	e.Static("/static", "assets")
	e.File("/favicon.ico", "assets/favicon.ico")

	// websocket for live updates
	e.GET("/ws", handlers.WebsocketHandler)

	e.HTTPErrorHandler = handlers.HTTPErrorHandler

}

package handlers

import (
	"net/http"

	"github.com/Molnes/Nyhetsjeger/internal/api/web/views/pages/dashboard_pages"
	"github.com/Molnes/Nyhetsjeger/internal/utils"
	"github.com/labstack/echo/v4"
)

// Registers handlers for dashboard related pages
func RegisterDashboardHandlers(e *echo.Group) {
	e.GET("", dashboardHomePage)
	e.GET("/create-quiz", dashboardCreateQuiz)
}

// Renders the dashboard home page
func dashboardHomePage(c echo.Context) error {
	return utils.Render(c, http.StatusOK, dashboard_pages.DashboardPage())
}

// Renders the page for creating a new quiz
func dashboardCreateQuiz(c echo.Context) error {
	return utils.Render(c, http.StatusOK, dashboard_pages.CreateQuiz())
}

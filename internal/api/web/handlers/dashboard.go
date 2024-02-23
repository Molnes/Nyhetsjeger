package handlers

import (
	"net/http"

	"github.com/Molnes/Nyhetsjeger/internal/api/web/views/pages/dashboard_pages"
	"github.com/Molnes/Nyhetsjeger/internal/data/quizzes"
	"github.com/Molnes/Nyhetsjeger/internal/utils"
	"github.com/labstack/echo/v4"
)

// Registers handlers for dashboard related pages
func RegisterDashboardHandlers(e *echo.Group) {
	e.GET("", dashboardHomePage)
	e.GET("/quiz/edit/:quizId", dashboardEditQuiz)
}

// Renders the dashboard home page
func dashboardHomePage(c echo.Context) error {
	return utils.Render(c, http.StatusOK, dashboard_pages.DashboardPage())
}

// Renders the page for creating a new quiz
func dashboardEditQuiz(c echo.Context) error {
	// TODO: Fetch the quiz from the database
	quiz, _ := quizzes.CreateDefaultQuiz()

	return utils.Render(c, http.StatusOK, dashboard_pages.EditQuiz(quiz))
}

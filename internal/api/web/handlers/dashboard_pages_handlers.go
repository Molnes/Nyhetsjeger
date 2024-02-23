package handlers

import (
	"net/http"

	"github.com/Molnes/Nyhetsjeger/internal/api/web/views/pages/dashboard_pages"
	"github.com/Molnes/Nyhetsjeger/internal/config"
	"github.com/Molnes/Nyhetsjeger/internal/data/articles"
	"github.com/Molnes/Nyhetsjeger/internal/data/quizzes"
	"github.com/Molnes/Nyhetsjeger/internal/utils"
	"github.com/labstack/echo/v4"
)

type DashboardPagesHandler struct {
	sharedData *config.SharedData
}

// Creates a new DashboardPagesHandler
func NewDashboardPagesHandler(sharedData *config.SharedData) *DashboardPagesHandler {
	return &DashboardPagesHandler{sharedData}
}

// Registers handlers for dashboard related pages
func (dph *DashboardPagesHandler) RegisterDashboardHandlers(e *echo.Group) {
	e.GET("", dph.dashboardHomePage)
	e.GET("/quiz/edit/:quizId", dph.dashboardEditQuiz)
}

// Renders the dashboard home page
func (dph *DashboardPagesHandler) dashboardHomePage(c echo.Context) error {
	return utils.Render(c, http.StatusOK, dashboard_pages.DashboardPage())
}

// Renders the page for creating a new quiz
func (dph *DashboardPagesHandler) dashboardEditQuiz(c echo.Context) error {
	// TODO: Fetch the quiz from the database
	quiz, _ := quizzes.CreateDefaultQuiz()
	articles, _ := articles.GetAllArticles()

	return utils.Render(c, http.StatusOK, dashboard_pages.EditQuiz(&quiz, &articles))
}

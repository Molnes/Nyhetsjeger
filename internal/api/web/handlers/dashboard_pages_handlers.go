package handlers

import (
	"net/http"

	"github.com/Molnes/Nyhetsjeger/internal/api/web/views/pages/dashboard_pages"
	"github.com/Molnes/Nyhetsjeger/internal/config"
	"github.com/Molnes/Nyhetsjeger/internal/data/articles"
	"github.com/Molnes/Nyhetsjeger/internal/data/quizzes"
	"github.com/Molnes/Nyhetsjeger/internal/utils"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type DashboardPagesHandler struct {
	sharedData *config.SharedData
}

// Creates a new DashboardPagesHandler.
func NewDashboardPagesHandler(sharedData *config.SharedData) *DashboardPagesHandler {
	return &DashboardPagesHandler{sharedData}
}

// Registers handlers for dashboard related pages.
func (dph *DashboardPagesHandler) RegisterDashboardHandlers(e *echo.Group) {
	e.GET("", dph.dashboardHomePage)
	e.GET("/quiz/edit/:quizId", dph.dashboardEditQuiz)
}

// Renders the dashboard home page.
func (dph *DashboardPagesHandler) dashboardHomePage(c echo.Context) error {
	return utils.Render(c, http.StatusOK, dashboard_pages.DashboardPage())
}

// Renders the page for editing quiz.
func (dph *DashboardPagesHandler) dashboardEditQuiz(c echo.Context) error {
	uuid_id, _ := uuid.Parse(c.Param("quizId"))
	quiz, _ := quizzes.GetFullQuizByID(dph.sharedData.DB, uuid_id)

	// Collect all the articles from each question in the quiz.
	// Only add the ones that are valid.
	articles := []articles.Article{}
	for _, question := range quiz.Questions {
		article := question.Article
		if article.ID.Valid {
			articles = append(articles, article)
		}
	}

	return utils.Render(c, http.StatusOK, dashboard_pages.EditQuiz(quiz, &articles))
}

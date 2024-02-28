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
	uuid_id, _ := uuid.Parse(c.Param("quizId"))
	println("Real Quiz ID: " + uuid_id.String())

	quiz, _ := quizzes.GetQuizByID(dph.sharedData.DB, uuid_id)
	println("Quiz ID: " + quiz.ID.String())
	println("Title: " + quiz.Title)
	println("Image: " + quiz.ImageURL.String())
	println("Available from: " + quiz.AvailableFrom.String())
	println("Available to: " + quiz.AvailableTo.String())
	println("Created at: " + quiz.CreatedAt.String())
	println("Last modified at: " + quiz.LastModifiedAt.String())

	// questions := questions.GetQuestionsByQuizID(dph.sharedData.DB, c.Params("quizId"))
	// questions, _ := questions.GetQuestionsByQuizID(dph.sharedData.DB, uuid_id)
	/* for _, question := range *questions {
		println("Question ID:" + question.ID.String())
	} */
	// quiz, _ := quizzes.CreateDefaultQuiz()

	// TODO: Get the actual articles for the quiz
	articles, _ := articles.GetAllArticles()

	return utils.Render(c, http.StatusOK, dashboard_pages.EditQuiz(quiz, &articles))
}

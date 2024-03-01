package handlers

import (
	"net/http"

	dashboard_components "github.com/Molnes/Nyhetsjeger/internal/api/web/views/components/dashboard_components/edit_quiz"
	"github.com/Molnes/Nyhetsjeger/internal/api/web/views/pages/dashboard_pages"
	"github.com/Molnes/Nyhetsjeger/internal/config"
	"github.com/Molnes/Nyhetsjeger/internal/data/articles"
	"github.com/Molnes/Nyhetsjeger/internal/data/questions"
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
	e.GET("/edit-quiz", dph.dashboardEditQuiz)
	e.GET("/edit-quiz/new-question", dph.dashboardNewQuestionModal)
}

// Renders the dashboard home page.
func (dph *DashboardPagesHandler) dashboardHomePage(c echo.Context) error {
	return utils.Render(c, http.StatusOK, dashboard_pages.DashboardPage())
}

// Renders the page for editing quiz.
func (dph *DashboardPagesHandler) dashboardEditQuiz(c echo.Context) error {
	uuid_id, _ := uuid.Parse(c.QueryParam("quiz-id"))
	if uuid_id == uuid.Nil {
		// TODO: Redirect to proper error handling page with descriptive error message.
		return utils.Render(c, http.StatusNotFound, dashboard_pages.DashboardPage())
	}

	quiz, _ := quizzes.GetFullQuizByID(dph.sharedData.DB, uuid_id)

	// Get all the articles for a quiz.
	articles, _ := articles.GetArticlesByQuizID(dph.sharedData.DB, uuid_id)

	return utils.Render(c, http.StatusOK, dashboard_pages.EditQuiz(quiz, articles))
}

// Renders the modal for creating a new question.
func (dph *DashboardPagesHandler) dashboardNewQuestionModal(c echo.Context) error {
	quiz_id, _ := uuid.Parse(c.QueryParam("quiz-id"))

	// Create a new question with no actual data.
	// Set the default points to be 10.
	newQuestion := questions.Question{
		ID:           uuid.New(),
		Text:         "",
		Article:      articles.Article{},
		QuizID:       quiz_id,
		Points:       10,
		Alternatives: []questions.Alternative{},
	}

	// Save the new question to the database.
	questions.PostNewQuestion(dph.sharedData.DB, newQuestion)

	// Get all the articles.
	articles, _ := articles.GetArticlesByQuizID(dph.sharedData.DB, quiz_id)

	return utils.Render(c, http.StatusOK, dashboard_components.EditQuestionModal(articles))
}

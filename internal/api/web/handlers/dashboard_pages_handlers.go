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
	// e.GET("/edit-quiz/new-question", dph.das)
}

// Renders the dashboard home page.
func (dph *DashboardPagesHandler) dashboardHomePage(c echo.Context) error {
	return utils.Render(c, http.StatusOK, dashboard_pages.DashboardPage())
}

// Renders the page for editing quiz.
func (dph *DashboardPagesHandler) dashboardEditQuiz(c echo.Context) error {
	uuid_id, _ := uuid.Parse(c.QueryParam("quiz-id"))
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

	question_id, _ := questions.PostNewQuestion(dph.sharedData.DB, newQuestion)

	// TODO: Create a new question in the DB and return the question ID.
	// Return the modal.

	return utils.Render(c, http.StatusOK, dashboard_components.EditQuestionModal())
}

package api

import (
	"net/http"
	"time"

	"github.com/Molnes/Nyhetsjeger/internal/config"
	"github.com/Molnes/Nyhetsjeger/internal/models/quizzes"
	"github.com/Molnes/Nyhetsjeger/internal/utils"
	dashboard_components "github.com/Molnes/Nyhetsjeger/internal/web_server/web/views/components/dashboard_components/edit_quiz"
	"github.com/Molnes/Nyhetsjeger/internal/web_server/web/views/pages/dashboard_pages"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type AdminApiHandler struct {
	sharedData *config.SharedData
}

// Creates a new AdminApiHandler
func NewAdminApiHandler(sharedData *config.SharedData) *AdminApiHandler {
	return &AdminApiHandler{sharedData}
}

// Registers handlers for admin api
func (aah *AdminApiHandler) RegisterAdminApiHandlers(e *echo.Group) {
	e.POST("/quiz/create-new", aah.PostDefaultQuiz)
	e.POST("/quiz/edit-title", aah.EditQuizTitle)
	e.DELETE("/delete-quiz", aah.DeleteQuiz)
}

// Handles the creation of a new default quiz in the DB.
// Redirects to the edit quiz page for the newly created quiz.
func (aah *AdminApiHandler) PostDefaultQuiz(c echo.Context) error {
	quiz := quizzes.CreateDefaultQuiz()
	quizzes.CreateQuiz(aah.sharedData.DB, quiz)

	c.Response().Header().Set("HX-Redirect", "/dashboard/edit-quiz?quiz-id="+quiz.ID.String())
	return c.Redirect(http.StatusOK, "/dashboard/edit-quiz?quiz-id="+quiz.ID.String())
}

func (aah *AdminApiHandler) EditQuizTitle(c echo.Context) error {
	// Get the quiz ID
	quiz_id, err := uuid.Parse(c.QueryParam("quiz-id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid or missing quiz id")
	}

	// Update the quiz title
	title := c.FormValue(dashboard_pages.QuizTitle)
	err = quizzes.UpdateTitleByQuizID(aah.sharedData.DB, quiz_id, title)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Failed to update quiz title")
	}

	time.Sleep(2 * time.Second)

	return utils.Render(c, http.StatusOK, dashboard_components.EditTitleInput(title, quiz_id.String(), dashboard_pages.QuizTitle))
}

// Deletes a quiz from the database.
func (aah *AdminApiHandler) DeleteQuiz(c echo.Context) error {
	quiz_id, err := uuid.Parse(c.QueryParam("quiz-id"))

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid or missing quiz id")
	}

	quizzes.DeleteQuizByID(aah.sharedData.DB, quiz_id)

	c.Response().Header().Set("HX-Redirect", "/dashboard")
	return c.Redirect(http.StatusOK, "/dashboard")
}

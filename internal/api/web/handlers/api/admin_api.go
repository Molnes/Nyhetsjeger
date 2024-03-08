package api

import (
	"net/http"

	"github.com/Molnes/Nyhetsjeger/internal/config"
	"github.com/Molnes/Nyhetsjeger/internal/data/quizzes"
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
	e.POST("/quiz/create-new", PostQuiz)
	e.DELETE("/delete-quiz", aah.DeleteQuiz)
}

// Handles the creation of a new quiz.
func PostQuiz(c echo.Context) error {
	// TODO: Add quiz to the database

	quiz, _ := quizzes.CreateDefaultQuiz()

	c.Response().Header().Set("HX-Redirect", "/dashboard/edit-quiz?quiz-id="+quiz.ID.String())
	return c.Redirect(http.StatusOK, "/dashboard/edit-quiz?quiz-id="+quiz.ID.String())
}

// Deletes a quiz from the database.
func (aah *AdminApiHandler) DeleteQuiz(c echo.Context) error {
	quiz_id, _ := uuid.Parse(c.QueryParam("quiz-id"))
	quizzes.DeleteQuizByID(aah.sharedData.DB, quiz_id)

	c.Response().Header().Set("HX-Redirect", "/dashboard")
	return c.Redirect(http.StatusOK, "/dashboard")
}

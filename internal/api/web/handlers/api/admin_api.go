package api

import (
	"net/http"

	"github.com/Molnes/Nyhetsjeger/internal/config"
	"github.com/Molnes/Nyhetsjeger/internal/data/quizzes"
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
}

// Handles the creation of a new quiz.
func PostQuiz(c echo.Context) error {
	// TODO: Add quiz to the database

	quiz, _ := quizzes.CreateDefaultQuiz()

	c.Response().Header().Set("HX-Redirect", "/dashboard/edit-quiz?quiz_id="+quiz.ID.String())
	return c.Redirect(http.StatusOK, "/dashboard/edit-quiz?quiz_id="+quiz.ID.String())
}

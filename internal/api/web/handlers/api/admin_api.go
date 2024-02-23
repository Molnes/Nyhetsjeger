package api

import (
	"net/http"

	"github.com/Molnes/Nyhetsjeger/internal/data/quizzes"
	"github.com/labstack/echo/v4"
)

func RegisterAdminApiHandlers(e *echo.Group) {
	e.POST("/quiz/create-new", PostQuiz)
}

// Handles the creation of a new quiz.
func PostQuiz(c echo.Context) error {
	// TODO: Add quiz to the database

	quiz, _ := quizzes.CreateDefaultQuiz()

	c.Response().Header().Set("HX-Redirect", "/dashboard/quiz/edit/"+quiz.ID.String())
	return c.Redirect(http.StatusOK, "/dashboard/quiz/edit/"+quiz.ID.String())
}

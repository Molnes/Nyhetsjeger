package api

import (
	"github.com/Molnes/Nyhetsjeger/internal/data/quizzes"
	"github.com/labstack/echo/v4"
)

func RegisterAdminApiHandlers(e *echo.Group) {
	e.POST("/create-new", PostQuiz)
}

func PostQuiz(c echo.Context) error {
	// TODO: Add quiz to the database

	quiz, _ := quizzes.CreateDefaultQuiz()
	return c.JSON(200, quiz)
}

package handlers

import (
	"net/http"

	"github.com/Molnes/Nyhetsjeger/internal/api/web/views/pages/quiz_pages"
	"github.com/Molnes/Nyhetsjeger/internal/data/quizzes"
	"github.com/Molnes/Nyhetsjeger/internal/utils"
	"github.com/labstack/echo/v4"
)

// Registers handlers for quiz related pages
func RegisterQuizHandlers(e *echo.Echo) {
	e.GET("/", quizHomePage)
	e.GET("/quizpage", GetQuizPage)
}

// Renders the quiz home page
func quizHomePage(c echo.Context) error {
	return utils.Render(c, http.StatusOK, quiz_pages.QuizHomePage())
}

func GetQuizPage(c echo.Context) error {
	sampleQuiz := quizzes.SampleQuiz.Questions[len(quizzes.SampleQuiz.Questions)-1]

	return utils.Render(c, http.StatusOK, quiz_pages.QuizPage(sampleQuiz))
}

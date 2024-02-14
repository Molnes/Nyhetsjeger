package handlers

import (
	"fmt"
	"net/http"

	"github.com/Molnes/Nyhetsjeger/internal/api/web/views/components/quiz_components"
	"github.com/Molnes/Nyhetsjeger/internal/api/web/views/pages/quiz_pages"
	"github.com/Molnes/Nyhetsjeger/internal/data/quizzes"
	"github.com/Molnes/Nyhetsjeger/internal/utils"
	"github.com/labstack/echo/v4"
)

// Registers handlers for quiz related pages
func RegisterQuizHandlers(e *echo.Echo) {
	e.GET("/", quizHomePage)
	e.GET("/quizpage", GetQuizPage)
	e.GET("/nextquestion", GetNextQuestion)
}

// Renders the quiz home page
func quizHomePage(c echo.Context) error {
	return utils.Render(c, http.StatusOK, quiz_pages.QuizHomePage())
}

func GetQuizPage(c echo.Context) error {
	sampleQuiz := quizzes.SampleQuiz.Questions[len(quizzes.SampleQuiz.Questions)-1]
	title := quizzes.SampleQuiz.Title

	return utils.Render(c, http.StatusTeapot, quiz_pages.QuizPage(sampleQuiz, title))
}

func GetNextQuestion(c echo.Context) error {
	answer := c.QueryParam("answer")
	correct := ""
	alternatives := quizzes.SampleQuiz.Questions[len(quizzes.SampleQuiz.Questions)-1].Alternatives
	for _, aswr := range alternatives {
		if aswr.IsCorrect {
			correct = aswr.Text
		}
	}

	fmt.Println(answer)
	return utils.Render(c, http.StatusOK, quiz_components.Answers(alternatives, quiz_components.CorrectAndAnswered(correct, answer)))
}

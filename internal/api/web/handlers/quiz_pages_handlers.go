package handlers

import (
	"fmt"
	"net/http"

	"github.com/Molnes/Nyhetsjeger/internal/api/web/views/components/quiz_components"
	"github.com/Molnes/Nyhetsjeger/internal/api/web/views/pages/quiz_pages"
	"github.com/Molnes/Nyhetsjeger/internal/config"
	"github.com/Molnes/Nyhetsjeger/internal/data/quizzes"
	"github.com/Molnes/Nyhetsjeger/internal/utils"
	"github.com/labstack/echo/v4"
)

type QuizPagesHandler struct {
	sharedData *config.SharedData
}

func NewQuizPagesHandler(sharedData *config.SharedData) *QuizPagesHandler {
	return &QuizPagesHandler{sharedData}
}

// Registers handlers for quiz pages
func (qph *QuizPagesHandler) RegisterQuizHandlers(e *echo.Group) {
	e.GET("", qph.quizHomePage)
	e.GET("/quizpage", qph.GetQuizPage)
}

// Renders the quiz home page
func (qph *QuizPagesHandler) quizHomePage(c echo.Context) error {
	quizzList, err := quizzes.GetQuizzes(qph.sharedData.DB)
	if err != nil {
		return err
	}
	return utils.Render(c, http.StatusOK, quiz_pages.QuizHomePage(
		quizzList,
	))
}

var questionIndex = 0

// Gets the quiz page
func getQuizPage(c echo.Context) error {
	sampleQuiz := quizzes.SampleQuiz.Questions[questionIndex]
	title := quizzes.SampleQuiz.Title

	return utils.Render(c, http.StatusOK, quiz_pages.QuizPage(sampleQuiz, title))
}

// Checks if the answer was correct, and returns the results
func getIsCorrect(c echo.Context) error {
	answer := c.QueryParam("answer")
	correct := ""
	alternatives := quizzes.SampleQuiz.Questions[questionIndex].Alternatives
	for _, aswr := range alternatives {
		if aswr.IsCorrect {
			correct = aswr.Text
		}
	}

	fmt.Println(answer)
	return utils.Render(c, http.StatusOK, quiz_components.Answers(alternatives, quiz_components.CorrectAndAnswered(correct, answer)))
}

// Posts the next question
func postNextQuestion(c echo.Context) error {
	questionIndex++
	if questionIndex >= len(quizzes.SampleQuiz.Questions) {
		questionIndex = 0
	}

	progress := float64(questionIndex) / float64(len(quizzes.SampleQuiz.Questions))

	return utils.Render(c, http.StatusOK, quiz_components.QuizContent(quizzes.SampleQuiz.Questions[questionIndex], progress))
}

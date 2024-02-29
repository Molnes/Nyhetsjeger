package handlers

import (
	"net/http"

	"github.com/Molnes/Nyhetsjeger/internal/api/web/views/components/quiz_components"
	"github.com/Molnes/Nyhetsjeger/internal/api/web/views/pages/quiz_pages"
	"github.com/Molnes/Nyhetsjeger/internal/config"
	"github.com/Molnes/Nyhetsjeger/internal/data/quizzes"
	"github.com/Molnes/Nyhetsjeger/internal/utils"
	"github.com/google/uuid"
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
	e.GET("/quizpage", qph.getQuizPage)
	e.GET("/checkanswer", qph.getIsCorrect)
	e.POST("/nextquestion", qph.postNextQuestion)
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

// Gets the quiz page
func (qph *QuizPagesHandler) getQuizPage(c echo.Context) error {
	questionId, _ := uuid.Parse(c.QueryParam("questionid"))
	question := quizzes.GetQuestionFromId(questionId)
	title := quizzes.SampleQuiz.Title

	return utils.Render(c, http.StatusOK, quiz_pages.QuizQuestion(question, title))
}

// Checks if the answer was correct, and returns the results
func (qph *QuizPagesHandler) getIsCorrect(c echo.Context) error {
	answer, _ := uuid.Parse(c.QueryParam("answerid"))
	questionId, err := uuid.Parse(c.QueryParam("questionid"))
	correct := uuid.UUID{}

	question := quizzes.GetQuestionFromId(questionId)

	//If the id is wrong, return not found error
	if (question == nil || err != nil) {
		return c.NoContent(http.StatusNotFound)
	}

	alternatives := question.Alternatives

	for _, alternative := range alternatives {
		if alternative.IsCorrect {
			correct = alternative.ID
			break
		}
	}

	return utils.Render(c, http.StatusOK, quiz_components.Answers(alternatives, questionId, quiz_components.CorrectAndAnswered(correct, answer)))
}

// Posts the next question
func (qph *QuizPagesHandler) postNextQuestion(c echo.Context) error {
	questionID, err := uuid.Parse(c.QueryParam("questionid"))

	if err != nil {
		return c.NoContent(http.StatusNotFound)
	}

	questionArrangement := quizzes.GetQuestionFromId(questionID).Arrangement

	progress := float64(questionArrangement) / float64(len(quizzes.SampleQuiz.Questions))

	return utils.Render(c, http.StatusOK, quiz_components.QuizContent(&quizzes.SampleQuiz.Questions[questionArrangement], progress))
}

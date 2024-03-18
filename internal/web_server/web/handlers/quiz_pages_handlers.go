package handlers

import (
	"database/sql"
	"net/http"

	"github.com/Molnes/Nyhetsjeger/internal/config"
	"github.com/Molnes/Nyhetsjeger/internal/models/questions"
	"github.com/Molnes/Nyhetsjeger/internal/models/quizzes"
	"github.com/Molnes/Nyhetsjeger/internal/models/users"
	"github.com/Molnes/Nyhetsjeger/internal/utils"
	"github.com/Molnes/Nyhetsjeger/internal/web_server/web/views/components/quiz_components"
	"github.com/Molnes/Nyhetsjeger/internal/web_server/web/views/pages/quiz_pages"
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
	e.GET("/:id", qph.getQuizPageByQuizID)
	e.GET("/checkanswer", qph.getIsCorrect)
	e.POST("/nextquestion", qph.postNextQuestion)
	e.GET("/quiz-summary", qph.getQuizSummary)

	e.GET("/toppliste", qph.getScoreboard)
	e.GET("/fullforte", qph.getFinishedQuizzes)
	e.GET("/arkiv", qph.getArchivedQuizzes)
	e.GET("/profil", qph.getQuizProfile)
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
	questionId, err := uuid.Parse(c.QueryParam("questionid"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid or missing question id")
	}
	question, err := questions.GetQuestionByID(qph.sharedData.DB, questionId)
	if err != nil {
		if err == sql.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, "No question found")
		} else {
			return err
		}
	}

	title := quizzes.SampleQuiz.Title

	err = users.StartQuestion(qph.sharedData.DB, utils.GetUserIDFromCtx(c), questionId)
	if err != nil {
		return err
	}

	return utils.Render(c, http.StatusOK, quiz_pages.QuizQuestion(question, title))
}

func (qph *QuizPagesHandler) getQuizPageByQuizID(c echo.Context) error {

	quizID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid or missing quiz id")
	}

	questionId, err := questions.GetFirstQuestionID(qph.sharedData.DB, quizID)
	if err != nil {
		if err == sql.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, "No questions found for quiz")
		} else {
			return err
		}
	}

	//redirect to quiz page with first question
	return c.Redirect(http.StatusFound, "/quiz/quizpage?questionid="+questionId.String())
}

// Checks if the answer was correct, and returns the results
func (qph *QuizPagesHandler) getIsCorrect(c echo.Context) error {
	answerId, _ := uuid.Parse(c.QueryParam("answerid"))
	questionId, err := uuid.Parse(c.QueryParam("questionid"))

	question, err := questions.GetQuestionByID(qph.sharedData.DB, questionId)

	//If the id is wrong, return not found error
	if question == nil || err != nil {
		return c.NoContent(http.StatusNotFound)
	}

	alternative, err := questions.GetAlternativeByID(qph.sharedData.DB, answerId)
	if err != nil {
		return c.NoContent(http.StatusNotFound)
	}

	if alternative.QuestionID != questionId {
		return c.NoContent(http.StatusNotFound)
	}

	err = users.AnswerQuestion(qph.sharedData.DB, utils.GetUserIDFromCtx(c), questionId, alternative.ID)

	if err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}

	alternatives, err := questions.GetAlternativesByQuestionID(qph.sharedData.DB, questionId)
	if err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}

	return utils.Render(c, http.StatusOK, quiz_components.Answers((*alternatives), questionId, true, answerId))

}

// Posts the next question
func (qph *QuizPagesHandler) postNextQuestion(c echo.Context) error {
	questionID, err := uuid.Parse(c.QueryParam("questionid"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid or missing question id")
	}
	question, err := questions.GetNextQuestion(qph.sharedData.DB, questionID)
	if err != nil {
		if err == sql.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, "No question found")
		} else {
			return err
		}
	}

	err = users.StartQuestion(qph.sharedData.DB, utils.GetUserIDFromCtx(c), questionID)
	if err != nil {
		return c.NoContent(http.StatusNotFound)
	}

	questionArrangement := question.Arrangement

	progress := float64(questionArrangement) / float64(len(quizzes.SampleQuiz.Questions))

	return utils.Render(c, http.StatusOK, quiz_components.QuizContent(question, progress))
}

func (qph *QuizPagesHandler) getQuizSummary(c echo.Context) error {
	quizID, err := uuid.Parse(c.QueryParam("quizid"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid or missing quiz id")
	}

	quizSummary, err := users.GetQuizSummary(qph.sharedData.DB, utils.GetUserIDFromCtx(c), quizID)
	if err != nil {
		return err
	}

	return utils.Render(c, http.StatusOK, quiz_pages.QuizSummaryPage(quizSummary))

}

func (qph *QuizPagesHandler) getScoreboard(c echo.Context) error {
	return utils.Render(c, http.StatusOK, quiz_pages.Scoreboard())
}

func (qph *QuizPagesHandler) getFinishedQuizzes(c echo.Context) error {
	return utils.Render(c, http.StatusOK, quiz_pages.FinishedQuizzes())
}

func (qph *QuizPagesHandler) getArchivedQuizzes(c echo.Context) error {
	return utils.Render(c, http.StatusOK, quiz_pages.ArchivedQuizzes())
}

func (qph *QuizPagesHandler) getQuizProfile(c echo.Context) error {
	return utils.Render(c, http.StatusOK, quiz_pages.QuizProfile())
}

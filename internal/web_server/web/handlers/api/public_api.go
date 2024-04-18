package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/Molnes/Nyhetsjeger/internal/config"
	"github.com/Molnes/Nyhetsjeger/internal/models/quizzes"
	"github.com/Molnes/Nyhetsjeger/internal/models/users/user_quiz"
	"github.com/Molnes/Nyhetsjeger/internal/models/users/user_quiz_summary"
	utils "github.com/Molnes/Nyhetsjeger/internal/utils"
	"github.com/Molnes/Nyhetsjeger/internal/web_server/web/views/components/quiz_components/play_quiz_components"
	"github.com/Molnes/Nyhetsjeger/internal/web_server/web/views/pages/quiz_pages"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type publicApiHandler struct {
	sharedData *config.SharedData
}

// Creates a new PublicApiHandler
func NewPublicApiHandler(sharedData *config.SharedData) *publicApiHandler {
	return &publicApiHandler{sharedData}
}

// Registers the public api handlers to the given echo group
func (h *publicApiHandler) RegisterPublicApiHandlers(g *echo.Group) {
	g.POST("/user-answer", h.postAnswer)
	g.GET("/question", h.getQuestion)
	g.POST("/generate-summary", h.postGenerateSummary)
}

// Handles a post request with question answer from a guest user.
func (h *publicApiHandler) postAnswer(c echo.Context) error {
	questionID, err := uuid.Parse(c.QueryParam("question-id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Ugyldig eller manglende question-id")
	}
	pickedAnswerID, err := uuid.Parse(c.FormValue("answer-id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Ugyldig eller manglende answer-id i formdata")
	}

	questionPresentedAt, err := time.Parse(time.RFC3339, c.FormValue("last_question_presented_at"))
	if err != nil {
		return err
	}

	answered, err := user_quiz.AnswerQuestionGuest(h.sharedData.DB, questionID, pickedAnswerID, questionPresentedAt)
	if err != nil {
		return err
	}

	publicQuizId, err := user_quiz.GetOpenQuizId(h.sharedData.DB)
	if err != nil {
		return err
	}
	if publicQuizId != answered.Question.QuizID {
		return echo.NewHTTPError(http.StatusForbidden, "Kan ikke svare på spørsmål i uåpnede quizer uten å være innlogget.")
	}

	summaryRow := user_quiz_summary.AnsweredQuestion{
		QuestionID:            questionID,
		QuestionText:          answered.Question.Text,
		MaxPoints:             answered.Question.Points,
		ChosenAlternativeID:   answered.ChosenAnswerID,
		ChosenAlternativeText: answered.Question.GetAnswerTextById(answered.ChosenAnswerID),
		IsCorrect:             answered.Question.IsAnswerCorrect(answered.ChosenAnswerID),
		PointsAwarded:         answered.PointsAwarded,
	}

	return utils.Render(c, http.StatusOK, play_quiz_components.FeedbackButtonsWithClientState(answered, &summaryRow))

}

// Handles a get request for next question in a public (open) quiz.
func (h *publicApiHandler) getQuestion(c echo.Context) error {
	openQuizId, err := user_quiz.GetOpenQuizId(h.sharedData.DB)
	if err != nil {
		return err
	}

	quizId, err := uuid.Parse(c.QueryParam("quiz-id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Ugyldig eller manglende quiz-id")
	}
	if quizId != openQuizId {
		return echo.NewHTTPError(http.StatusNotFound, "Ingen åpen quiz med den angitte ID-en")
	}

	currentQuestion, err := strconv.ParseUint(c.QueryParam("current-question"), 10, 64)
	if err != nil || currentQuestion < 1 {
		return echo.NewHTTPError(http.StatusBadRequest, "Ugyldig eller manglende spørsmålsnummer")
	}

	totalPoints, err := strconv.ParseUint(c.FormValue("total-points"), 10, 64)
	if err != nil {
		totalPoints = 0
	}

	data, err := user_quiz.GetQuestionByNumberInQuiz(h.sharedData.DB, quizId, uint(currentQuestion))
	if err != nil {
		if err == sql.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, "Ingen spørsmål med det angitte nummeret")
		}
		return err
	}

	data.PointsGathered = uint(totalPoints)

	return utils.Render(c, http.StatusOK, play_quiz_components.QuizPlayContent(data))

}

// Handles a post request to generate a summary page out of data stored in the local storage.
// Expects a formdata with name "summaryRows", with value as stringified JSON array of user_quiz_summary.AnsweredQuestion.
func (h *publicApiHandler) postGenerateSummary(c echo.Context) error {
	openQuizId, err := user_quiz.GetOpenQuizId(h.sharedData.DB)
	if err != nil {
		return err
	}
	quizIdParam, err := uuid.Parse(c.QueryParam("quiz-id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Ugyldig eller manglende quiz-id")
	}
	if quizIdParam != openQuizId {
		return echo.NewHTTPError(http.StatusNotFound, "Ingen åpen quiz med den angitte ID-en")
	}
	quiz, err := quizzes.GetPartialQuizByID(h.sharedData.DB, openQuizId)
	if err != nil {
		return err
	}

	stringifiedSummaryRows := c.FormValue("summaryRows")
	var answeredQuestions []user_quiz_summary.AnsweredQuestion
	err = json.Unmarshal([]byte(stringifiedSummaryRows), &answeredQuestions)
	if err != nil {
		return err
	}

	summary := user_quiz_summary.UserQuizSummary{
		QuizID:            quiz.ID,
		QuizTitle:         quiz.Title,
		QuizActiveTo:      quiz.ActiveTo,
		MaxScore:          quiz.MaxScore,
		AchievedScore:     0,
		AnsweredQuestions: answeredQuestions,
	}
	summary.CalculateAchievedScoreFromAnswered()

	return utils.Render(c, http.StatusOK, quiz_pages.QuizSummaryContent(&summary))
}

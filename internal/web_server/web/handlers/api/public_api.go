package api

import (
	"net/http"
	"time"

	"github.com/Molnes/Nyhetsjeger/internal/config"
	"github.com/Molnes/Nyhetsjeger/internal/models/users/user_quiz"
	utils "github.com/Molnes/Nyhetsjeger/internal/utils"
	"github.com/Molnes/Nyhetsjeger/internal/web_server/web/views/components/quiz_components/play_quiz_components"
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
}

func (h *publicApiHandler) postAnswer(c echo.Context) error {
	questionID, err := uuid.Parse(c.QueryParam("question-id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid or missing question-id")
	}
	pickedAnswerID, err := uuid.Parse(c.FormValue("answer-id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid or missing answer-id in formdata")
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
		return echo.NewHTTPError(http.StatusForbidden, "Cannot answer question in non-open quiz without being authenticated.")
	}

	return utils.Render(c, http.StatusOK, play_quiz_components.FeedbackButtons(answered))

}

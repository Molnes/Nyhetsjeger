package api

import (
	"net/http"

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

	// TODO 
	answered, err := user_quiz.AnswerQuestion(h.sharedData.DB, uuid.UUID{}, questionID, pickedAnswerID)
	if err != nil {
		if err == user_quiz.ErrQuestionAlreadyAnswered {
			return echo.NewHTTPError(http.StatusConflict, "Question already answered")
		} else {
			return err
		}
	}

	return utils.Render(c, http.StatusOK, play_quiz_components.FeedbackButtons(answered, true))

}

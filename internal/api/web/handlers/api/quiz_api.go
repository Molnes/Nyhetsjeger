package api

import (
	"net/http"

	"github.com/Molnes/Nyhetsjeger/internal/api/web/views/components/quiz_components"
	"github.com/Molnes/Nyhetsjeger/internal/data/users"
	"github.com/Molnes/Nyhetsjeger/internal/utils"
	"github.com/labstack/echo/v4"
)

func RegisterQuizApiHandlers(e *echo.Group) {
	e.GET("/question", GetQuestion)
	e.GET("/:quizId/summary", GetQuizSummary)

}

func GetQuestion(c echo.Context) error {
	question := "Exampl question?"
	alts := []string{"alt1", "alt2", "alt3", "alt4"}
	return utils.Render(c, http.StatusOK, quiz_components.Question(question, alts))
}

func GetQuizSummary(c echo.Context) error {
	return utils.Render(c, http.StatusOK, quiz_components.QuizSummary(users.SampleUserQuizSummary))
}

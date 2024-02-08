package api

import (
	"net/http"
	"github.com/Molnes/Nyhetsjeger/internal/api/web/views/components/quiz_components"
	"github.com/Molnes/Nyhetsjeger/internal/utils"
	"github.com/labstack/echo/v4"
)

func RegisterQuizApiHandlers(e *echo.Group) {
	e.GET("/question", GetQuestion)

}

func GetQuestion(c echo.Context) error {
	question := "Exampl question?"
	alts := []string{"alt1", "alt2", "alt3", "alt4"}
	return utils.Render(c, http.StatusOK, quiz_components.Question(question, alts))
}

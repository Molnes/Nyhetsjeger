package handlers

import (
	"net/http"

	"github.com/Molnes/Nyhetsjeger/internal/api/web/views"
	"github.com/Molnes/Nyhetsjeger/internal/utils"
	"github.com/labstack/echo/v4"
)

// func(c echo.Context) error {
// 	return utils.Render(c, http.StatusOK, views.Index())
// })

func RegisterQuizHandlers(e *echo.Echo) {
	e.GET("/", quizHomePage)
}

func quizHomePage(c echo.Context) error {
	return utils.Render(c, http.StatusOK, views.QuizHomePage())
}

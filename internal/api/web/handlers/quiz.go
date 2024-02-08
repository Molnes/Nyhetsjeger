package handlers

import (
	"net/http"

	"github.com/Molnes/Nyhetsjeger/internal/api/web/views/pages"
	"github.com/Molnes/Nyhetsjeger/internal/utils"
	"github.com/labstack/echo/v4"
)

// Registers handlers for quiz related pages
func RegisterQuizHandlers(e *echo.Echo) {
	e.GET("/", quizHomePage)
}

// Renders the quiz home page
func quizHomePage(c echo.Context) error {
	return utils.Render(c, http.StatusOK, pages.QuizHomePage())
}

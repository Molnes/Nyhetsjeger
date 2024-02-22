package handlers

import (
	"net/http"

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
	quizzList, err := quizzes.GetQuizzes(qph.sharedData.DatabaseConn)
	if err != nil {
		return err
	}
	return utils.Render(c, http.StatusOK, quiz_pages.QuizHomePage(
		quizzList,
	))
}

func (qph *QuizPagesHandler) GetQuizPage(c echo.Context) error {
	return utils.Render(c, http.StatusOK, quiz_pages.QuizPage())
}

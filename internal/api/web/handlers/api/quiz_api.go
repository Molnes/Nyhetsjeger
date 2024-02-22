package api

import (
	"net/http"

	"github.com/Molnes/Nyhetsjeger/internal/api/web/views/components/quiz_components"
	"github.com/Molnes/Nyhetsjeger/internal/config"
	"github.com/Molnes/Nyhetsjeger/internal/data/articles"
	"github.com/Molnes/Nyhetsjeger/internal/data/users"
	"github.com/Molnes/Nyhetsjeger/internal/utils"
	"github.com/labstack/echo/v4"
)

type QuizApiHandler struct {
	sharedData *config.SharedData
}

func NewQuizApiHandler(sharedData *config.SharedData) *QuizApiHandler {
	return &QuizApiHandler{sharedData}
}

func (qah *QuizApiHandler) RegisterQuizApiHandlers(e *echo.Group) {
	e.GET("/question", qah.GetQuestion)
	e.GET("/article", qah.GetArticle)
	e.GET("/articles", qah.GetArticles)
	e.GET("/:quizId/summary", qah.GetQuizSummary)
}

func (qah *QuizApiHandler) GetQuestion(c echo.Context) error {
	question := "Exampl question?"
	alts := []string{"alt1", "alt2", "alt3", "alt4"}
	return utils.Render(c, http.StatusOK, quiz_components.Question(question, alts))
}

func (qah *QuizApiHandler) GetQuizSummary(c echo.Context) error {
	return utils.Render(c, http.StatusOK, quiz_components.QuizSummary(users.SampleUserQuizSummary))
}

func (qah *QuizApiHandler) GetArticle(c echo.Context) error {
	article := articles.SampleArticles[0]
	return utils.Render(c, http.StatusOK, quiz_components.ArticleCard(article))
}

func (qah *QuizApiHandler) GetArticles(c echo.Context) error {
	articles := articles.SampleArticles
	return utils.Render(c, http.StatusOK, quiz_components.ArticleList(articles))
}

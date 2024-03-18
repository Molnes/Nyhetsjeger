package api

import (
	"net/http"

	"github.com/Molnes/Nyhetsjeger/internal/config"
	"github.com/Molnes/Nyhetsjeger/internal/models/articles"
	"github.com/Molnes/Nyhetsjeger/internal/utils"
	"github.com/Molnes/Nyhetsjeger/internal/web_server/web/views/components/quiz_components"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type QuizApiHandler struct {
	sharedData *config.SharedData
}

func NewQuizApiHandler(sharedData *config.SharedData) *QuizApiHandler {
	return &QuizApiHandler{sharedData}
}

func (qah *QuizApiHandler) RegisterQuizApiHandlers(e *echo.Group) {
	e.GET("/question", qah.getQuestion)
	e.GET("/article", qah.getArticle)
	e.GET("/articles", qah.getArticles)
	e.POST("/user-answer", qah.postUserAnswer)
}

func (qah *QuizApiHandler) getQuestion(c echo.Context) error {
	question := "Exampl question?"
	alts := []string{"alt1", "alt2", "alt3", "alt4"}
	return utils.Render(c, http.StatusOK, quiz_components.Question(question, alts))
}

func (qah *QuizApiHandler) getArticle(c echo.Context) error {
	article := articles.SampleArticles[0]
	return utils.Render(c, http.StatusOK, quiz_components.ArticleCard(&article))
}

func (qah *QuizApiHandler) getArticles(c echo.Context) error {
	articles := articles.SampleArticles
	return utils.Render(c, http.StatusOK, quiz_components.ArticleList(&articles))
}

func (qah *QuizApiHandler) postUserAnswer(c echo.Context) error {
	_, err := uuid.Parse(c.QueryParam("questionid"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid or missing questionid")
	}
	_, err = uuid.Parse(c.FormValue("answer"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid or missing answerid in formdata")
	}

	return echo.NewHTTPError(http.StatusNotImplemented, "Not implemented")

}

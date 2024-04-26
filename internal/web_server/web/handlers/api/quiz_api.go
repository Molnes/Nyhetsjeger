package api

import (
	"fmt"
	"net/http"

	"github.com/Molnes/Nyhetsjeger/internal/config"
	"github.com/Molnes/Nyhetsjeger/internal/models/articles"
	"github.com/Molnes/Nyhetsjeger/internal/models/sessions"
	"github.com/Molnes/Nyhetsjeger/internal/models/users"
	"github.com/Molnes/Nyhetsjeger/internal/models/users/user_quiz"
	"github.com/Molnes/Nyhetsjeger/internal/utils"
	"github.com/Molnes/Nyhetsjeger/internal/web_server/web/views/components/profile_components"
	"github.com/Molnes/Nyhetsjeger/internal/web_server/web/views/components/quiz_components"
	"github.com/Molnes/Nyhetsjeger/internal/web_server/web/views/components/quiz_components/play_quiz_components"
	"github.com/Molnes/Nyhetsjeger/internal/web_server/web/views/components/user_management"

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
	e.GET("/articles", qah.getArticles)
	e.GET("/next-question", qah.getNextQuestion)
	e.POST("/user-answer", qah.postUserAnswer)
	e.PATCH("/brukernavn", qah.patchRandomUsername)
	e.DELETE("/profil", qah.deleteProfile)
	e.POST("/accept-terms", qah.postAcceptTerms)
	e.POST("/participation", qah.postParticipation)
}

func (qah *QuizApiHandler) postParticipation(c echo.Context) error {
	userID := utils.GetUserIDFromCtx(c)
	currentParticipationStauts, err := users.GetParticipationStatus(qah.sharedData.DB, userID)
	if err != nil {
		return err
	}

	err = users.SetParticipationStatus(qah.sharedData.DB, userID, !currentParticipationStauts)
	if err != nil {
		return err
	}
	//returning the new participation status
	return utils.Render(c, http.StatusOK, profile_components.Participation_chackbox(!currentParticipationStauts))
}

func (qah *QuizApiHandler) getArticles(c echo.Context) error {
	quizId, err := uuid.Parse(c.QueryParam("quiz-id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid or missing quiz-id")
	}

	articles, err := articles.GetUsedArticlesByQuizID(qah.sharedData.DB, quizId)
	if err != nil {
		return err
	}
	return utils.Render(c, http.StatusOK, quiz_components.ArticleList(articles))
}

// Handles get request for the next question in a given quiz.
// Renders the QuizPlayContent component with the quiz data.
func (qah *QuizApiHandler) getNextQuestion(c echo.Context) error {
	quizID, err := uuid.Parse(c.QueryParam("quiz-id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid or missing quiz-id")
	}

	quizData, err := user_quiz.NextQuestionInQuiz(qah.sharedData.DB, utils.GetUserIDFromCtx(c), quizID)
	if err != nil {
		if err == user_quiz.ErrNoSuchQuiz {
			return echo.NewHTTPError(http.StatusNotFound, "No such quiz")
		} else if err == user_quiz.ErrNoMoreQuestions {
			return echo.NewHTTPError(http.StatusNotFound, "No more questions")
		} else {
			return err
		}
	}

	return utils.Render(c, http.StatusOK, play_quiz_components.QuizPlayContent(quizData))
}

// Handles post request for a user's answer to a question.
//
// Expects the question-id as a query parameter and the answer-id as form data.
//
// Renders the FeedbackButtons component with the answer feedback.
func (qah *QuizApiHandler) postUserAnswer(c echo.Context) error {
	questionID, err := uuid.Parse(c.QueryParam("question-id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid or missing question-id")
	}
	pickedAnswerID, err := uuid.Parse(c.FormValue("answer-id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid or missing answer-id in formdata")
	}

	answered, err := user_quiz.AnswerQuestion(qah.sharedData.DB, utils.GetUserIDFromCtx(c), questionID, pickedAnswerID)
	if err != nil {
		if err == user_quiz.ErrQuestionAlreadyAnswered {
			return echo.NewHTTPError(http.StatusConflict, "Question already answered")
		} else {
			return err
		}
	}

	return utils.Render(c, http.StatusOK, play_quiz_components.FeedbackButtons(answered))
}

// Handles patch request for a random username.
func (qah *QuizApiHandler) patchRandomUsername(c echo.Context) error {
	username, err := users.AssignUsernameToUser(qah.sharedData.DB, utils.GetUserIDFromCtx(c), c.Request().Context())
	if err != nil {
		return err
	}

	return utils.Render(c, http.StatusOK, user_management.UsernameInput("usrn-in", username))
}

// Deletes the user from the database and logs the user out
func (qah *QuizApiHandler) deleteProfile(c echo.Context) error {
	//TODO: Avoid duplicate logout code. Have agreed to look at it later.
	err := users.DeleteUserByID(qah.sharedData.DB, utils.GetUserIDFromCtx(c))
	if err != nil {
		return err
	}

	session, err := qah.sharedData.SessionStore.Get(c.Request(), sessions.SESSION_NAME)
	if err != nil {
		return fmt.Errorf("failed to get session: %s", err.Error())
	}
	session.Options.MaxAge = -1
	err = session.Save(c.Request(), c.Response())
	if err != nil {
		return fmt.Errorf("failed to save session: %s", err.Error())
	}
	c.Response().Header().Set("HX-Redirect", "/")
	return c.NoContent(http.StatusNoContent)
}

// Handles a post request with expected accepted-terms=on form data. Sets te caller's accepted terms of service value.
func (h *QuizApiHandler) postAcceptTerms(c echo.Context) error {
	isAccepted := c.FormValue("accepted-terms") == "on"
	if !isAccepted {
		return echo.NewHTTPError(http.StatusBadRequest, "Terms of service must be accepted.")
	}
	err := users.UpdateAcceptedTermsByUserID(h.sharedData.DB, utils.GetUserIDFromCtx(c), true)
	if err != nil {
		return err
	}
	c.Response().Header().Set("HX-Redirect", "/")
	return c.NoContent(http.StatusNoContent)
}

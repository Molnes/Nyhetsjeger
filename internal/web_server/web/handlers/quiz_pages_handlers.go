package handlers

import (
	"fmt"
	"net/http"
	"regexp"

	"github.com/Molnes/Nyhetsjeger/internal/config"
	"github.com/Molnes/Nyhetsjeger/internal/models/quizzes"
	"github.com/Molnes/Nyhetsjeger/internal/models/sessions"
	"github.com/Molnes/Nyhetsjeger/internal/models/users"
	"github.com/Molnes/Nyhetsjeger/internal/models/users/user_quiz"
	"github.com/Molnes/Nyhetsjeger/internal/models/users/user_quiz_summary"
	"github.com/Molnes/Nyhetsjeger/internal/models/users/user_ranking"
	"github.com/Molnes/Nyhetsjeger/internal/utils"
	"github.com/Molnes/Nyhetsjeger/internal/web_server/web/views/pages/quiz_pages"
	"github.com/google/uuid"
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
	e.GET("/summary", qph.getQuizSummary)
	e.GET("/play", qph.getPlayQuizPage)
	e.GET("/toppliste", qph.getScoreboard)
	e.GET("/fullforte", qph.getFinishedQuizzes)
	e.GET("/profil", qph.getQuizProfile)

	e.GET("/brukernavn", qph.usernamePage)
	e.POST("/brukernavn", qph.postUsername)

	e.GET("/profil", qph.getProfile)
	e.DELETE("/profil", qph.deleteProfile)
	e.POST("/profil", qph.postUsername)
}

// Renders the quiz home page
func (qph *QuizPagesHandler) quizHomePage(c echo.Context) error {
	quizzList, err := quizzes.GetUnfinishedQuizzesByUserID(qph.sharedData.DB, utils.GetUserIDFromCtx(c))
	if err != nil {
		return err
	}

	userRankingInfo, err := user_ranking.GetUserRanking(qph.sharedData.DB, utils.GetUserIDFromCtx(c))
	if err != nil {
		return err
	}
	return utils.Render(c, http.StatusOK, quiz_pages.QuizHomePage(
		quizzList,
        userRankingInfo,
	))
}

// Renders the play quiz page, expects a query parameter quiz-id.
// If the quiz is not found, returns a 404 error.
// If the quiz is completed, redirects to the quiz summary page.
func (qph *QuizPagesHandler) getPlayQuizPage(c echo.Context) error {
	quizID, err := uuid.Parse(c.QueryParam("quiz-id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid or missing quiz-id")
	}

	startQuizData, err := user_quiz.NextQuestionInQuiz(qph.sharedData.DB, utils.GetUserIDFromCtx(c), quizID)
	if err != nil {
		if err == user_quiz.ErrNoSuchQuiz {
			return echo.NewHTTPError(http.StatusNotFound, "No such quiz")
		} else if err == user_quiz.ErrNoMoreQuestions {
			return c.Redirect(http.StatusTemporaryRedirect, "/quiz/summary?quiz-id="+quizID.String())
		} else {
			return err
		}
	}

	return utils.Render(c, http.StatusOK, quiz_pages.QuizPlayPage(startQuizData.PartialQuiz.Title, startQuizData))
}

// Renders the quiz summary page, expects a query parameter quiz-id.
// If the quiz is not found, returns a 404 error.
// If the quiz is not completed, returns a 409 (conflict) error, since the summary is not available at that point.
func (qph *QuizPagesHandler) getQuizSummary(c echo.Context) error {
	quizID, err := uuid.Parse(c.QueryParam("quiz-id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid or missing quiz-id")
	}

	quizSummary, err := user_quiz_summary.GetQuizSummary(qph.sharedData.DB, utils.GetUserIDFromCtx(c), quizID)
	if err != nil {
		if err == user_quiz_summary.ErrNoSuchQuiz {
			return echo.NewHTTPError(http.StatusNotFound, "No such quiz")
		}
		if err == user_quiz_summary.ErrQuizNotCompleted {
			return echo.NewHTTPError(http.StatusConflict, "Quiz not completed - no summary available")
		}
	}

	return utils.Render(c, http.StatusOK, quiz_pages.QuizSummaryPage(quizSummary))

}

// Renders the scoreboard page.
func (qph *QuizPagesHandler) getScoreboard(c echo.Context) error {
	rankings, err := user_ranking.GetRanking(qph.sharedData.DB)
	if err != nil {
		return err
	}
	return utils.Render(c, http.StatusOK, quiz_pages.Scoreboard(rankings))
}

func (qph *QuizPagesHandler) getFinishedQuizzes(c echo.Context) error {
	return utils.Render(c, http.StatusOK, quiz_pages.FinishedQuizzes())
}

func (qph *QuizPagesHandler) getQuizProfile(c echo.Context) error {
	return utils.Render(c, http.StatusOK, quiz_pages.QuizProfile())
}

func (qph *QuizPagesHandler) usernamePage(c echo.Context) error {

	user, error := users.GetUserByID(qph.sharedData.DB, utils.GetUserIDFromCtx(c))
	if error != nil {
		return error
	}

	return utils.Render(c, http.StatusOK, quiz_pages.UsernamePage(user))
}

// Adds phone number and the leaderboards opt-in status to the user
func (qph *QuizPagesHandler) postUsername(c echo.Context) error {
	phonenumber := c.FormValue("phonenumber")
	match, _ := regexp.MatchString(`^(\d{2} \d{2} \d{2} \d{2}|\d{3} \d{2} \d{3}|\d{8})$`, phonenumber)
	if !match {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid phone number")
	}
	enterCompetion := c.FormValue("competition") == "on"

	err := users.AssignPhonenumberToUser(qph.sharedData.DB, utils.GetUserIDFromCtx(c), phonenumber)
	if err != nil {
		return err
	}
	err = users.AssignOptInRankingToUser(qph.sharedData.DB, utils.GetUserIDFromCtx(c), enterCompetion)
	if err != nil {
		return err
	}

	return c.Redirect(http.StatusFound, "/quiz")
}

func (qph *QuizPagesHandler) getProfile(c echo.Context) error {

	user, error := users.GetUserByID(qph.sharedData.DB, utils.GetUserIDFromCtx(c))
	if error != nil {
		return error
	}

	return utils.Render(c, http.StatusOK, quiz_pages.UserProfile(user))
}

// Deletes the user from the database and logs the user out
func (qph *QuizPagesHandler) deleteProfile(c echo.Context) error {
	//TODO: Avoid duplicate logout code. Have agreed to look at it later.
	err := users.DeleteUserByID(qph.sharedData.DB, utils.GetUserIDFromCtx(c))
	if err != nil {
		return err
	}

	session, err := qph.sharedData.SessionStore.Get(c.Request(), sessions.SESSION_NAME)
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

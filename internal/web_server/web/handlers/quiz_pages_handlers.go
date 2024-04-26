package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/Molnes/Nyhetsjeger/internal/config"
	"github.com/Molnes/Nyhetsjeger/internal/models/quizzes"
	"github.com/Molnes/Nyhetsjeger/internal/models/users"
	"github.com/Molnes/Nyhetsjeger/internal/models/users/user_quiz"
	"github.com/Molnes/Nyhetsjeger/internal/models/users/user_quiz_summary"
	"github.com/Molnes/Nyhetsjeger/internal/models/users/user_ranking"
	"github.com/Molnes/Nyhetsjeger/internal/utils"
	"github.com/Molnes/Nyhetsjeger/internal/web_server/web/views/pages/quiz_pages"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	data_converting "github.com/Molnes/Nyhetsjeger/internal/utils/converting"
)

const (
	errInvalidOrMissingQuizID    = "Manglende eller ugyldig quiz-id"
	errNoSuchQuiz                = "Quizen ble ikke funnet"
	errQuizNotCompletedNoSummary = "Quizen er ikke fullført, oppsummering er ikke tilgjengelig"
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
	e.GET("/accept-terms", qph.getAcceptTermsPage)

	e.GET("/brukernavn", qph.usernamePage)

	e.GET("/profil", qph.getProfile)
}

// Renders the quiz home page
func (qph *QuizPagesHandler) quizHomePage(c echo.Context) error {
	quizList, err := quizzes.GetQuizzesByUserIDAndFinishedOrNot(qph.sharedData.DB, utils.GetUserIDFromCtx(c), false)
	if err != nil {
		return err
	}

	partialQuizList, err := data_converting.ConvertQuizzesToPartial(quizList, qph.sharedData.DB)
	if err != nil {
		return err

	}

	oldQuizzes, err := quizzes.GetQuizzesByUserIDAndFinishedOrNotAndNotActive(qph.sharedData.DB, utils.GetUserIDFromCtx(c), false)
	if err != nil {
		return err
	}

	oldPartialQuizList, err := data_converting.ConvertQuizzesToPartial(oldQuizzes, qph.sharedData.DB)
	if err != nil {
		return err
	}

	userRankingInfo := user_ranking.UserRanking{}

	userRankingInfo, err = user_ranking.GetUserRanking(qph.sharedData.DB, utils.GetUserIDFromCtx(c), time.Now().Month(), time.Now().Year(), time.Local, user_ranking.Month)
	if err != nil {
		if err == sql.ErrNoRows {
			user, err := users.GetUserByID(qph.sharedData.DB, utils.GetUserIDFromCtx(c))
			if err != nil {
				return err
			}
			userRankingInfo = user_ranking.UserRanking{
				Username:  user.Username,
				Points:    0,
				Placement: 0,
			}
		} else {

			return err
		}
	}
	return utils.Render(c, http.StatusOK, quiz_pages.QuizHomePage(
		partialQuizList,
		oldPartialQuizList,
		userRankingInfo,
	))
}

// Renders the play quiz page, expects a query parameter quiz-id.
// If the quiz is not found, returns a 404 error.
// If the quiz is completed, redirects to the quiz summary page.
func (qph *QuizPagesHandler) getPlayQuizPage(c echo.Context) error {
	quizID, err := uuid.Parse(c.QueryParam("quiz-id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errInvalidOrMissingQuizID)
	}

	startQuizData, err := user_quiz.NextQuestionInQuiz(qph.sharedData.DB, utils.GetUserIDFromCtx(c), quizID)
	if err != nil {
		if err == user_quiz.ErrNoSuchQuiz {
			return echo.NewHTTPError(http.StatusNotFound, errNoSuchQuiz)
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
		return echo.NewHTTPError(http.StatusBadRequest, errInvalidOrMissingQuizID)
	}

	quizSummary, err := user_quiz_summary.GetQuizSummary(qph.sharedData.DB, utils.GetUserIDFromCtx(c), quizID)
	if err != nil {
		if err == user_quiz_summary.ErrNoSuchQuiz {
			return echo.NewHTTPError(http.StatusNotFound, errNoSuchQuiz)
		}
		if err == user_quiz_summary.ErrQuizNotCompleted {
			return echo.NewHTTPError(http.StatusConflict, errQuizNotCompletedNoSummary)
		}
	}

	return utils.Render(c, http.StatusOK, quiz_pages.QuizSummaryPage(quizSummary))

}

// Renders the scoreboard page.
func (qph *QuizPagesHandler) getScoreboard(c echo.Context) error {
	rankings, err := user_ranking.GetRanking(qph.sharedData.DB, time.Now().Month(), time.Now().Year(), time.Local, user_ranking.Month)
	if err != nil {
		return err
	}

	userRankingInfo, err := user_ranking.GetUserRanking(qph.sharedData.DB, utils.GetUserIDFromCtx(c), time.Now().Month(), time.Now().Year(), time.Local, user_ranking.Month)

	return utils.Render(c, http.StatusOK, quiz_pages.Scoreboard(rankings, userRankingInfo))
}

func (qph *QuizPagesHandler) getFinishedQuizzes(c echo.Context) error {
	quizList, err := quizzes.GetQuizzesByUserIDAndFinishedOrNot(qph.sharedData.DB, utils.GetUserIDFromCtx(c), true)
	if err != nil {
		return err
	}
	// Convert to partial quizzes
	partialQuizzes, err := data_converting.ConvertQuizzesToPartial(quizList, qph.sharedData.DB)

	if err != nil {
		return err
	}

	return utils.Render(c, http.StatusOK, quiz_pages.FinishedQuizzes(partialQuizzes))
}

func (qph *QuizPagesHandler) usernamePage(c echo.Context) error {

	user, error := users.GetUserByID(qph.sharedData.DB, utils.GetUserIDFromCtx(c))
	if error != nil {
		return error
	}

	return utils.Render(c, http.StatusOK, quiz_pages.UsernamePage(user))
}

func (qph *QuizPagesHandler) getProfile(c echo.Context) error {

	user, error := users.GetUserByID(qph.sharedData.DB, utils.GetUserIDFromCtx(c))
	if error != nil {
		return error
	}

	return utils.Render(c, http.StatusOK, quiz_pages.UserProfile(user))
}

func (gph *QuizPagesHandler) getAcceptTermsPage(c echo.Context) error {
	user, err := users.GetUserByID(gph.sharedData.DB, utils.GetUserIDFromCtx(c))
	if err != nil {
		return err
	}
	if user.AcceptedTerms {
		return c.Redirect(http.StatusTemporaryRedirect, "/quiz")
	}

	return utils.Render(c, http.StatusOK, quiz_pages.AcceptTermsPage())
}

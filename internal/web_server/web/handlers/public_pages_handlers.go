package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/Molnes/Nyhetsjeger/internal/config"
	"github.com/Molnes/Nyhetsjeger/internal/models/quizzes"
	"github.com/Molnes/Nyhetsjeger/internal/models/sessions"
	"github.com/Molnes/Nyhetsjeger/internal/models/users"
	"github.com/Molnes/Nyhetsjeger/internal/models/users/user_quiz"
	"github.com/Molnes/Nyhetsjeger/internal/models/users/user_roles"
	"github.com/Molnes/Nyhetsjeger/internal/utils"
	"github.com/Molnes/Nyhetsjeger/internal/web_server/web/views/pages/public_pages"
	"github.com/Molnes/Nyhetsjeger/internal/web_server/web/views/pages/quiz_pages"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type PublicPagesHandler struct {
	sharedData *config.SharedData
}

// Creates a new PublicPagesHandler
func NewPublicPagesHandler(sharedData *config.SharedData) *PublicPagesHandler {
	return &PublicPagesHandler{sharedData}
}

// Registers handlers for public pages
func (pph *PublicPagesHandler) RegisterPublicPages(e *echo.Echo) {
	e.GET("", pph.homePage)
	e.GET("/login", pph.loginPage)
	e.GET("/betingelser", pph.termsPage)
	e.GET("/gjest", pph.getGuestHomePage)
	e.GET("/gjest-quiz", pph.getGuestQuiz)
}

// Handles get request to get home page. If user is authenticated, they get redirected to /quiz or /dashboard
func (pph *PublicPagesHandler) homePage(c echo.Context) error {
	session, err := pph.sharedData.SessionStore.Get(c.Request(), sessions.SESSION_NAME)
	if err == nil {
		userData := session.Values[sessions.USER_DATA_VALUE]
		if userData != nil {
			sessiondata, ok := userData.(users.UserSessionData)
			if !ok {
				return fmt.Errorf("AuthenticationOnHomePage: failed to cast user data to UserSessionData")
			}
			role, err := users.GetUserRole(pph.sharedData.DB, sessiondata.ID)
			if err != nil {
				return err
			}
			switch role {
			case user_roles.OrganizationAdmin, user_roles.QuizAdmin:
				return c.Redirect(http.StatusTemporaryRedirect, "/dashboard")
			default:
				return c.Redirect(http.StatusTemporaryRedirect, "/quiz")
			}
		}
	}
	return utils.Render(c, http.StatusOK, public_pages.HomePage())
}

// Handles get request to the login page. If user is authenticated, they get redirected.
func (pph *PublicPagesHandler) loginPage(c echo.Context) error {
	session, err := pph.sharedData.SessionStore.Get(c.Request(), sessions.SESSION_NAME)
	if err == nil && session.Values[sessions.USER_DATA_VALUE] != nil {
		return c.Redirect(http.StatusTemporaryRedirect, "/quiz")
	}
	return utils.Render(c, http.StatusOK, public_pages.LoginPage())
}

// Handles get request to the terms of service page.
func (pph *PublicPagesHandler) termsPage(c echo.Context) error {
	return utils.Render(c, http.StatusOK, public_pages.TermsOfServicePage())
}

// Handles get request to the guest home page.
func (pph *PublicPagesHandler) getGuestHomePage(c echo.Context) error {
	quizId, err := user_quiz.GetOpenQuizId(pph.sharedData.DB)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	var partialQuiz quizzes.PartialQuiz
	if quizId != uuid.Nil {
		selectedQuiz, err := quizzes.GetPartialQuizByID(pph.sharedData.DB, quizId)
		if err != nil {
			return err
		}
		partialQuiz = *selectedQuiz
	}

	return utils.Render(c, http.StatusOK, public_pages.GuestHomePage(&partialQuiz))
}

const quizIdQueryParam = "quiz-id"
const currentQuestionQueryParam = "current-question"
const totalPointsQueryParam = "total-points"

// Handles get request to the guest play quiz page. If no parameters provided, an open quiz is found and user is redirected there.
func (h *PublicPagesHandler) getGuestQuiz(c echo.Context) error {
	openQuizId, err := user_quiz.GetOpenQuizId(h.sharedData.DB)
	if err != nil {
		if err == sql.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, "Ingen åpen quiz")
		}
		return err
	}

	quizIdParam := c.QueryParam(quizIdQueryParam)
	currentQuestionParam := c.QueryParam(currentQuestionQueryParam)

	if quizIdParam == "" || currentQuestionParam == "" {
		c.Redirect(http.StatusTemporaryRedirect,
			fmt.Sprintf("/gjest-quiz?%s=%s&%s=%s", quizIdQueryParam, openQuizId.String(), currentQuestionQueryParam, "1"),
		)
	}

	quizId, err := uuid.Parse(quizIdParam)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Ugyldig quiz-id")
	}
	if quizId != openQuizId {
		return echo.NewHTTPError(http.StatusNotFound, "Ingen åpen quiz med den angitte ID-en")
	}

	currentQuestion, err := strconv.ParseUint(currentQuestionParam, 10, 64)
	if err != nil || currentQuestion < 1 {
		return echo.NewHTTPError(http.StatusBadRequest, "Ugyldig spørsmålsnummer")
	}

	totalPoints, err := strconv.ParseUint(c.FormValue(totalPointsQueryParam), 10, 64)
	if err != nil {
		totalPoints = 0
	}

	data, err := user_quiz.GetQuestionByNumberInQuiz(h.sharedData.DB, quizId, uint(currentQuestion))
	if err != nil {
		if err == sql.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, "Ingen spørsmål med det angitte nummeret")
		}
		return err
	}
	data.PointsGathered = uint(totalPoints)

	return utils.Render(c, http.StatusOK, quiz_pages.QuizPlayPage(data.PartialQuiz.Title, data))
}

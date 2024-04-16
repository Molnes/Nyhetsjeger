package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/Molnes/Nyhetsjeger/internal/config"
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
	e.GET("terms-of-service", pph.termsPage)
	e.GET("/guest-quiz", pph.getGuestQuiz)
}

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

func (pph *PublicPagesHandler) loginPage(c echo.Context) error {
	session, err := pph.sharedData.SessionStore.Get(c.Request(), sessions.SESSION_NAME)
	if err == nil && session.Values["user"] != nil {
		return c.Redirect(http.StatusTemporaryRedirect, "/quiz")
	}
	return utils.Render(c, http.StatusOK, public_pages.LoginPage())
}

func (pph *PublicPagesHandler) termsPage(c echo.Context) error {
	return utils.Render(c, http.StatusOK, public_pages.TermsOfServicePage())
}

const quizIdQueryParam = "quiz-id"
const currentQuestionQueryParam = "current-question"
const totalPointsQueryParm = "total-points"

func (h *PublicPagesHandler) getGuestQuiz(c echo.Context) error {
	openQuizId, err := user_quiz.GetOpenQuizId(h.sharedData.DB)
	if err != nil {
		return err
	}

	quizIdParam := c.QueryParam(quizIdQueryParam)
	currentQuestionParam := c.QueryParam(currentQuestionQueryParam)

	if quizIdParam == "" || currentQuestionParam == "" {
		c.Redirect(http.StatusTemporaryRedirect,
			fmt.Sprintf("/guest-quiz?%s=%s&%s=%s", quizIdQueryParam, openQuizId.String(), currentQuestionQueryParam, "1"),
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
		return echo.NewHTTPError(http.StatusBadRequest, "Ugyldig såørsmål nummer")
	}

	totalPoints, err := strconv.ParseUint(c.FormValue(totalPointsQueryParm), 10, 64)
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

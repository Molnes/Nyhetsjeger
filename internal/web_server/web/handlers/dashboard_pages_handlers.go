package handlers

import (
	"database/sql"
	"net/http"
	"net/url"

	"github.com/Molnes/Nyhetsjeger/internal/config"
	"github.com/Molnes/Nyhetsjeger/internal/models/articles"
	"github.com/Molnes/Nyhetsjeger/internal/models/questions"
	"github.com/Molnes/Nyhetsjeger/internal/models/quizzes"
	"github.com/Molnes/Nyhetsjeger/internal/models/users"
	"github.com/Molnes/Nyhetsjeger/internal/models/users/access_control"
	"github.com/Molnes/Nyhetsjeger/internal/models/users/user_roles"
	"github.com/Molnes/Nyhetsjeger/internal/utils"
	"github.com/Molnes/Nyhetsjeger/internal/web_server/middlewares"
	dashboard_components "github.com/Molnes/Nyhetsjeger/internal/web_server/web/views/components/dashboard_components/edit_quiz"
	"github.com/Molnes/Nyhetsjeger/internal/web_server/web/views/components/dashboard_components/side_menu"
	"github.com/Molnes/Nyhetsjeger/internal/web_server/web/views/pages/dashboard_pages"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type DashboardPagesHandler struct {
	sharedData *config.SharedData
}

// Creates a new DashboardPagesHandler.
func NewDashboardPagesHandler(sharedData *config.SharedData) *DashboardPagesHandler {
	return &DashboardPagesHandler{sharedData}
}

// Registers handlers for dashboard related pages.
func (dph *DashboardPagesHandler) RegisterDashboardHandlers(g *echo.Group) {
	g.GET("", dph.dashboardHomePage)
	g.GET("/edit-quiz", dph.dashboardEditQuiz)
	g.GET("/edit-quiz/new-question", dph.dashboardNewQuestionModal)
	g.GET("/edit-question", dph.dashboardEditQuestionModal)
	g.GET("/leaderboard", dph.leaderboard)
	g.GET("/user-details", dph.userDetails)
	g.GET("/user-admin", dph.getUserAdministration)

	mw := middlewares.NewAuthorizationMiddleware(dph.sharedData, []user_roles.Role{user_roles.OrganizationAdmin})
	organizationAdminGroup := g.Group("", mw.EnforceRole)
	organizationAdminGroup.GET("/access-settings", dph.accessSettings)
}

// Renders the dashboard home page.
func (dph *DashboardPagesHandler) dashboardHomePage(c echo.Context) error {
	addMenuContext(c, side_menu.Home)

	nonPublishedQuizzes, err := quizzes.GetQuizzesByPublishStatus(dph.sharedData.DB, false)
	if err != nil {
		return err
	}
	publishedQuizzes, err := quizzes.GetQuizzesByPublishStatus(dph.sharedData.DB, true)
	if err != nil {
		return err
	}

	return utils.Render(c, http.StatusOK, dashboard_pages.DashboardHomePage(nonPublishedQuizzes, publishedQuizzes))
}

// Renders the page for editing quiz.
func (dph *DashboardPagesHandler) dashboardEditQuiz(c echo.Context) error {
	// Get the quiz ID.
	uuid_id, err := uuid.Parse(c.QueryParam("quiz-id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid or missing quiz id")
	}

	// Get the quiz by ID.
	quiz, err := quizzes.GetQuizByID(dph.sharedData.DB, uuid_id)
	if err != nil {
		if err == sql.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, "No quiz with given id found.")
		} else {
			return err
		}
	}

	// Get all the articles for the quiz by quiz ID.
	articles, _ := articles.GetArticlesByQuizID(dph.sharedData.DB, uuid_id)

	// Get all the questions for the quiz by quiz ID.
	questions, _ := questions.GetQuestionsByQuizID(dph.sharedData.DB, &uuid_id)

	return utils.Render(c, http.StatusOK, dashboard_pages.EditQuiz(quiz, articles, questions))
}

// Renders the modal for creating a new question.
func (dph *DashboardPagesHandler) dashboardNewQuestionModal(c echo.Context) error {
	// Get the quiz ID.
	quiz_id, err := uuid.Parse(c.QueryParam("quiz-id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid or missing quiz id")
	}

	// Create a new question with no actual data.
	// Set the default points to be 10.
	newQuestion := questions.Question{
		ID:           uuid.New(),
		Text:         "",
		ImageURL:     url.URL{},
		Article:      articles.Article{},
		QuizID:       quiz_id,
		Points:       100,
		Alternatives: []questions.Alternative{},
	}

	// Get all the articles for the quiz by quiz ID.
	articles, _ := articles.GetArticlesByQuizID(dph.sharedData.DB, quiz_id)

	return utils.Render(c, http.StatusOK, dashboard_components.EditQuestionForm(newQuestion, articles, quiz_id.String(), true))
}

// Renders the modal for editing a question.
func (dph *DashboardPagesHandler) dashboardEditQuestionModal(c echo.Context) error {
	// Get the question ID.
	question_id, err := uuid.Parse(c.QueryParam("question-id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid or missing question-id")
	}

	// Get the question by ID from the database.
	question, err := questions.GetQuestionByID(dph.sharedData.DB, question_id)
	if err != nil {
		if err == sql.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, "No question with given question-id found.")
		} else {
			return err
		}
	}

	// Get all the articles for the quiz by quiz ID.
	articles, _ := articles.GetArticlesByQuizID(dph.sharedData.DB, question.QuizID)

	return utils.Render(c, http.StatusOK, dashboard_components.EditQuestionForm(*question, articles, question.QuizID.String(), false))
}

func (dph *DashboardPagesHandler) leaderboard(c echo.Context) error {
	addMenuContext(c, side_menu.Leaderboard)
	return utils.Render(c, http.StatusOK, dashboard_pages.LeaderboardPage())
}

func (dph *DashboardPagesHandler) accessSettings(c echo.Context) error {
	addMenuContext(c, side_menu.AccessSettings)
	admins, err := access_control.GetAllAdmins(dph.sharedData.DB)
	if err != nil {
		return err
	}
	return utils.Render(c, http.StatusOK, dashboard_pages.AccessSettingsPage(admins))
}

func (dph *DashboardPagesHandler) userDetails(c echo.Context) error {
	// userId := c.QueryParam("user-id")
	return utils.Render(c, http.StatusOK, dashboard_pages.UserDetailsPage())
}

// Adds chosen menu item to the context, so it can be used in the template.
func addMenuContext(c echo.Context, menuContext side_menu.SideMenuItem) {
	utils.AddToContext(c, side_menu.MENU_CONTEXT_KEY, menuContext)
}

func (dph *DashboardPagesHandler) getUserAdministration(c echo.Context) error {
	userRows, err := users.GetUsersTableRows(dph.sharedData.DB, 1)
	if err != nil {
		return err
	}

	return utils.Render(c, http.StatusOK, dashboard_pages.UserAdminPage(
		userRows,
		1,
		20,
	))
}
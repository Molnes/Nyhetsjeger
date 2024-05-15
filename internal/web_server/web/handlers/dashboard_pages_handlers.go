package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/Molnes/Nyhetsjeger/internal/config"
	"github.com/Molnes/Nyhetsjeger/internal/models/articles"
	"github.com/Molnes/Nyhetsjeger/internal/models/labels"
	"github.com/Molnes/Nyhetsjeger/internal/models/questions"
	"github.com/Molnes/Nyhetsjeger/internal/models/quizzes"
	"github.com/Molnes/Nyhetsjeger/internal/models/users"
	"github.com/Molnes/Nyhetsjeger/internal/models/users/access_control"
	"github.com/Molnes/Nyhetsjeger/internal/models/users/user_ranking"
	"github.com/Molnes/Nyhetsjeger/internal/models/users/user_roles"
	"github.com/Molnes/Nyhetsjeger/internal/models/users/usernames"
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

	g.GET("/labels", dph.labels)
	g.GET("/user", dph.userDetails)
	g.GET("/username-admin", dph.getUsernameAdministration)

	mw := middlewares.NewAuthorizationMiddleware(dph.sharedData, []user_roles.Role{user_roles.OrganizationAdmin})
	organizationAdminGroup := g.Group("/organization-admin", mw.EnforceRole)
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

// Renders the labels page.
func (dph *DashboardPagesHandler) labels(c echo.Context) error {
	addMenuContext(c, side_menu.Labels)
	labels, err := labels.GetLabels(dph.sharedData.DB)
	if err != nil {
		return err
	}
	return utils.Render(c, http.StatusOK, dashboard_pages.LabelsPage(labels))
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

	// Get all available labels
	labels, err := labels.GetActiveLabels(dph.sharedData.DB)
	if err != nil {
		return err
	}

	return utils.Render(c, http.StatusOK, dashboard_pages.EditQuiz(quiz, articles, questions, labels))
}

// Renders the modal for creating a new question.
func (dph *DashboardPagesHandler) dashboardNewQuestionModal(c echo.Context) error {
	// Get the quiz ID.
	quizId, err := uuid.Parse(c.QueryParam("quiz-id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid or missing quiz id")
	}

	newQuestion := questions.GetDefaultQuestion(quizId)

	// Get all the articles for the quiz by quiz ID.
	articleList, _ := articles.GetArticlesByQuizID(dph.sharedData.DB, quizId)

	return utils.Render(c, http.StatusOK, dashboard_components.EditQuestionForm(&newQuestion, &articles.Article{}, articleList, quizId.String(), true))
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

	// Get the article if the ID is valid
	article := &articles.Article{}
	if question.ArticleID.Valid {
		article, err = articles.GetArticleByID(dph.sharedData.DB, question.ArticleID.UUID)
		if err != nil {
			if err == sql.ErrNoRows {
				return echo.NewHTTPError(http.StatusNotFound, "The article for this question could not be found.")
			} else {
				return err
			}
		}
	}

	// Get all the articles for the quiz by quiz ID.
	articles, _ := articles.GetArticlesByQuizID(dph.sharedData.DB, question.QuizID)

	return utils.Render(c, http.StatusOK, dashboard_components.EditQuestionForm(question, article, articles, question.QuizID.String(), false))
}

// Renders the leaderboard page.
func (dph *DashboardPagesHandler) leaderboard(c echo.Context) error {
	addMenuContext(c, side_menu.Leaderboard)

	labelID, err := uuid.Parse(c.QueryParam("label-id"))
	if err != nil {
		labelID = uuid.Nil
	}

	labels, err := labels.GetLabels(dph.sharedData.DB)
	if err != nil {
		return err
	}

	if labelID == uuid.Nil {
		if len(labels) > 0 {
			labelID = labels[0].ID
		}
	}

	rankings, err := user_ranking.GetRanking(dph.sharedData.DB, labelID)
	if err != nil {
		return err
	}

	return utils.Render(c, http.StatusOK, dashboard_pages.LeaderboardPage(rankings, labels, labelID))
}

// Renders the access settings page.
func (dph *DashboardPagesHandler) accessSettings(c echo.Context) error {
	addMenuContext(c, side_menu.AccessSettings)
	admins, err := access_control.GetAllAdmins(dph.sharedData.DB)
	if err != nil {
		return err
	}
	return utils.Render(c, http.StatusOK, dashboard_pages.AccessSettingsPage(admins))
}

// Renders the user details page.
func (dph *DashboardPagesHandler) userDetails(c echo.Context) error {
	uuid_id, err := uuid.Parse(c.QueryParam("user-id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Ugyldig eller manglende user-id")
	}
	user, err := users.GetUserByID(dph.sharedData.DB, uuid_id)
	if err != nil {
		if err == sql.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, "Fant ikke brukeren med den angitte ID-en")
		} else {
			return err
		}
	}

	allLabels, err := labels.GetLabels(dph.sharedData.DB)
	if err != nil {
		return err
	}
	labelID, err := uuid.Parse(c.QueryParam("label-id"))
	if err != nil {
		if len(allLabels) > 0 {
			labelID = allLabels[0].ID
		}

	}

	label, err := labels.GetLabelByID(dph.sharedData.DB, labelID)
	if err != nil {
		if err == sql.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, "Fant ikke label med den angitte ID-en")
		} else {
			return err
		}
	}

	rankingCollection, err := user_ranking.GetUserRankingsInAllRanges(dph.sharedData.DB, uuid_id, label)
	if err != nil {
		return err
	}

	return utils.Render(c, http.StatusOK, dashboard_pages.UserDetailsPage(user, rankingCollection, allLabels, labelID))
}

// Adds chosen menu item to the context, so it can be used in the template.
func addMenuContext(c echo.Context, menuContext side_menu.SideMenuItem) {
	utils.AddToContext(c, side_menu.MENU_CONTEXT_KEY, menuContext)
}

// Gets the usernames administration page and returns it.
func (dph *DashboardPagesHandler) getUsernameAdministration(c echo.Context) error {
	addMenuContext(c, side_menu.UserAdmin)

	adjPage, err := strconv.Atoi(c.QueryParam("adj"))
	if err != nil { // If the page number is not a number, set it to 1.
		adjPage = 1
	}
	nounPage, err := strconv.Atoi(c.QueryParam("noun"))
	if err != nil { // If the page number is not a number, set it to 1.
		nounPage = 1
	}

	pages, err := strconv.Atoi(c.QueryParam("rows-per-page"))
	if err != nil || pages < 5 || pages > 255 { // Sets to 25 if between a certain range.
		pages = 25
	}

	search := c.QueryParam("search")

	uai, err := usernames.GetUsernameAdminInfo(dph.sharedData.DB, adjPage, nounPage, pages, search)
	if err != nil {
		return err
	}

	return utils.Render(c, http.StatusOK, dashboard_pages.UsernameAdminPage(uai, c.Request().URL))
}

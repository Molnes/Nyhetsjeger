package handlers

import (
	"database/sql"
	"net/http"
	"net/url"

	"github.com/Molnes/Nyhetsjeger/internal/config"
	"github.com/Molnes/Nyhetsjeger/internal/models/articles"
	"github.com/Molnes/Nyhetsjeger/internal/models/questions"
	"github.com/Molnes/Nyhetsjeger/internal/models/quizzes"
	"github.com/Molnes/Nyhetsjeger/internal/utils"
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
func (dph *DashboardPagesHandler) RegisterDashboardHandlers(e *echo.Group) {
	e.GET("", dph.dashboardHomePage)
	e.GET("/edit-quiz", dph.dashboardEditQuiz)
	e.GET("/edit-quiz/new-question", dph.dashboardNewQuestionModal)
	e.GET("/leaderboard", dph.leaderboard)
	e.GET("/access-settings", dph.accessSettings)
	e.GET("/user-details", dph.userDetails)
}

// Renders the dashboard home page.
func (dph *DashboardPagesHandler) dashboardHomePage(c echo.Context) error {
	addMenuContext(c, side_menu.Home)

	nonPublishedQuizzes, err := quizzes.GetNonPublishedQuizzes(dph.sharedData.DB)
	if err != nil {
		return err
	}
	publishedQuizzes, err := quizzes.GetAllPublishedQuizzes(dph.sharedData.DB)
	if err != nil {
		return err
	}

	return utils.Render(c, http.StatusOK, dashboard_pages.DashboardHomePage(nonPublishedQuizzes, publishedQuizzes))
}

// Renders the page for editing quiz.
func (dph *DashboardPagesHandler) dashboardEditQuiz(c echo.Context) error {
	uuid_id, err := uuid.Parse(c.QueryParam("quiz-id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid or missing quiz id")
	}

	quiz, err := quizzes.GetQuizByID(dph.sharedData.DB, uuid_id)
	if err != nil {
		if err == sql.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, "No quiz with given id found.")
		} else {
			return err
		}
	}

	// Get all the articles for a quiz.
	articles, _ := articles.GetArticlesByQuizID(dph.sharedData.DB, uuid_id)

	return utils.Render(c, http.StatusOK, dashboard_pages.EditQuiz(quiz, articles))
}

// Renders the modal for creating a new question.
func (dph *DashboardPagesHandler) dashboardNewQuestionModal(c echo.Context) error {
	quiz_id, err := uuid.Parse(c.QueryParam("quiz-id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid or missing quiz id")
	}

	// Create a new question with no actual data.
	// Set the default points to be 10.
	newQuestion := questions.Question{
		ID:           uuid.New(),
		Text:         "SPØRSMÅL",
		ImageURL:     url.URL{},
		Article:      articles.Article{},
		QuizID:       quiz_id,
		Points:       10,
		Alternatives: []questions.Alternative{},
	}

	// Get all the articles.
	articles, _ := articles.GetArticlesByQuizID(dph.sharedData.DB, quiz_id)

	return utils.Render(c, http.StatusOK, dashboard_components.EditQuestionForm(newQuestion, articles))
}

// Update the image for a quiz.
func (dph *DashboardPagesHandler) removeImageFromQuiz(c echo.Context) error {
	quizID, _ := uuid.Parse(c.QueryParam("quiz-id"))
	quizImageURL := c.FormValue("quiz-image-url")
	quiz, _ := quizzes.GetQuizByID(dph.sharedData.DB, quizID)

	image_url, err := url.Parse(quizImageURL)

	if err != nil {
		return err
	}

	quizzes.UpdateImageByQuizID(dph.sharedData.DB, quizID, *image_url)

	return utils.Render(c, http.StatusOK, dashboard_components.ImagePreview(quiz.ImageURL))
}

func (dph *DashboardPagesHandler) leaderboard(c echo.Context) error {
	addMenuContext(c, side_menu.Leaderboard)
	return utils.Render(c, http.StatusOK, dashboard_pages.LeaderboardPage())
}

func (dph *DashboardPagesHandler) accessSettings(c echo.Context) error {
	addMenuContext(c, side_menu.AccessSettings)
	return utils.Render(c, http.StatusOK, dashboard_pages.AccessSettingsPage())
}

func (dph *DashboardPagesHandler) userDetails(c echo.Context) error {
	// userId := c.QueryParam("user-id")
	return utils.Render(c, http.StatusOK, dashboard_pages.UserDetailsPage())
}

// Adds chosen menu item to the context, so it can be used in the template.
func addMenuContext(c echo.Context, menuContext side_menu.SideMenuItem) {
	utils.AddToContext(c, side_menu.MENU_CONTEXT_KEY, menuContext)
}

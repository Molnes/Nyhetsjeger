package api

import (
	"net/http"
	"net/url"
	"time"

	"github.com/Molnes/Nyhetsjeger/internal/config"
	"github.com/Molnes/Nyhetsjeger/internal/models/quizzes"
	"github.com/Molnes/Nyhetsjeger/internal/utils"
	dashboard_components "github.com/Molnes/Nyhetsjeger/internal/web_server/web/views/components/dashboard_components/edit_quiz"
	"github.com/Molnes/Nyhetsjeger/internal/web_server/web/views/pages/dashboard_pages"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type AdminApiHandler struct {
	sharedData *config.SharedData
}

// Constants
const errorInvalidQuizID = "Invalid or missing quiz id"
const queryParamQuizID = "quiz-id"

// Creates a new AdminApiHandler
func NewAdminApiHandler(sharedData *config.SharedData) *AdminApiHandler {
	return &AdminApiHandler{sharedData}
}

// Registers handlers for admin api
func (aah *AdminApiHandler) RegisterAdminApiHandlers(e *echo.Group) {
	e.POST("/quiz/create-new", aah.postDefaultQuiz)
	e.POST("/quiz/edit-title", aah.editQuizTitle)
	e.POST("/quiz/edit-image", aah.editQuizImage)
	e.POST("/quiz/edit-published-status", aah.editQuizPublished)
	e.DELETE("/quiz/edit-image", aah.deleteQuizImage)
	e.DELETE("/delete-quiz", aah.deleteQuiz)
}

// Handles the creation of a new default quiz in the DB.
// Redirects to the edit quiz page for the newly created quiz.
func (aah *AdminApiHandler) postDefaultQuiz(c echo.Context) error {
	quiz := quizzes.CreateDefaultQuiz()
	quizzes.CreateQuiz(aah.sharedData.DB, quiz)

	c.Response().Header().Set("HX-Redirect", "/dashboard/edit-quiz?quiz-id="+quiz.ID.String())
	return c.Redirect(http.StatusOK, "/dashboard/edit-quiz?quiz-id="+quiz.ID.String())
}

// Updates the title of a quiz in the database.
func (aah *AdminApiHandler) editQuizTitle(c echo.Context) error {
	// Get the quiz ID
	quiz_id, err := uuid.Parse(c.QueryParam(queryParamQuizID))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorInvalidQuizID)
	}

	// Update the quiz title
	title := c.FormValue(dashboard_pages.QuizTitle)
	err = quizzes.UpdateTitleByQuizID(aah.sharedData.DB, quiz_id, title)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Failed to update quiz title")
	}

	time.Sleep(500 * time.Millisecond) // TODO: Remove

	return utils.Render(c, http.StatusOK, dashboard_components.EditTitleInput(title, quiz_id.String(), dashboard_pages.QuizTitle))
}

// Updates the image of a quiz in the database.
func (aah *AdminApiHandler) editQuizImage(c echo.Context) error {
	// Get the quiz ID
	quiz_id, err := uuid.Parse(c.QueryParam(queryParamQuizID))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorInvalidQuizID)
	}

	// Update the quiz image
	image := c.FormValue(dashboard_pages.QuizImageURL)
	imageURL, _ := url.Parse(image)
	err = quizzes.UpdateImageByQuizID(aah.sharedData.DB, quiz_id, *imageURL)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Failed to update quiz image")
	}

	time.Sleep(500 * time.Millisecond) // TODO: Remove

	return utils.Render(c, http.StatusOK, dashboard_components.EditImageInput(imageURL, quiz_id.String(), dashboard_pages.QuizImageURL))
}

// Removes the image for a quiz in the database.
func (dph *AdminApiHandler) deleteQuizImage(c echo.Context) error {
	// Get the quiz ID
	quiz_id, err := uuid.Parse(c.QueryParam(queryParamQuizID))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorInvalidQuizID)
	}

	// Set the image URL to nil
	err = quizzes.RemoveImageByQuizID(dph.sharedData.DB, quiz_id)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Failed to remove quiz image")
	}

	time.Sleep(500 * time.Millisecond) // TODO: Remove

	return utils.Render(c, http.StatusOK, dashboard_components.EditImageInput(&url.URL{}, quiz_id.String(), dashboard_pages.QuizImageURL))
}

// Deletes a quiz from the database.
func (aah *AdminApiHandler) deleteQuiz(c echo.Context) error {
	quiz_id, err := uuid.Parse(c.QueryParam(queryParamQuizID))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorInvalidQuizID)
	}

	quizzes.DeleteQuizByID(aah.sharedData.DB, quiz_id)

	c.Response().Header().Set("HX-Redirect", "/dashboard")
	return c.Redirect(http.StatusOK, "/dashboard")
}

// Updates the published status of a quiz in the database.
// If the quiz is published, it will be unpublished, and vice versa.
func (aah *AdminApiHandler) editQuizPublished(c echo.Context) error {
	// Get the quiz ID
	quiz_id, err := uuid.Parse(c.QueryParam(queryParamQuizID))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorInvalidQuizID)
	}

	// Update the quiz published status
	published := c.FormValue(dashboard_pages.QuizPublished)
	err = quizzes.UpdatePublishedStatusByQuizID(aah.sharedData.DB, quiz_id, !(published == "on"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Failed to update quiz published status")
	}

	time.Sleep(500 * time.Millisecond) // TODO: Remove

	return utils.Render(c, http.StatusOK, dashboard_components.ToggleQuizPublished(!(published == "on"), quiz_id.String(), dashboard_pages.QuizPublished))
}

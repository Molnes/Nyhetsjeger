package api

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/Molnes/Nyhetsjeger/internal/config"
	"github.com/Molnes/Nyhetsjeger/internal/models/articles"
	"github.com/Molnes/Nyhetsjeger/internal/models/questions"
	"github.com/Molnes/Nyhetsjeger/internal/models/quizzes"
	utils "github.com/Molnes/Nyhetsjeger/internal/utils"
	data_handling "github.com/Molnes/Nyhetsjeger/internal/utils/data"
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
	e.POST("/quiz/create-new", aah.createDefaultQuiz)
	e.POST("/quiz/edit-title", aah.editQuizTitle)
	e.POST("/quiz/edit-image", aah.editQuizImage)
	e.DELETE("/quiz/edit-image", aah.deleteQuizImage)
	e.POST("/quiz/edit-start", aah.editQuizActiveStart)
	e.POST("/quiz/edit-end", aah.editQuizActiveEnd)
	e.POST("/quiz/edit-published-status", aah.editQuizPublished)
	e.DELETE("/quiz/delete-quiz", aah.deleteQuiz)
	e.POST("/quiz/add-article", aah.addArticleToQuiz)
	e.DELETE("/quiz/delete-article", aah.deleteArticle)
	e.POST("/question/create-new", aah.createQuestion)
}

// Handles the creation of a new default quiz in the DB.
// Redirects to the edit quiz page for the newly created quiz.
func (aah *AdminApiHandler) createDefaultQuiz(c echo.Context) error {
	// Create a default quiz object
	quiz := quizzes.CreateDefaultQuiz()

	// Add quiz to database
	quizID, err := quizzes.CreateQuiz(aah.sharedData.DB, quiz)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Failed to create new quiz")
	}

	c.Response().Header().Set("HX-Redirect", "/dashboard/edit-quiz?quiz-id="+quizID.String())
	return c.Redirect(http.StatusOK, "/dashboard/edit-quiz?quiz-id="+quizID.String())
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

	return utils.Render(c, http.StatusOK, dashboard_components.EditImageInput(fmt.Sprintf("/api/v1/admin/quiz/edit-image?quiz-id=%s", quiz_id), imageURL, dashboard_pages.QuizImageURL, true))
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

	return utils.Render(c, http.StatusOK,
		dashboard_components.EditImageInput(fmt.Sprintf("/api/v1/admin/quiz/edit-image?quiz-id=%s", quiz_id), &url.URL{}, dashboard_pages.QuizImageURL, true))
}

// Deletes a quiz from the database.
func (aah *AdminApiHandler) deleteQuiz(c echo.Context) error {
	// Parse the quiz ID from the query parameter
	quiz_id, err := uuid.Parse(c.QueryParam(queryParamQuizID))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorInvalidQuizID)
	}

	// Sets the quiz as deleted in the database
	quizzes.DeleteQuizByID(aah.sharedData.DB, quiz_id)

	time.Sleep(500 * time.Millisecond) // TODO: Remove

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
	err = quizzes.UpdatePublishedStatusByQuizID(aah.sharedData.DB, quiz_id, published != "on")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Failed to update quiz published status")
	}

	time.Sleep(500 * time.Millisecond) // TODO: Remove

	return utils.Render(c, http.StatusOK, dashboard_components.ToggleQuizPublished(published != "on", quiz_id.String(), dashboard_pages.QuizPublished))
}

// Updates the active start time of a quiz in the database.
func (aah *AdminApiHandler) editQuizActiveStart(c echo.Context) error {
	// Get the quiz ID
	quiz_id, err := uuid.Parse(c.QueryParam(queryParamQuizID))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorInvalidQuizID)
	}

	// Get the time in Norway's timezone
	activeStart := c.FormValue(dashboard_pages.QuizActiveFrom)
	activeStartTime, err := data_handling.DateStringToNorwayTime(activeStart, c)
	if err != nil {
		return err
	}

	// Update the quiz active start
	err = quizzes.UpdateActiveStartByQuizID(aah.sharedData.DB, quiz_id, activeStartTime)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Failed to update quiz active start time")
	}

	time.Sleep(500 * time.Millisecond) // TODO: Remove

	return utils.Render(c, http.StatusOK, dashboard_components.EditActiveStartInput(activeStartTime, quiz_id.String(), dashboard_pages.QuizActiveFrom))
}

// Updates the active end time of a quiz in the database.
func (aah *AdminApiHandler) editQuizActiveEnd(c echo.Context) error {
	// Get the quiz ID
	quiz_id, err := uuid.Parse(c.QueryParam(queryParamQuizID))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorInvalidQuizID)
	}

	// Get the time in Norway's timezone
	activeEnd := c.FormValue(dashboard_pages.QuizActiveTo)
	activeEndTime, err := data_handling.DateStringToNorwayTime(activeEnd, c)
	if err != nil {
		return err
	}

	// Update the quiz active end
	err = quizzes.UpdateActiveEndByQuizID(aah.sharedData.DB, quiz_id, activeEndTime)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Failed to update quiz active end time")
	}

	time.Sleep(500 * time.Millisecond) // TODO: Remove

	return utils.Render(c, http.StatusOK, dashboard_components.EditActiveEndInput(activeEndTime, quiz_id.String(), dashboard_pages.QuizActiveFrom))
}

func (aah *AdminApiHandler) addArticleToQuiz(c echo.Context) error {
	// Get the quiz ID
	quiz_id, err := uuid.Parse(c.QueryParam(queryParamQuizID))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorInvalidQuizID)
	}

	// Get the article URL
	articleURL := c.FormValue(dashboard_pages.QuizArticleURL)
	tempURL, err := url.Parse(articleURL)
	if err != nil && err == sql.ErrNoRows {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid article URL")
	}

	// Check if the article is already in the DB
	article, err := articles.GetArticleByURL(aah.sharedData.DB, tempURL)
	if err != nil && err != sql.ErrNoRows {
		return echo.NewHTTPError(http.StatusBadRequest, "Failed to get article ID")
	}

	// If it exists, check if it already is in the quiz
	if article != nil && article.ID.Valid {
		articleInQuiz, err := articles.IsArticleInQuiz(aah.sharedData.DB, &article.ID.UUID, &quiz_id)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Failed to check if article is in quiz")
		}
		if articleInQuiz {
			return echo.NewHTTPError(http.StatusConflict, "Article is already in quiz")
		}
	} else {
		// If not in DB, fetch the relevant article data and add it to the DB
		tempArticle, err := articles.GetSmpArticleByURL(articleURL)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Failed to fetch article data")
		}

		articles.AddArticle(aah.sharedData.DB, &tempArticle)

		article = &tempArticle
	}

	// Add the article to the quiz
	err = articles.AddArticleToQuiz(aah.sharedData.DB, &article.ID.UUID, &quiz_id)

	time.Sleep(500 * time.Millisecond) // TODO: Remove

	return utils.Render(c, http.StatusOK, dashboard_components.ArticleListItem(articleURL, article.ID.UUID.String(), quiz_id.String()))
}

// Deletes an article from a quiz in the database.
func (aah *AdminApiHandler) deleteArticle(c echo.Context) error {
	// Get the quiz ID
	quiz_id, err := uuid.Parse(c.QueryParam(queryParamQuizID))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorInvalidQuizID)
	}

	// Get the article ID
	article_id, err := uuid.Parse(c.QueryParam("article-id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid or missing article ID")
	}

	// Remove the article from the quiz
	err = articles.DeleteArticleFromQuiz(aah.sharedData.DB, &quiz_id, &article_id)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Failed to delete article from quiz")
	}

	time.Sleep(500 * time.Millisecond) // TODO: Remove

	return c.NoContent(http.StatusOK)
}

// Creates a new question in the database with the given data.
func (aah *AdminApiHandler) createQuestion(c echo.Context) error {
	// Get the quiz ID
	quizID, err := uuid.Parse(c.QueryParam("quiz-id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid or missing quiz ID")
	}

	// Get the data from the form
	articleURLString := c.FormValue(dashboard_components.QuestionArticleURL)
	questionText := c.FormValue(dashboard_components.QuestionText)
	correctAnswerNumber := c.FormValue(dashboard_components.QuestionCorrectAlternative)
	alternative1 := c.FormValue(dashboard_components.QuestionAlternative1)
	alternative2 := c.FormValue(dashboard_components.QuestionAlternative2)
	alternative3 := c.FormValue(dashboard_components.QuestionAlternative3)
	alternative4 := c.FormValue(dashboard_components.QuestionAlternative4)
	imageURL := c.FormValue(dashboard_components.QuestionImageURL)
	questionPoints := c.FormValue(dashboard_components.QuestionPoints)

	// Parse the data and validate
	if questionText == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Missing question text")
	}

	points, err := strconv.ParseInt(questionPoints, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Failed to parse points")
	}
	if points < 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "Points must be positive")
	}

	tempArticleURL, err := url.Parse(articleURLString)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Failed to parse article URL")
	}

	image, err := url.Parse(imageURL)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Failed to parse image URL")
	}

	var article *articles.Article

	// Check if the article URL is in the database
	article, err = articles.GetArticleByURL(aah.sharedData.DB, tempArticleURL)
	if err != nil && err != sql.ErrNoRows {
		return echo.NewHTTPError(http.StatusBadRequest, "Failed to get article ID")
	}

	// If not in DB, fetch the relevant article data and add it to the DB
	if article == nil {
		tempArticle, err := articles.GetSmpArticleByURL(articleURLString)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Failed to fetch article data")
		}

		articles.AddArticle(aah.sharedData.DB, &tempArticle)
		article = &tempArticle
	}

	// Create a new question object
	questionID := uuid.New()

	question := questions.Question{
		ID:       questionID,
		Text:     questionText,
		ImageURL: *image,
		Article:  *article,
		QuizID:   quizID,
		Points:   uint(points),
		Alternatives: []questions.Alternative{
			{
				ID:         uuid.New(),
				Text:       alternative1,
				IsCorrect:  correctAnswerNumber == "1",
				QuestionID: questionID,
			},
			{
				ID:         uuid.New(),
				Text:       alternative2,
				IsCorrect:  correctAnswerNumber == "2",
				QuestionID: questionID,
			},
			{
				ID:         uuid.New(),
				Text:       alternative3,
				IsCorrect:  correctAnswerNumber == "3",
				QuestionID: questionID,
			},
			{
				ID:         uuid.New(),
				Text:       alternative4,
				IsCorrect:  correctAnswerNumber == "4",
				QuestionID: questionID,
			},
		},
	}

	// Save the question to the database
	err = questions.AddNewQuestion(aah.sharedData.DB, question)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Failed to create new question")
	}
	// Return the "question item" element
	return utils.Render(c, http.StatusOK, dashboard_components.QuestionListItem(&question))
}

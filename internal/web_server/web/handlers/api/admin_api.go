package api

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/Molnes/Nyhetsjeger/internal/config"
	"github.com/Molnes/Nyhetsjeger/internal/models/articles"
	"github.com/Molnes/Nyhetsjeger/internal/models/questions"
	"github.com/Molnes/Nyhetsjeger/internal/models/quizzes"
	utils "github.com/Molnes/Nyhetsjeger/internal/utils"
	data_handling "github.com/Molnes/Nyhetsjeger/internal/utils/data"
	dashboard_components "github.com/Molnes/Nyhetsjeger/internal/web_server/web/views/components/dashboard_components/edit_quiz"
	"github.com/Molnes/Nyhetsjeger/internal/web_server/web/views/components/dashboard_components/edit_quiz/composite_components"
	"github.com/Molnes/Nyhetsjeger/internal/web_server/web/views/pages/dashboard_pages"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type AdminApiHandler struct {
	sharedData *config.SharedData
}

// Constants
const queryParamQuizID = "quiz-id"
const errorInvalidQuizID = "Ugyldig eller manglende quiz-id"
const queryParamQuestionID = "question-id"
const errorInvalidQuestionID = "Ugyldig eller manglende question-id"

// URLs
const editQuizURL = "/api/v1/admin/quiz/edit-image?quiz-id=%s"
const editQuestionURL = "/api/v1/admin/question/edit-image?question-id=%s"

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
	e.POST("/question/edit", aah.editQuestion)
	e.POST("/question/edit-image", aah.editQuestionImage)
	e.DELETE("/question/edit-image", aah.deleteQuestionImage)
	e.DELETE("/question/delete", aah.deleteQuestion)
}

// Handles the creation of a new default quiz in the DB.
// Redirects to the edit quiz page for the newly created quiz.
func (aah *AdminApiHandler) createDefaultQuiz(c echo.Context) error {
	// Create a default quiz object
	quiz := quizzes.CreateDefaultQuiz()

	// Add quiz to database
	quizID, err := quizzes.CreateQuiz(aah.sharedData.DB, quiz)
	if err != nil {
		return err
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
	if strings.TrimSpace(title) == "" {
		return utils.Render(c, http.StatusOK,
			dashboard_components.EditTitleInput(title, quiz_id.String(), dashboard_pages.QuizTitle, "Tittelen kan ikke være tom"),
		)
	}

	err = quizzes.UpdateTitleByQuizID(aah.sharedData.DB, quiz_id, title)
	if err != nil {
		return err
	}

	return utils.Render(c, http.StatusOK, dashboard_components.EditTitleInput(title, quiz_id.String(), dashboard_pages.QuizTitle, ""))
}

// Updates the image of a quiz in the database.
func (aah *AdminApiHandler) editQuizImage(c echo.Context) error {
	// Get the quiz ID
	quiz_id, err := uuid.Parse(c.QueryParam(queryParamQuizID))
	if err != nil {
		return utils.Render(c, http.StatusOK, dashboard_components.EditImageInput(
			fmt.Sprintf(editQuizURL, quiz_id), &url.URL{}, dashboard_pages.QuizImageURL, true, errorInvalidQuizID))
	}

	// Update the quiz image
	image := c.FormValue(dashboard_pages.QuizImageURL)
	imageURL, err := url.Parse(image)
	if err != nil {
		return utils.Render(c, http.StatusOK, dashboard_components.EditImageInput(
			fmt.Sprintf(editQuizURL, quiz_id), &url.URL{}, dashboard_pages.QuizImageURL, true, "Ugyldig bilde URL"))
	}

	// Set the image URL for the quiz
	err = quizzes.UpdateImageByQuizID(aah.sharedData.DB, quiz_id, *imageURL)
	if err != nil {
		if err == questions.ErrNoImageUpdated {
			return utils.Render(c, http.StatusOK, dashboard_components.EditImageInput(
				fmt.Sprintf(editQuizURL, quiz_id), &url.URL{}, dashboard_pages.QuizImageURL, true, "Quiz bilde kunne ikke bli oppdatert. Prøv igjen senere"))
		}

		return err
	}

	return utils.Render(c, http.StatusOK, dashboard_components.EditImageInput(
		fmt.Sprintf(editQuizURL, quiz_id), imageURL, dashboard_pages.QuizImageURL, true, ""))
}

// Removes the image for a quiz in the database.
func (dph *AdminApiHandler) deleteQuizImage(c echo.Context) error {
	// Get the quiz ID
	quiz_id, err := uuid.Parse(c.QueryParam(queryParamQuizID))
	if err != nil {
		return utils.Render(c, http.StatusOK, dashboard_components.EditImageInput(
			fmt.Sprintf(editQuizURL, quiz_id), &url.URL{}, dashboard_pages.QuizImageURL, true, errorInvalidQuizID))
	}

	// Set the image URL to nil
	err = quizzes.RemoveImageByQuizID(dph.sharedData.DB, quiz_id)
	if err != nil {
		return utils.Render(c, http.StatusOK, dashboard_components.EditImageInput(
			fmt.Sprintf(editQuizURL, quiz_id), &url.URL{}, dashboard_pages.QuizImageURL, true, "Quiz bilde kunne ikke bli fjernet. Prøv igjen senere"))
	}

	return utils.Render(c, http.StatusOK, dashboard_components.EditImageInput(
		fmt.Sprintf(editQuizURL, quiz_id), &url.URL{}, dashboard_pages.QuizImageURL, true, ""))
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
		return err
	}

	return utils.Render(c, http.StatusOK, dashboard_components.ToggleQuizPublished(published != "on", quiz_id.String(), dashboard_pages.QuizPublished))
}

// Updates the active start time of a quiz in the database.
func (aah *AdminApiHandler) editQuizActiveStart(c echo.Context) error {
	// Get the quiz ID
	quiz_id, err := uuid.Parse(c.QueryParam(queryParamQuizID))

	// Get the time in Norway's timezone
	activeStart := c.FormValue(dashboard_pages.QuizActiveFrom)
	activeStartTime, startErr := data_handling.DateStringToNorwayTime(activeStart, c)
	if startErr != nil {
		return startErr
	}

	activeEnd := c.FormValue(dashboard_pages.QuizActiveTo)
	activeEndTime, endErr := data_handling.DateStringToNorwayTime(activeEnd, c)
	if endErr != nil {
		return endErr
	}

	if err != nil {
		return utils.Render(c, http.StatusOK, composite_components.EditActiveTimeInput(
			quiz_id.String(), activeStartTime, dashboard_pages.QuizActiveFrom,
			activeEndTime, dashboard_pages.QuizActiveTo, errorInvalidQuizID))
	}

	// Ensure that the start time is before end time
	if !activeStartTime.Before(activeEndTime) {
		return utils.Render(c, http.StatusOK, composite_components.EditActiveTimeInput(
			quiz_id.String(), activeStartTime, dashboard_pages.QuizActiveFrom,
			activeEndTime, dashboard_pages.QuizActiveTo, "Starttidspunktet må være før sluttidspunktet"))
	}

	// Update the quiz active start
	err = quizzes.UpdateActiveStartByQuizID(aah.sharedData.DB, quiz_id, activeStartTime)
	if err != nil {
		return err
	}

	return utils.Render(c, http.StatusOK, composite_components.EditActiveTimeInput(
		quiz_id.String(), activeStartTime, dashboard_pages.QuizActiveFrom,
		activeEndTime, dashboard_pages.QuizActiveTo, ""))
}

// Updates the active end time of a quiz in the database.
func (aah *AdminApiHandler) editQuizActiveEnd(c echo.Context) error {
	// Get the quiz ID
	quiz_id, err := uuid.Parse(c.QueryParam(queryParamQuizID))

	// Get the time in Norway's timezone
	activeEnd := c.FormValue(dashboard_pages.QuizActiveTo)
	activeEndTime, startErr := data_handling.DateStringToNorwayTime(activeEnd, c)
	if startErr != nil {
		return startErr
	}

	activeStart := c.FormValue(dashboard_pages.QuizActiveFrom)
	activeStartTime, endErr := data_handling.DateStringToNorwayTime(activeStart, c)
	if endErr != nil {
		return endErr
	}

	if err != nil {
		return utils.Render(c, http.StatusOK, composite_components.EditActiveTimeInput(
			quiz_id.String(), activeStartTime, dashboard_pages.QuizActiveFrom,
			activeEndTime, dashboard_pages.QuizActiveTo, errorInvalidQuizID))
	}

	// Ensure that the end time is after start time
	if !activeEndTime.After(activeStartTime) {
		return utils.Render(c, http.StatusOK, composite_components.EditActiveTimeInput(
			quiz_id.String(), activeStartTime, dashboard_pages.QuizActiveFrom,
			activeEndTime, dashboard_pages.QuizActiveTo, "Sluttidspunktet må være etter starttidspunktet"))
	}

	// Update the quiz active end
	err = quizzes.UpdateActiveEndByQuizID(aah.sharedData.DB, quiz_id, activeEndTime)
	if err != nil {
		return err
	}

	return utils.Render(c, http.StatusOK, composite_components.EditActiveTimeInput(
		quiz_id.String(), activeStartTime, dashboard_pages.QuizActiveFrom,
		activeEndTime, dashboard_pages.QuizActiveTo, ""))
}

// Adds the article to the database if it doesn't already exist.
// If the article is already in the DB, it will check if it is already in the quiz.
// If article already is in the quiz, return an error.
// If not in the DB, it will fetch the relevant article data and add it to the DB.
func conditionallyAddArticle(db *sql.DB, articleURL *url.URL, quizID *uuid.UUID) (*articles.Article, string) {
	// Check if the article is already in the DB
	article, err := articles.GetArticleByURL(db, articleURL)
	if err != nil && err != sql.ErrNoRows {
		return article, "Klarte ikke å finne artikkel ID i URL"
	}

	// If it exists, check if it already is in the quiz
	if article != nil && article.ID.Valid {
		articleInQuiz, err := articles.IsArticleInQuiz(db, &article.ID.UUID, quizID)
		if err != nil {
			return article, "Klarte ikke å sjekke om artikkelen allerede er i quizen. Prøv igjen senere"
		}
		if articleInQuiz {
			return article, "Artikkelen er allerede i quizen"
		}
	} else {
		// If not in DB, fetch the relevant article data and add it to the DB
		tempArticle, err := articles.GetSmpArticleByURL(articleURL.String())
		if err != nil {
			if err == articles.ErrInvalidArticleID {
				return article, "Ugyldig artikkel ID"
			} else if err == articles.ErrInvalidArticleURL {
				return article, "Ugyldig artikkel URL"
			} else if err == articles.ErrArticleNotFound {
				return article, "Klarte ikke å finne artikkel data for denne URLen. Sjekk at URLen er riktig eller prøv igjen senere"
			} else {
				return article, err.Error()
			}
		}

		// Add the article to the DB
		articles.AddArticle(db, &tempArticle)

		article = &tempArticle
	}

	return article, ""
}

// Adds an article to a quiz in the database.
func (aah *AdminApiHandler) addArticleToQuiz(c echo.Context) error {
	// Get the quiz ID
	quiz_id, err := uuid.Parse(c.QueryParam(queryParamQuizID))
	if err != nil {
		return utils.Render(c, http.StatusOK, composite_components.ArticleInputAndItem(
			"", "", quiz_id.String(), dashboard_pages.QuizArticleURL, errorInvalidQuizID))
	}

	// Get the article URL
	articleURL := c.FormValue(dashboard_pages.QuizArticleURL)
	tempURL, err := url.Parse(articleURL)
	if err != nil && err == sql.ErrNoRows {
		return utils.Render(c, http.StatusOK, composite_components.ArticleInputAndItem(
			"", "", quiz_id.String(), dashboard_pages.QuizArticleURL, "Ugyldig artikkel URL"))
	}

	// Ensure the article is in the database
	article, errText := conditionallyAddArticle(aah.sharedData.DB, tempURL, &quiz_id)
	if errText != "" {
		return utils.Render(c, http.StatusOK, composite_components.ArticleInputAndItem(
			"", "", quiz_id.String(), dashboard_pages.QuizArticleURL, errText))
	}

	// Add the article to the quiz
	err = articles.AddArticleToQuiz(aah.sharedData.DB, &article.ID.UUID, &quiz_id)
	if err != nil {
		return utils.Render(c, http.StatusOK, composite_components.ArticleInputAndItem(
			"", "", quiz_id.String(), dashboard_pages.QuizArticleURL, "Artikkelen kunne ikke bli lagt til. Prøv igjen senere"))
	}

	return utils.Render(c, http.StatusOK, composite_components.ArticleInputAndItem(
		articleURL, article.ID.UUID.String(), quiz_id.String(), dashboard_pages.QuizArticleURL, ""),
	)
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
		return echo.NewHTTPError(http.StatusBadRequest, "Ugyldig eller manglende artikkel ID")
	}

	// Remove the article from the quiz
	err = articles.DeleteArticleFromQuiz(aah.sharedData.DB, &quiz_id, &article_id)
	if err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

// Edit a question with the given data.
// If the question ID is not found, a new question will be created.
// If the question ID is found, the question will be updated.
func (aah *AdminApiHandler) editQuestion(c echo.Context) error {
	// Get the quiz ID
	quizID, err := uuid.Parse(c.QueryParam(queryParamQuizID))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorInvalidQuizID)
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
	timeLimit := c.FormValue(dashboard_components.QuestionTimeLimit)

	// Parse the data and validate
	points, articleURL, image, time, errorText := questions.ParseAndValidateQuestionData(questionText, questionPoints, articleURLString, imageURL, timeLimit)
	if errorText != "" {
		return echo.NewHTTPError(http.StatusBadRequest, errorText)
	}

	article := &articles.Article{}

	// Only add article to DB if it is not empty.
	// I.e. allow for no article, but not invalid article.
	if articleURLString != "" {
		// Check if the article URL is in the database
		article, err = articles.GetArticleByURL(aah.sharedData.DB, articleURL)
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
	}

	// Get the question ID.
	questionID, err := uuid.Parse(c.QueryParam(queryParamQuestionID))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorInvalidQuestionID)
	}

	// Create a new question object
	questionForm := questions.QuestionForm{
		ID:                  questionID,
		Text:                questionText,
		ImageURL:            image,
		Article:             article,
		QuizID:              &quizID,
		Points:              points,
		TimeLimitSeconds:    time,
		CorrectAnswerNumber: correctAnswerNumber,
		Alternative1:        alternative1,
		Alternative2:        alternative2,
		Alternative3:        alternative3,
		Alternative4:        alternative4,
	}
	question, errorText := questions.CreateQuestionFromForm(questionForm)
	if errorText != "" {
		return echo.NewHTTPError(http.StatusBadRequest, errorText)
	}

	// Get the question by ID from the database.
	tempQuestion, err := questions.GetQuestionByID(aah.sharedData.DB, questionID)

	// If the question doesn't exist in the database.
	if err != nil && err == sql.ErrNoRows {
		// Save the question to the database.
		err = questions.AddNewQuestion(aah.sharedData.DB, c.Request().Context(), &question)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Failed to create new question")
		}
	} else if tempQuestion.ID == questionID {
		// If the question ID is found, update the question.
		question.ID = questionID
		err = questions.UpdateQuestion(aah.sharedData.DB, c.Request().Context(), &question)

		if err != nil {
			if err == questions.ErrNoQuestionUpdated {
				return echo.NewHTTPError(http.StatusNotFound, "Question not found")
			}

			return err
		}
	} else if err != nil {
		return err
	}

	// Return the "question item" element.
	return utils.Render(c, http.StatusOK, dashboard_components.QuestionListItem(&question))
}

// Delete a question with the given ID from the database.
func (aah *AdminApiHandler) deleteQuestion(c echo.Context) error {
	// Get the question ID
	questionID, err := uuid.Parse(c.QueryParam(queryParamQuestionID))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorInvalidQuestionID)
	}

	// Delete the question from the database
	err = questions.DeleteQuestionByID(aah.sharedData.DB, c.Request().Context(), &questionID)
	if err != nil {
		if err == questions.ErrNoQuestionDeleted {
			return echo.NewHTTPError(http.StatusNotFound, "Question not found")
		}

		return err
	}

	return c.NoContent(http.StatusOK)
}

// Edit the image for a question in the database.
func (aah *AdminApiHandler) editQuestionImage(c echo.Context) error {
	// Get the question ID
	questionID, err := uuid.Parse(c.QueryParam(queryParamQuestionID))
	if err != nil {
		return utils.Render(c, http.StatusOK, dashboard_components.EditImageInput(
			fmt.Sprintf(editQuestionURL, questionID), &url.URL{}, dashboard_components.QuestionImageURL, true, errorInvalidQuestionID))
	}

	// Get the new image URL
	image := c.FormValue(dashboard_components.QuestionImageURL)
	imageURL, err := url.Parse(image)
	if err != nil {
		return utils.Render(c, http.StatusOK, dashboard_components.EditImageInput(
			fmt.Sprintf(editQuestionURL, questionID), &url.URL{}, dashboard_components.QuestionImageURL, true, "Ugyldig bilde URL"))
	}

	// Set the image URL for the question
	err = questions.SetImageByQuestionID(aah.sharedData.DB, &questionID, imageURL)
	if err != nil {
		return utils.Render(c, http.StatusOK, dashboard_components.EditImageInput(
			fmt.Sprintf(editQuestionURL, questionID), &url.URL{}, dashboard_components.QuestionImageURL, true, "Spørsmål bilde kunne ikke bli oppdatert. Prøv igjen senere"))
	}

	return utils.Render(c, http.StatusOK, dashboard_components.EditImageInput(
		fmt.Sprintf(editQuestionURL, questionID), imageURL, dashboard_components.QuestionImageURL, true, ""))
}

// Delete the image for a question in the database.
func (aah *AdminApiHandler) deleteQuestionImage(c echo.Context) error {
	// Get the question ID
	questionID, err := uuid.Parse(c.QueryParam(queryParamQuestionID))
	if err != nil {
		return utils.Render(c, http.StatusOK, dashboard_components.EditImageInput(
			fmt.Sprintf(editQuestionURL, questionID), &url.URL{}, dashboard_components.QuestionImageURL, true, errorInvalidQuestionID))
	}

	// Remove the image URL from the question
	err = questions.RemoveImageByQuestionID(aah.sharedData.DB, &questionID)
	if err != nil {
		if err == questions.ErrNoImageRemoved {
			return utils.Render(c, http.StatusOK, dashboard_components.EditImageInput(
				fmt.Sprintf(editQuestionURL, questionID), &url.URL{}, dashboard_components.QuestionImageURL, true, "Spørsmål bilde kunne ikke bli fjernet. Prøv igjen senere"))
		}

		return err
	}

	return utils.Render(c, http.StatusOK, dashboard_components.EditImageInput(
		fmt.Sprintf(editQuestionURL, questionID), &url.URL{}, dashboard_components.QuestionImageURL, true, ""))
}

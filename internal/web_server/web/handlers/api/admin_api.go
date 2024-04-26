package api

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Molnes/Nyhetsjeger/internal/config"
	"github.com/Molnes/Nyhetsjeger/internal/models/articles"
	"github.com/Molnes/Nyhetsjeger/internal/models/questions"
	"github.com/Molnes/Nyhetsjeger/internal/models/quizzes"
	"github.com/Molnes/Nyhetsjeger/internal/models/users/usernames"
	utils "github.com/Molnes/Nyhetsjeger/internal/utils"
	data_handling "github.com/Molnes/Nyhetsjeger/internal/utils/data"
	"github.com/Molnes/Nyhetsjeger/internal/web_server/web/views/components"
	dashboard_components "github.com/Molnes/Nyhetsjeger/internal/web_server/web/views/components/dashboard_components/edit_quiz"
	"github.com/Molnes/Nyhetsjeger/internal/web_server/web/views/components/dashboard_components/edit_quiz/composite_components"
	"github.com/Molnes/Nyhetsjeger/internal/web_server/web/views/components/dashboard_components/user_admin"
	"github.com/Molnes/Nyhetsjeger/internal/web_server/web/views/pages/dashboard_pages"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/minio/minio-go/v7"
)

type AdminApiHandler struct {
	sharedData *config.SharedData
}

// Constants
const (
	queryParamQuizID       = "quiz-id"
	errorInvalidQuizID     = "Ugyldig eller manglende quiz-id"
	queryParamQuestionID   = "question-id"
	errorInvalidQuestionID = "Ugyldig eller manglende question-id"
	errorQuestionElementID = "error-question"
	errorUploadImage       = "Kunne ikke laste opp bildet"
	errorFetchingImage     = "Kunne ikke hente bildet"
	imageURLInput          = "image-url"
	imageFileInput         = "image-file"
	errorImageElementID    = "error-image"
)

// URLs
const (
	editQuizImageURL         = "/api/v1/admin/quiz/edit-image?quiz-id=%s"
	editQuizImageFile        = "/api/v1/admin/quiz/upload-image?quiz-id=%s"
	editQuestionImageURL     = "/api/v1/admin/question/edit-image?question-id=%s"
	editQuestionImageFile    = "/api/v1/admin/question/upload-image?question-id=%s"
	imageSuggestionsQuiz     = "/api/v1/admin/quiz/image/update-suggestions?quiz-id=%s"
	imageSuggestionsQuestion = "/api/v1/admin/question/image/update-suggestions?question-id=%s"
	bucketImageURL           = "/images/"
)

// Creates a new AdminApiHandler
func NewAdminApiHandler(sharedData *config.SharedData) *AdminApiHandler {
	return &AdminApiHandler{sharedData}
}

// Registers handlers for admin api
func (aah *AdminApiHandler) RegisterAdminApiHandlers(e *echo.Group) {
	e.POST("/quiz/create-new", aah.createDefaultQuiz)
	e.POST("/quiz/edit-title", aah.editQuizTitle)
	e.POST("/quiz/edit-image", aah.editQuizImage)
	e.POST("/quiz/upload-image", aah.uploadQuizImage)
	e.DELETE("/quiz/edit-image", aah.deleteQuizImage)
	e.POST("/quiz/edit-start", aah.editQuizActiveStart)
	e.POST("/quiz/edit-end", aah.editQuizActiveEnd)
	e.POST("/quiz/edit-published-status", aah.editQuizPublished)
	e.DELETE("/quiz/delete-quiz", aah.deleteQuiz)
	e.POST("/quiz/add-article", aah.addArticleToQuiz)
	e.DELETE("/quiz/delete-article", aah.deleteArticle)
	e.POST("/quiz/rearrange-questions", aah.rearrangeQuestions)
	e.GET("/quiz/image/update-suggestions", aah.imageSuggestionsQuiz)

	e.POST("/question/edit", aah.editQuestion)
	e.POST("/question/edit-image", aah.editQuestionImage)
	e.POST("/question/upload-image", aah.uploadQuestionImage)
	e.DELETE("/question/edit-image", aah.deleteQuestionImage)
	e.DELETE("/question/delete", aah.deleteQuestion)
	e.POST("/question/randomize-alternatives", aah.randomizeAlternatives)
	e.GET("/question/image/update-suggestions", aah.imageSuggestionsQuestion)

	e.POST("/username", aah.addUsername)
	e.DELETE("/username", aah.deleteUsername)
	e.POST("/username/edit", aah.editUsername)
	e.POST("/username/page", aah.getUsernamePages)
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
		return utils.Render(c, http.StatusBadRequest, components.ErrorText("error-title", errorInvalidQuizID))
	}

	// Update the quiz title
	title := c.FormValue(dashboard_pages.QuizTitle)
	if strings.TrimSpace(title) == "" {
		return utils.Render(c, http.StatusBadRequest, components.ErrorText("error-title", "Tittelen kan ikke være tom"))
	}

	err = quizzes.UpdateTitleByQuizID(aah.sharedData.DB, quiz_id, title)
	if err != nil {
		return err
	}

	return utils.Render(c, http.StatusOK, dashboard_components.EditTitleInput(
		title, quiz_id.String(), dashboard_pages.QuizTitle, ""))
}

// Updates the image of a quiz in the database.
func (aah *AdminApiHandler) editQuizImage(c echo.Context) error {
	// Get the quiz ID
	quiz_id, err := uuid.Parse(c.QueryParam(queryParamQuizID))
	if err != nil {
		return utils.Render(c, http.StatusBadRequest, components.ErrorText(errorImageElementID, errorInvalidQuizID))
	}

	// Update the quiz image
	image := c.FormValue(imageURLInput)
	imageURL, err := url.Parse(image)
	if err != nil {
		return utils.Render(c, http.StatusBadRequest, components.ErrorText(errorImageElementID, "Ugyldig bilde URL"))
	}

	imageName, err := aah.uploadImageFromURL(c, *imageURL)
	if err != nil {
		return utils.Render(c, http.StatusBadRequest, components.ErrorText(errorImageElementID, errorUploadImage))
	}

	imageURL, err = url.Parse(aah.sharedData.Bucket.EndpointURL().String() + bucketImageURL + imageName)
	if err != nil {
		return utils.Render(c, http.StatusBadRequest, components.ErrorText(errorImageElementID, errorUploadImage))
	}

	// Set the image URL for the quiz
	err = quizzes.UpdateImageByQuizID(aah.sharedData.DB, quiz_id, *imageURL)
	if err != nil {
		return err
	}

	return utils.Render(c, http.StatusOK, dashboard_components.EditImageInput(
		fmt.Sprintf(editQuizImageURL, quiz_id), fmt.Sprintf(editQuizImageFile, quiz_id),
		fmt.Sprintf(imageSuggestionsQuiz, quiz_id), imageURL, true, "", dashboard_components.IdPrefixQuiz))
}

func (aah *AdminApiHandler) uploadQuizImage(c echo.Context) error {
	// Get the quiz ID
	quiz_id, err := uuid.Parse(c.QueryParam(queryParamQuizID))
	if err != nil {
		return utils.Render(c, http.StatusBadRequest, components.ErrorText(errorImageElementID, errorInvalidQuizID))

	}

	// Get the image file
	image, err := c.FormFile(imageFileInput)
	if err != nil {
		log.Println(err)
		return utils.Render(c, http.StatusBadRequest, components.ErrorText(errorImageElementID, errorFetchingImage))
	}

	// Upload the image to the bucket
	imageName, err := aah.uploadImage(c, image)
	if err != nil {
		log.Println(err)
		return utils.Render(c, http.StatusBadRequest, components.ErrorText(errorImageElementID, errorUploadImage))
	}

	// Set the image URL for the quiz
	imageURL := aah.sharedData.Bucket.EndpointURL().String() + bucketImageURL + imageName

	// Parse the image URL
	imageAsURL, err := url.Parse(imageURL)
	if err != nil {
		log.Println(err)
		return utils.Render(c, http.StatusBadRequest, components.ErrorText(errorImageElementID, "Kunne ikke fullføre bildeopplastingen"))
	}

	err = quizzes.UpdateImageByQuizID(aah.sharedData.DB, quiz_id, *imageAsURL)
	if err != nil {
		log.Println(err)
		return err
	}

	return utils.Render(c, http.StatusOK, dashboard_components.EditImageInput(
		fmt.Sprintf(editQuizImageURL, quiz_id), fmt.Sprintf(editQuizImageFile, quiz_id),
		fmt.Sprintf(imageSuggestionsQuiz, quiz_id), imageAsURL, true, "", dashboard_components.IdPrefixQuiz))
}

// Removes the image for a quiz in the database.
func (dph *AdminApiHandler) deleteQuizImage(c echo.Context) error {
	// Get the quiz ID
	quiz_id, err := uuid.Parse(c.QueryParam(queryParamQuizID))
	if err != nil {
		return utils.Render(c, http.StatusBadRequest, components.ErrorText(errorImageElementID, errorInvalidQuizID))
	}

	// Set the image URL to nil
	err = quizzes.RemoveImageByQuizID(dph.sharedData.DB, quiz_id)
	if err != nil {
		return err
	}

	return utils.Render(c, http.StatusOK, dashboard_components.EditImageInput(
		fmt.Sprintf(editQuizImageURL, quiz_id), fmt.Sprintf(editQuizImageFile, quiz_id),
		fmt.Sprintf(imageSuggestionsQuiz, quiz_id), &url.URL{}, true, "", dashboard_components.IdPrefixQuiz))
}

// Deletes a quiz from the database.
func (aah *AdminApiHandler) deleteQuiz(c echo.Context) error {
	// Parse the quiz ID from the query parameter
	quiz_id, err := uuid.Parse(c.QueryParam(queryParamQuizID))
	if err != nil {
		return utils.Render(c, http.StatusBadRequest, components.ErrorText("error-quiz",
			fmt.Sprintf("Kunne ikke slette quiz: %s. (Feil oppstod: %s)", errorInvalidQuizID, data_handling.GetNorwayTime(time.Now()))))
	}

	// Sets the quiz as deleted in the database
	err = quizzes.DeleteQuizByID(aah.sharedData.DB, quiz_id)
	if err != nil {
		return err
	}

	c.Response().Header().Set("HX-Redirect", "/dashboard")
	return c.Redirect(http.StatusOK, "/dashboard")
}

// Updates the published status of a quiz in the database.
// If the quiz is published, it will be unpublished, and vice versa.
func (aah *AdminApiHandler) editQuizPublished(c echo.Context) error {
	// Get the quiz ID
	quiz_id, err := uuid.Parse(c.QueryParam(queryParamQuizID))
	if err != nil {
		return utils.Render(c, http.StatusBadRequest, components.ErrorText("error-quiz",
			fmt.Sprintf("Kunne ikke skjule/publisere quiz: %s. (Feil oppstod: %s)", errorInvalidQuizID, data_handling.GetNorwayTime(time.Now()))))
	}

	// Update the quiz published status
	published := c.FormValue(dashboard_pages.QuizPublished)
	err = quizzes.UpdatePublishedStatusByQuizID(aah.sharedData.DB, c.Request().Context(), quiz_id, published == "on")
	if err != nil {
		if err == quizzes.ErrNoQuestions {
			return utils.Render(c, http.StatusBadRequest, components.ErrorText("error-quiz",
				"Kan ikke publisere en quiz uten spørsmål. Legg til minst ett spørsmål før du publiserer quizen."))
		}

		return err
	}

	return utils.Render(c, http.StatusOK, dashboard_components.ToggleQuizPublished(published == "on", quiz_id.String(), dashboard_pages.QuizPublished))
}

const errorActiveTimeElementID = "error-active-time"

// Updates the active start time of a quiz in the database.
func (aah *AdminApiHandler) editQuizActiveStart(c echo.Context) error {
	// Get the quiz ID
	quiz_id, err := uuid.Parse(c.QueryParam(queryParamQuizID))
	if err != nil {
		return utils.Render(c, http.StatusBadRequest, components.ErrorText(errorActiveTimeElementID, errorInvalidQuizID))
	}

	// Get the time in Norway's timezone
	activeStart := c.FormValue(dashboard_pages.QuizActiveFrom)
	activeStartTime, err := data_handling.DateStringToNorwayTime(activeStart)
	if err != nil {
		return err
	}

	activeEnd := c.FormValue(dashboard_pages.QuizActiveTo)
	activeEndTime, err := data_handling.DateStringToNorwayTime(activeEnd)
	if err != nil {
		return err
	}

	// Ensure that the start time is before end time
	if !activeStartTime.Before(activeEndTime) {
		return utils.Render(c, http.StatusBadRequest, components.ErrorText(
			errorActiveTimeElementID, "Starttidspunktet må være før sluttidspunktet"))
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
	if err != nil {
		return utils.Render(c, http.StatusBadRequest, components.ErrorText(errorActiveTimeElementID, errorInvalidQuizID))
	}

	// Get the time in Norway's timezone
	activeEnd := c.FormValue(dashboard_pages.QuizActiveTo)
	activeEndTime, err := data_handling.DateStringToNorwayTime(activeEnd)
	if err != nil {
		return err
	}

	activeStart := c.FormValue(dashboard_pages.QuizActiveFrom)
	activeStartTime, err := data_handling.DateStringToNorwayTime(activeStart)
	if err != nil {
		return err
	}

	// Ensure that the end time is after start time
	if !activeEndTime.After(activeStartTime) {
		return utils.Render(c, http.StatusBadRequest, components.ErrorText(
			errorActiveTimeElementID, "Sluttidspunktet må være etter starttidspunktet"))
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
	// Get the article ID from the URL
	articleID, err := articles.GetSmpIdFromString(articleURL.String())
	articleURL = articles.GetSmpURLFromID(articleID)

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

			switch err {
			case articles.ErrInvalidArticleID:
				return article, "Ugyldig artikkel ID"
			case articles.ErrInvalidArticleURL:
				return article, "Ugyldig artikkel URL"
			case articles.ErrArticleNotFound:
				return article, "Klarte ikke å finne artikkel data for denne URLen. Sjekk at URLen er riktig eller prøv igjen senere"
			default:
				return article, err.Error()
			}
		}

		// Add the article to the DB
		articles.AddArticle(db, &tempArticle)

		article = &tempArticle
	}

	return article, ""
}

const errorArticleElementID = "error-article"

// Adds an article to a quiz in the database.
func (aah *AdminApiHandler) addArticleToQuiz(c echo.Context) error {
	// Set HX-Reswap header to "outerHTML" for error response
	c.Response().Header().Set("HX-Reswap", "outerHTML")

	// Get the quiz ID
	quiz_id, err := uuid.Parse(c.QueryParam(queryParamQuizID))
	if err != nil {
		return utils.Render(c, http.StatusBadRequest, components.ErrorText(errorArticleElementID, errorInvalidQuizID))
	}

	// Get the article URL
	articleURL := c.FormValue(dashboard_pages.QuizArticleURL)
	tempURL, err := url.Parse(articleURL)
	if err != nil && err == sql.ErrNoRows {
		return utils.Render(c, http.StatusBadRequest, components.ErrorText(errorArticleElementID, "Ugyldig artikkel URL"))
	}

	// Ensure the article is in the database
	article, errText := conditionallyAddArticle(aah.sharedData.DB, tempURL, &quiz_id)
	if errText != "" {
		return utils.Render(c, http.StatusBadRequest, components.ErrorText(errorArticleElementID, errText))
	}

	// Add the article to the quiz
	err = articles.AddArticleToQuizByID(aah.sharedData.DB, &article.ID.UUID, &quiz_id)
	if err != nil {
		return err
	}

	// Set HX-Reswap header to "outerHTML" for error response
	c.Response().Header().Set("HX-Reswap", "beforeend")

	return utils.Render(c, http.StatusOK, dashboard_components.ArticleListItem(
		article.ArticleURL.String(), article.Title, article.ID.UUID.String(), quiz_id.String()))
}

// Deletes an article from a quiz in the database.
func (aah *AdminApiHandler) deleteArticle(c echo.Context) error {
	// Get the quiz ID
	quiz_id, err := uuid.Parse(c.QueryParam(queryParamQuizID))
	if err != nil {
		return utils.Render(c, http.StatusBadRequest, components.ErrorText(errorArticleElementID, fmt.Sprintf("Kunne ikke slette artikkel: %s", errorInvalidQuizID)))
	}

	// Get the article ID
	article_id, err := uuid.Parse(c.QueryParam("article-id"))
	if err != nil {
		return utils.Render(c, http.StatusBadRequest, components.ErrorText(errorArticleElementID, "Kunne ikke slette artikkel: Ugyldig eller manglende artikkel ID"))
	}

	// Remove the article from the quiz
	err = articles.DeleteArticleFromQuiz(aah.sharedData.DB, c.Request().Context(), &quiz_id, &article_id)
	if err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

// Rearrange the sequence of questions for a quiz.
func (aah *AdminApiHandler) rearrangeQuestions(c echo.Context) error {
	// Get the quiz ID
	quizID, err := uuid.Parse(c.QueryParam(queryParamQuizID))
	if err != nil {
		return utils.Render(c, http.StatusBadRequest, components.ErrorText("error-question-list", errorInvalidQuizID))
	}

	// Get the map of question IDs and their new arrangement number.
	// This is a map in JSON body.
	questionsList := make(map[int]uuid.UUID)
	err = c.Bind(&questionsList)
	if err != nil {
		return utils.Render(c, http.StatusBadRequest, components.ErrorText("error-question-list", "Ugyldig liste med spørsmål. Det må være en map med rekkefølge og IDer"))
	}

	// Rearrange the questions
	err = questions.RearrangeQuestions(aah.sharedData.DB, c.Request().Context(), quizID, questionsList)
	if err != nil {
		if err == questions.ErrNonSequentialQuestions {
			return utils.Render(c, http.StatusBadRequest, components.ErrorText("error-question-list", "Spørsmålene må ha en sekvensiell rekkefølge"))
		}
		return err
	}

	return c.NoContent(http.StatusOK)
}

// Edit a question with the given data.
// If the question ID is not found, a new question will be created.
// If the question ID is found, the question will be updated.
func (aah *AdminApiHandler) editQuestion(c echo.Context) error {
	// Set HX-Reswap header to "outerHTML" for error response
	c.Response().Header().Set("HX-Reswap", "outerHTML")

	// Get the quiz ID
	quizID, err := uuid.Parse(c.QueryParam(queryParamQuizID))
	if err != nil {
		return utils.Render(c, http.StatusBadRequest, components.ErrorText(errorQuestionElementID, errorInvalidQuizID))
	}

	// Get the data from the form
	articleURLString := c.FormValue(dashboard_components.QuestionArticleURL)
	questionText := c.FormValue(dashboard_components.QuestionText)
	imageURLString := c.FormValue(imageURLInput)
	questionPoints := c.FormValue(dashboard_components.QuestionPoints)
	timeLimit := c.FormValue(dashboard_components.QuestionTimeLimit)

	// Parse the data and validate
	points, articleURL, imageURL, time, errorText := questions.ParseAndValidateQuestionData(questionText, questionPoints, articleURLString, imageURLString, timeLimit)
	if errorText != "" {
		return utils.Render(c, http.StatusBadRequest, components.ErrorText(errorQuestionElementID, errorText))
	}

	articleId := uuid.NullUUID{}

	// Only add article to DB if it is not empty.
	// I.e. allow for no article, but not invalid article.
	if articleURLString != "" {
		tempArticle, err := articles.GetArticleByURL(aah.sharedData.DB, articleURL)
		if err == sql.ErrNoRows {
			return utils.Render(c, http.StatusBadRequest, components.ErrorText(
				errorQuestionElementID, "Fant ikke artikkelen. Sjekk at URLen er riktig eller prøv igjen senere"))
		} else if err != nil {
			return err
		}

		articles.AddArticle(aah.sharedData.DB, tempArticle)
		articleId = tempArticle.ID
	}

	// Get the question ID.
	questionID, err := uuid.Parse(c.QueryParam(queryParamQuestionID))
	if err != nil {
		return utils.Render(c, http.StatusBadRequest, components.ErrorText(errorQuestionElementID, errorInvalidQuestionID))
	}

	// Get the alternatives from the form.
	// There are always 4 alternatives, but some may be empty.
	var alternatives [4]questions.PartialAlternative
	for index := range 4 {
		// The alternatives match the arrangement number (1, 2, 3, 4, etc.) not the index number.
		alternativeText := c.FormValue(fmt.Sprintf("question-alternative-%d", index+1))
		isCorrect := c.FormValue(fmt.Sprintf("question-alternative-%d-is-correct", index+1))
		alternatives[index] = questions.PartialAlternative{Text: alternativeText, IsCorrect: isCorrect == "on"}
	}

	var hasFile = true
	// Get the image file if it exists and upload it
	imageFile, err := c.FormFile(imageFileInput)
	if err != nil {
		if err == http.ErrMissingFile {
			hasFile = false
		} else {
			log.Println(err)
			return utils.Render(c, http.StatusBadRequest, components.ErrorText(errorQuestionElementID, errorFetchingImage))
		}
	}

	if hasFile {
		// Upload image from File
		imageName, err := aah.uploadImage(c, imageFile)
		if err != nil {
			log.Println(err)
			return utils.Render(c, http.StatusBadRequest, components.ErrorText(errorImageElementID, errorUploadImage))
		}

		// Set the image URL for the quiz
		imageURLString = aah.sharedData.Bucket.EndpointURL().String() + bucketImageURL + imageName

		// Parse the image URL
		imageURL, err = url.Parse(imageURLString)
		if err != nil {
			log.Println(err)
			return utils.Render(c, http.StatusBadRequest, components.ErrorText(errorImageElementID, errorUploadImage))
		}
	} else if c.FormValue(imageURLInput) != "" {
		// Upload image from URL
		imageName, err := aah.uploadImageFromURL(c, *imageURL)
		if err != nil {
			log.Println(err)
			return utils.Render(c, http.StatusBadRequest, components.ErrorText(errorQuestionElementID, errorUploadImage))
		}

		// Set the image URL for the quiz
		imageURLString = aah.sharedData.Bucket.EndpointURL().String() + bucketImageURL + imageName

		// Parse the image URL
		imageURL, err = url.Parse(imageURLString)
		if err != nil {
			log.Println(err)
			return utils.Render(c, http.StatusBadRequest, components.ErrorText(errorImageElementID, errorUploadImage))
		}
	}

	// Create a new question object
	questionForm := questions.QuestionForm{
		ID:               questionID,
		Text:             questionText,
		ImageURL:         imageURL,
		ArticleID:        &articleId,
		QuizID:           &quizID,
		Points:           points,
		TimeLimitSeconds: time,
		Alternatives:     alternatives,
	}
	question, errorText := questions.CreateQuestionFromForm(questionForm)
	if errorText != "" {
		return utils.Render(c, http.StatusBadRequest, components.ErrorText(errorQuestionElementID, errorText))
	}

	// Get the question by ID from the database.
	tempQuestion, err := questions.GetQuestionByID(aah.sharedData.DB, questionID)

	// If the question doesn't exist in the database.
	if err == sql.ErrNoRows {
		// Save the question to the database.
		err = questions.AddNewQuestion(aah.sharedData.DB, c.Request().Context(), &question)
		if err != nil {
			return err
		}

		// Set HX-Reswap header to "beforeend" for success response
		c.Response().Header().Set("HX-Reswap", "beforeend")
	} else if tempQuestion.ID == questionID {
		// If the question ID is found, update the question.
		question.ID = questionID
		err = questions.UpdateQuestion(aah.sharedData.DB, c.Request().Context(), &question)

		if err != nil {
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
		return utils.Render(c, http.StatusBadRequest, components.ErrorText(errorQuestionElementID,
			fmt.Sprintf("Kunne ikke slette spørsmål: %s", errorInvalidQuestionID)))
	}

	// Delete the question from the database
	err = questions.DeleteQuestionByID(aah.sharedData.DB, c.Request().Context(), &questionID)
	if err != nil {
		if err == questions.ErrLastQuestion {
			return utils.Render(c, http.StatusBadRequest, components.ErrorText(errorQuestionElementID,
				fmt.Sprintf("Kunne ikke slette spørsmål: %s", "det er siste spørsmålet i en publisert quiz. Lag et nytt spørsmål eller upubliser quizen for å slette spørsmålet.")))
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
		return utils.Render(c, http.StatusBadRequest, components.ErrorText(errorImageElementID, errorInvalidQuestionID))
	}

	// Get the new image URL
	image := c.FormValue(imageURLInput)
	imageURL, err := url.Parse(image)
	if err != nil {
		return utils.Render(c, http.StatusBadRequest, components.ErrorText(errorImageElementID, "Ugyldig bilde URL"))
	}

	// Upload the image to the bucket from URL
	imageName, err := aah.uploadImageFromURL(c, *imageURL)
	if err != nil {
		return utils.Render(c, http.StatusBadRequest, components.ErrorText(errorImageElementID, errorUploadImage))
	}

	imageURL, err = url.Parse(aah.sharedData.Bucket.EndpointURL().String() + bucketImageURL + imageName)
	if err != nil {
		return utils.Render(c, http.StatusBadRequest, components.ErrorText(errorImageElementID, errorUploadImage))
	}

	// Set the image URL for the question
	err = questions.SetImageByQuestionID(aah.sharedData.DB, &questionID, imageURL)
	if err != nil {
		if err == questions.ErrNoImageUpdated {
			return utils.Render(c, http.StatusBadRequest, components.ErrorText(errorImageElementID,
				"Quiz bilde kunne ikke bli oppdatert. Sjekk at informasjonen er korrekt eller prøv igjen senere"))
		}

		return err
	}

	return utils.Render(c, http.StatusOK, dashboard_components.EditImageInput(
		fmt.Sprintf(editQuestionImageURL, questionID), fmt.Sprintf(editQuestionImageFile, questionID),
		fmt.Sprintf(imageSuggestionsQuestion, questionID), imageURL, true, "", dashboard_components.IdPrefixQuestion))
}

// Upload a new image for a question.
func (aah *AdminApiHandler) uploadQuestionImage(c echo.Context) error {
	// Get the quiz ID
	questionID, err := uuid.Parse(c.QueryParam(queryParamQuestionID))
	if err != nil {
		return utils.Render(c, http.StatusBadRequest, components.ErrorText(errorImageElementID, errorInvalidQuestionID))

	}

	// Get the image file
	image, err := c.FormFile(imageFileInput)
	if err != nil {
		log.Println(err)
		return utils.Render(c, http.StatusBadRequest, components.ErrorText(errorImageElementID, errorFetchingImage))
	}

	// Upload the image to the bucket
	imageName, err := aah.uploadImage(c, image)
	if err != nil {
		log.Println(err)
		return utils.Render(c, http.StatusBadRequest, components.ErrorText(errorImageElementID, errorUploadImage))
	}

	// Set the image URL for the question
	imageURL := aah.sharedData.Bucket.EndpointURL().String() + bucketImageURL + imageName

	// Parse the image URL
	imageAsURL, err := url.Parse(imageURL)
	if err != nil {
		log.Println(err)
		return utils.Render(c, http.StatusBadRequest, components.ErrorText(errorImageElementID, errorUploadImage))
	}

	err = questions.SetImageByQuestionID(aah.sharedData.DB, &questionID, imageAsURL)
	if err != nil {
		log.Println(err)
		return err
	}

	return utils.Render(c, http.StatusOK, dashboard_components.EditImageInput(
		fmt.Sprintf(editQuestionImageURL, questionID), fmt.Sprintf(editQuestionImageFile, questionID),
		fmt.Sprintf(imageSuggestionsQuestion, questionID), imageAsURL, true, "", dashboard_components.IdPrefixQuestion))
}

// Delete the image for a question in the database.
func (aah *AdminApiHandler) deleteQuestionImage(c echo.Context) error {
	// Get the question ID
	questionID, err := uuid.Parse(c.QueryParam(queryParamQuestionID))
	if err != nil {
		return utils.Render(c, http.StatusOK, dashboard_components.EditImageInput(
			fmt.Sprintf(editQuestionImageURL, questionID), fmt.Sprintf(editQuestionImageFile, questionID),
			fmt.Sprintf(imageSuggestionsQuestion, questionID), &url.URL{}, true, errorInvalidQuestionID, dashboard_components.IdPrefixQuestion))
	}

	// Remove the image URL from the question
	err = questions.RemoveImageByQuestionID(aah.sharedData.DB, &questionID)
	if err != nil {
		if err == questions.ErrNoImageRemoved {
			return utils.Render(c, http.StatusOK, dashboard_components.EditImageInput(
				fmt.Sprintf(editQuestionImageURL, questionID), fmt.Sprintf(editQuestionImageFile, questionID),
				fmt.Sprintf(imageSuggestionsQuestion, questionID), &url.URL{}, true, "Spørsmål bilde kunne ikke bli fjernet. Prøv igjen senere", dashboard_components.IdPrefixQuestion))
		}

		return err
	}

	return utils.Render(c, http.StatusOK, dashboard_components.EditImageInput(
		fmt.Sprintf(editQuestionImageURL, questionID), fmt.Sprintf(editQuestionImageFile, questionID),
		fmt.Sprintf(imageSuggestionsQuestion, questionID), &url.URL{}, true, "", dashboard_components.IdPrefixQuestion))
}

// Uploads an image to the bucket from a form and returns the name of the image.
// If the image cannot be uploaded, an error is returned.
func (aah *AdminApiHandler) uploadImage(c echo.Context, image *multipart.FileHeader) (string, error) {
	file, err := image.Open()
	if err != nil {
		return "", err
	}
	defer file.Close()

	// get file size
	size := image.Size
	// get file name
	filename := image.Filename
	// get file content type
	contentType := image.Header.Get("Content-Type")
	// get file extension
	extension := strings.Split(filename, ".")[1]
	// generate random file name
	randomName := fmt.Sprintf("%s.%s", uuid.New().String(), extension)
	// create a new reader
	reader := io.LimitReader(file, size)
	// upload file to minio
	err = aah.uploadImageToBucket(c, reader, "images", randomName, size, contentType)
	if err != nil {
		return "", err
	}
	return randomName, nil
}

// Uploads an image to the bucket and returns an error if the image cannot be uploaded.
func (aah *AdminApiHandler) uploadImageToBucket(c echo.Context, imageData io.Reader, bucketName string, fileName string, fileSize int64, contentType string) error {
	// get minio client
	bucket := aah.sharedData.Bucket
	// create a new reader
	reader := io.LimitReader(imageData, fileSize)
	// upload file to minio
	_, err := bucket.PutObject(c.Request().Context(), bucketName, fileName, reader, fileSize, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

// Uploads an image from a URL to the bucket and returns the name of the image.
// If the image cannot be uploaded, an error is returned.
func (aah *AdminApiHandler) uploadImageFromURL(c echo.Context, imageURL url.URL) (string, error) {
	// get image from provided URL
	resp, err := http.Get(imageURL.String())
	if err != nil {
		return "", err
	}

	if resp.ContentLength <= 0 {
		//retry up to 5 times if content length is 0 or -1
		for i := 0; i < 5; i++ {
			resp, err = http.Get(imageURL.String())
			if err != nil {
				return "", err
			}
			if resp.ContentLength > 0 {
				break
			}
		}

		if resp.ContentLength <= 0 {
			return "", errors.New("could not fetch image")
		}
	}

	defer resp.Body.Close()

	fileType := resp.Header.Get("Content-Type")

	randomName := fmt.Sprintf("%s.%s", uuid.New().String(), strings.Split(fileType, "/")[1])

	err = aah.uploadImageToBucket(c, resp.Body, "images", randomName, resp.ContentLength, resp.Header.Get("Content-Type"))
	if err != nil {
		return "", err
	}

	return randomName, nil
}

// Randomizes the order of the alternatives for a question visually.
func (aah *AdminApiHandler) randomizeAlternatives(c echo.Context) error {
	// Get the alternatives
	var alternatives []questions.Alternative
	for index := range 4 {
		// The alternatives match the arrangement number (1, 2, 3, 4, etc.) not the index number.
		alternativeText := c.FormValue(fmt.Sprintf("question-alternative-%d", index+1))
		isCorrect := c.FormValue(fmt.Sprintf("question-alternative-%d-is-correct", index+1))
		alternatives = append(alternatives, questions.Alternative{Text: alternativeText, IsCorrect: isCorrect == "on"})
	}

	// Shuffle the alternatives
	rand.Shuffle(len(alternatives), func(i, j int) {
		alternatives[i], alternatives[j] = alternatives[j], alternatives[i]
	})

	// Return the "alternatives" table.
	return utils.Render(c, http.StatusOK, dashboard_components.QuestionAlternativesInput(alternatives))

}

// Add the given word to the username tables.
func (aah *AdminApiHandler) addUsername(c echo.Context) error {
	word := c.FormValue("username-word")
	table := c.QueryParam("table-id")

	if word == "" || table == "" {
		return c.NoContent(http.StatusBadRequest)
	}

	err := usernames.AddWordToTable(aah.sharedData.DB, word, table)
	if err != nil {
		return err
	}
	c.Response().Header().Set("HX-Refresh", "true")
	return c.NoContent(http.StatusOK)
}

// Delete the given words from the username tables.
func (aah *AdminApiHandler) deleteUsername(c echo.Context) error {
	//Get array of words to delete from JSON body
	var words []string
	err := c.Bind(&words)

	if err != nil {
		return err
	}

	usernames.DeleteWordsFromTable(aah.sharedData.DB, c.Request().Context(), words)

	return c.NoContent(http.StatusOK)
}

// Edit the username tables with the given data.
// Takes in a map of of tables, and the old and new words in the tables.
func (aah *AdminApiHandler) editUsername(c echo.Context) error {
	wordList := make(map[string][]usernames.OldNew)
	err := c.Bind(&wordList)
	if err != nil {
		return err
	}

	err = usernames.UpdateAdjectives(aah.sharedData.DB, wordList[usernames.AdjectiveTable])
	if err != nil {
		return err
	}

	err = usernames.UpdateNouns(aah.sharedData.DB, wordList[usernames.NounTable])
	if err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

// Get image suggestions for a quiz
func (aah *AdminApiHandler) imageSuggestionsQuiz(c echo.Context) error {
	// Get quiz ID
	quizId, err := uuid.Parse(c.QueryParam(queryParamQuizID))
	if err != nil {
		return utils.Render(c, http.StatusBadRequest, components.ErrorText(errorImageElementID, errorInvalidQuizID))
	}

	// Get articles in the quiz
	arts, err := articles.GetArticlesByQuizID(aah.sharedData.DB, quizId)
	if err != nil {
		return err
	}

	// Get the images from the articles
	images, err := articles.GetImagesFromArticles(arts)
	if err != nil {
		return err
	}

	return utils.Render(c, http.StatusOK, dashboard_components.ArticleImages(images, "quiz"))
}

// Get image suggestions for a question
func (aah *AdminApiHandler) imageSuggestionsQuestion(c echo.Context) error {
	// Get the article URL from form
	articleURL, err := url.Parse(c.FormValue("article-url"))
	if err != nil {
		return utils.Render(c, http.StatusBadRequest, components.ErrorText(errorImageElementID, "Ugyldig artikkel URL"))
	}

	if articleURL.String() == "" {
		return utils.Render(c, http.StatusOK, dashboard_components.ArticleImages([]url.URL{}, "question"))
	}

	// Get the article from the URL
	article, err := articles.GetArticleByURL(aah.sharedData.DB, articleURL)
	if err != nil {
		return err
	}

	// Get the images from the articles
	images, err := articles.GetImagesFromArticles(&[]articles.Article{*article})
	if err != nil {
		return err
	}

	return utils.Render(c, http.StatusOK, dashboard_components.ArticleImages(images, "question"))
}

// Get the username tables and render the page.
func (aah *AdminApiHandler) getUsernamePages(c echo.Context) error {
	adjPage, err := strconv.Atoi(c.QueryParam("adj"))
	// If the page number is not a number, set it to 1.
	if err != nil {
		adjPage = 1
	}
	nounPage, err := strconv.Atoi(c.QueryParam("noun"))
	// If the page number is not a number, set it to 1.
	if err != nil {
		nounPage = 1
	}

	pages, err := strconv.Atoi(c.QueryParam("rows-per-page"))
	// Sets to 25 if between a certain range.
	if err != nil || pages < 5 || pages > 255 {
		pages = 25
	}

	// Get the search query
	search := c.FormValue("search")
	if search == "" {
		search = c.QueryParam("search")
	}

	uai, err := usernames.GetUsernameAdminInfo(aah.sharedData.DB, adjPage, nounPage, pages, search)
	if err != nil {
		return utils.Render(c, http.StatusBadRequest, components.ErrorText("error-username", "Noe gikk galt under henting av data. Prøv igjen senere. Hvis problemet vedvarer, kontakt administrator."))
	}

	// Creates a relative path with the queryparams to update the url of
	// the client webpage.
	var relativePath url.URL
	relativeQuery := c.Request().URL.Query()
	requestUrl := c.Request().URL
	if search != "" {
		relativeQuery.Set("search", search)
	}
	requestUrl.RawQuery = relativeQuery.Encode()
	relativePath.RawQuery = relativeQuery.Encode()
	c.Response().Header().Set("HX-Replace-Url", relativePath.String())

	return utils.Render(c, http.StatusOK, user_admin.UsernameTables(uai, requestUrl))
}

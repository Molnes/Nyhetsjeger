package questions

import (
	"context"
	"database/sql"
	"errors"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Molnes/Nyhetsjeger/internal/models/articles"
	data_handling "github.com/Molnes/Nyhetsjeger/internal/utils/data"
	"github.com/google/uuid"
)

var ErrNoQuestionDeleted = errors.New("questions: no question deleted")
var ErrNoQuestionUpdated = errors.New("questions: no question updated")
var ErrNoImageRemoved = errors.New("questions: no image removed")
var ErrNoImageUpdated = errors.New("questions: no image updated")

type Question struct {
	ID               uuid.UUID
	Text             string
	ImageURL         url.URL
	Arrangement      uint
	Article          articles.Article // The article this question is based on.
	QuizID           uuid.UUID
	TimeLimitSeconds uint
	Points           uint
	Alternatives     []Alternative
}

type Alternative struct {
	ID            uuid.UUID
	Text          string
	IsCorrect     bool
	QuestionID    uuid.UUID
	chosenBy      uint
	PercentChosen float64
}

func (q *Question) IsAnswerCorrect(answerID uuid.UUID) bool {
	isCorrect := false
	for _, a := range q.Alternatives {
		if a.ID == answerID {
			isCorrect = a.IsCorrect
			break
		}
	}
	return isCorrect
}

// Subtract the given duration from the time limit of this question.
// If the duration is greater than the time limit, the time limit is set to 0.
func (q *Question) SubtractFromTimeLimit(duration time.Duration) {
	diffSeconds := duration.Seconds()
	if diffSeconds > float64(q.TimeLimitSeconds) {
		q.TimeLimitSeconds = 0
	} else {
		q.TimeLimitSeconds -= uint(diffSeconds)
	}
}

// Initializes the percentage of times each alternative has been chosen.
func (q *Question) initPercentChosen() {
	total := 0
	for _, a := range q.Alternatives {
		total += int(a.chosenBy)
	}
	if total != 0 {
		for i, a := range q.Alternatives {
			q.Alternatives[i].PercentChosen = float64(a.chosenBy) / float64(total)
		}
	}
}

// Returns the ID of first question in the given quiz.
func GetFirstQuestionID(db *sql.DB, quizID uuid.UUID) (uuid.UUID, error) {
	var id uuid.UUID
	row := db.QueryRow(
		`SELECT id
    	FROM questions
    	WHERE
      	quiz_id = $1 AND arrangement = 1`,
		quizID)
	err := row.Scan(&id)

	return id, err
}

// Returns all questions for a given quiz.
// Includes the article for each question.
// Includes the alternatives for each question.
func GetQuestionsByQuizID(db *sql.DB, id uuid.UUID) (*[]Question, error) {
	rows, err := db.Query(
		`SELECT
				q.id, q.question, q.image_url, q.arrangement, q.article_id, q.quiz_id, q.time_limit_seconds, q.points
			FROM
				questions q
			LEFT JOIN
				answer_alternatives a ON q.id = a.question_id
			WHERE
				quiz_id = $1
			GROUP BY
				q.id
			ORDER BY
				q.arrangement ASC`,
		id)
	if err != nil {
		return nil, err
	}
	defer rows.Close() // Close the rows when the function returns

	return scanQuestionsFromFullRows(db, rows)
}

// Convert a row from the database to a Question.
func scanQuestionFromFullRow(db *sql.DB, row *sql.Row) (*Question, error) {
	var q Question
	var articleID uuid.UUID
	var imageURL sql.NullString
	err := row.Scan(
		&q.ID, &q.Text, &imageURL, &q.Arrangement, &articleID, &q.QuizID, &q.Points,
	)
	if err != nil {
		return nil, err
	}

	// Set image URL
	tempURL, err := data_handling.ConvertNullStringToURL(&imageURL)
	if err != nil {
		return nil, err
	}
	q.ImageURL = *tempURL

	// Add the article to the question
	article, _ := articles.GetArticleByID(db, articleID)
	if article != nil {
		q.Article = *article
	}

	// Add the alternatives to the question
	altneratives, err := GetAlternativesByQuestionID(db, q.ID)
	if err != nil {
		return nil, err
	}
	q.Alternatives = *altneratives

	return &q, nil
}

// Converts a row from the database to a list of questions
func scanQuestionsFromFullRows(db *sql.DB, rows *sql.Rows) (*[]Question, error) {
	questions := []Question{}

	for rows.Next() {
		var q Question
		var articleID uuid.UUID
		var imageURL sql.NullString
		err := rows.Scan(
			&q.ID, &q.Text, &imageURL, &q.Arrangement, &articleID, &q.QuizID, &q.TimeLimitSeconds, &q.Points,
		)
		if err != nil {
			return nil, err
		}

		// Set image URL
		tempURL, err := data_handling.ConvertNullStringToURL(&imageURL)
		if err != nil {
			return nil, err
		}
		q.ImageURL = *tempURL

		// Add the article to the question
		article, _ := articles.GetArticleByID(db, articleID)
		if article != nil {
			q.Article = *article
		}

		// Add the alternatives to the question
		altneratives, err := GetAlternativesByQuestionID(db, q.ID)
		if err != nil {
			return nil, err
		}
		q.Alternatives = *altneratives

		// Add the question to the list of questions
		questions = append(questions, q)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &questions, nil
}

func IsCorrectAnswer(db *sql.DB, questionID uuid.UUID, answerID uuid.UUID) (bool, error) {
	row := db.QueryRow(
		`SELECT correct
		FROM answer_alternatives
		WHERE
			id = $1 AND question_id = $2`,
		answerID, questionID)

	var isCorrect bool
	err := row.Scan(&isCorrect)
	if err != nil {
		return false, err
	}
	return isCorrect, nil
}

// Get specific question by ID.
// Includes the article for the question.
// Includes the answer altneratives for the question.
func GetQuestionByID(db *sql.DB, id uuid.UUID) (*Question, error) {
	var q Question
	var imageUrlString sql.NullString
	row := db.QueryRow(
		`
		SELECT
			q.id, question, q.image_url AS quiz_image, q.arrangement, q.article_id, q.quiz_id, q.time_limit_seconds, q.points
		FROM
			questions q
		WHERE
			q.id = $1;
		`, id)
	err := row.Scan(
		&q.ID, &q.Text, &imageUrlString, &q.Arrangement, &q.Article.ID, &q.QuizID, &q.TimeLimitSeconds, &q.Points,
	)
	if err != nil {
		return nil, err
	}

	// Parse the image URL
	imageUrl, err := data_handling.ConvertNullStringToURL(&imageUrlString)
	if err != nil {
		return nil, err
	} else {
		q.ImageURL = *imageUrl
	}

	// Get the article if it exists
	if q.Article.ID.Valid {
		article, err := articles.GetArticleByID(db, q.Article.ID.UUID)
		if err != nil {
			return nil, err
		}
		q.Article = *article
	}

	answerRows, err := db.Query(
		`SELECT
			aa.id, aa.text, aa.correct, aa.question_id, COUNT(ua.chosen_answer_alternative_id)
		FROM
			answer_alternatives aa
		LEFT JOIN
			user_answers ua ON aa.id = ua.chosen_answer_alternative_id
		WHERE
			aa.question_id = $1
		GROUP BY
			aa.id;
		`, id)
	if err != nil {
		return nil, err
	}
	defer answerRows.Close()

	if err != nil {
		return nil, err
	}

	var alternatives []Alternative
	for answerRows.Next() {
		var a Alternative
		err := answerRows.Scan(
			&a.ID, &a.Text, &a.IsCorrect, &a.QuestionID, &a.chosenBy,
		)
		if err != nil {
			return nil, err
		}
		alternatives = append(alternatives, a)
	}
	if len(alternatives) == 0 {
		// not a "rows not found", it is a server error - questions must have alternatives
		return nil, errors.New("no alternatives found for question with id: " + id.String())
	}
	q.Alternatives = alternatives
	q.initPercentChosen()
	return &q, nil
}

// Get all alternatives for a given question.
func GetAlternativesByQuestionID(db *sql.DB, id uuid.UUID) (*[]Alternative, error) {
	rows, err := db.Query(
		`SELECT
      id, text, correct, question_id
    FROM
      answer_alternatives
    WHERE
      question_id = $1`,
		id)
	if err != nil {
		return nil, err
	}
	defer rows.Close() // Close the rows when the function returns

	return scanAlternativesFromFullRows(rows)
}

// Convert a row from the database to a list of alternatives
func scanAlternativesFromFullRows(rows *sql.Rows) (*[]Alternative, error) {
	alternatives := []Alternative{}

	for rows.Next() {
		var a Alternative
		err := rows.Scan(
			&a.ID, &a.Text, &a.IsCorrect, &a.QuestionID,
		)
		if err != nil {
			return nil, err
		}

		// Add the alternative to the list of alternatives
		alternatives = append(alternatives, a)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &alternatives, nil
}

// Return alternative for a given question.
func GetAlternativeByID(db *sql.DB, id uuid.UUID) (*Alternative, error) {
	row := db.QueryRow(
		`SELECT
				id, text, correct, question_id
			FROM
				answer_alternatives
			WHERE
				id = $1`,
		id)

	return scanAlternativeFromFullRow(row)
}

// Convert a row from the database to an Alternative.
func scanAlternativeFromFullRow(row *sql.Row) (*Alternative, error) {
	var a Alternative
	err := row.Scan(
		&a.ID,
		&a.Text,
		&a.IsCorrect,
		&a.QuestionID,
	)

	if err != nil {
		return nil, err
	}
	return &a, nil
}

type QuestionForm struct {
	ID                  uuid.UUID
	Text                string
	ImageURL            *url.URL
	Article             *articles.Article
	QuizID              *uuid.UUID
	Points              uint
	TimeLimitSeconds    uint
	Alternative1        string
	Alternative2        string
	Alternative3        string
	Alternative4        string
	CorrectAnswerNumber string
}

// Parse and validate question data.
// If the data is invalid, return an error describing the problem.
// Returns the points, article URL, image URL, time limit and error text.
func ParseAndValidateQuestionData(questionText string, questionPoints string, articleURLString string, imageURL string, timeLimit string) (uint, *url.URL, *url.URL, uint, string) {
	if strings.TrimSpace(questionText) == "" {
		return 0, nil, nil, 0, "Spørsmålsteksten kan ikke være tom"
	}

	points, err := strconv.ParseInt(questionPoints, 10, 64)
	if err != nil {
		return 0, nil, nil, 0, "Klarte ikke å tolke poeng verdien"
	}
	if points < 0 {
		return 0, nil, nil, 0, "Poeng kan ikke være negative"
	}

	time, err := strconv.ParseInt(timeLimit, 10, 64)
	if err != nil {
		return 0, nil, nil, 0, "Klarte ikke å tolke tidsbegrensningen"
	}
	if time < 0 {
		return 0, nil, nil, 0, "Tidsbegrensningen kan ikke være negativ"
	}

	articleURL, err := url.Parse(articleURLString)
	if err != nil {
		return 0, nil, nil, 0, "Klarte ikke å tolke artikkel URL"
	}

	image, err := url.Parse(imageURL)
	if err != nil {
		return 0, nil, nil, 0, "Klarte ikke å tolke bilde URL"
	}

	return uint(points), articleURL, image, uint(time), ""
}

// Create a question object from a form.
// Returns the created question and an error message.
func CreateQuestionFromForm(form QuestionForm) (Question, string) {
	question := Question{
		ID:               form.ID,
		Text:             form.Text,
		ImageURL:         *form.ImageURL,
		Article:          *form.Article,
		QuizID:           *form.QuizID,
		Points:           form.Points,
		TimeLimitSeconds: form.TimeLimitSeconds,
		Alternatives:     []Alternative{},
	}

	hasCorrectAlternative := false

	// Only add alternatives that are not empty
	for index, alt := range []string{form.Alternative1, form.Alternative2, form.Alternative3, form.Alternative4} {
		// Do not count empty white space as an alternative
		if strings.TrimSpace(alt) != "" {
			question.Alternatives = append(question.Alternatives, Alternative{
				ID:        uuid.New(),
				Text:      alt,
				IsCorrect: form.CorrectAnswerNumber == strconv.Itoa(index+1),
			})

			if form.CorrectAnswerNumber == strconv.Itoa(index+1) {
				hasCorrectAlternative = true
			}
		}
	}

	// Check that there is a correct alternative
	if !hasCorrectAlternative {
		return question, "Spørsmålet har ingen korrekt svar alternativ"
	}

	// Check that there are two to four alternatives
	if len(question.Alternatives) < 2 {
		return question, "Spørsmålet må ha minst 2 svar alternativer. Tomme alternativer teller ikke"
	}
	if len(question.Alternatives) > 4 {
		return question, "Spørsmålet kan ha maks 4 svar alternativer"
	}

	return question, ""
}

// Add a new question to the database.
// Adds the question alternatives to the database.
// Returns the ID of the new question.
func AddNewQuestion(db *sql.DB, ctx context.Context, question *Question) error {
	// Start a transaction
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	// Insert the question into the database
	_, err = tx.Exec(
		`INSERT INTO questions (id, question, image_url, article_id, quiz_id, points, time_limit_seconds)
		VALUES ($1, $2, $3, $4, $5, $6, $7);`,
		question.ID, question.Text, question.ImageURL.String(), question.Article.ID, question.QuizID, question.Points, question.TimeLimitSeconds,
	)

	if err != nil {
		tx.Rollback()
		return err
	}

	// Insert the alternatives into the database
	for _, a := range question.Alternatives {
		_, err = tx.Exec(
			`INSERT INTO answer_alternatives (id, text, correct, question_id)
			VALUES ($1, $2, $3, $4);`,
			a.ID, a.Text, a.IsCorrect, question.ID,
		)

		if err != nil {
			tx.Rollback()
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

// Update a question in the database.
func UpdateQuestion(db *sql.DB, ctx context.Context, question *Question) error {
	// Start a transaction
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	result, err := tx.Exec(
		`UPDATE questions
		SET question = $1, image_url = $2, article_id = $3, quiz_id = $4, points = $5, time_limit_seconds = $6
		WHERE id = $7;`,
		question.Text, question.ImageURL.String(), question.Article.ID, question.QuizID, question.Points, question.TimeLimitSeconds, question.ID,
	)

	if err != nil {
		tx.Rollback()
		return err
	}

	if rowsAffected, _ := result.RowsAffected(); rowsAffected == 0 {
		tx.Rollback()
		return ErrNoQuestionUpdated
	}

	// Delete old alternatives
	_, err = tx.Exec(
		`DELETE FROM answer_alternatives
		WHERE question_id = $1;`,
		question.ID,
	)

	if err != nil {
		tx.Rollback()
		return err
	}

	// Insert the alternatives into the database
	for _, a := range question.Alternatives {
		_, err = tx.Exec(
			`INSERT INTO answer_alternatives (id, text, correct, question_id)
			VALUES ($1, $2, $3, $4);`,
			a.ID, a.Text, a.IsCorrect, question.ID,
		)

		if err != nil {
			tx.Rollback()
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

// Delete a question from the database.
// Deletes the question alternatives from the database.
func DeleteQuestionByID(db *sql.DB, ctx context.Context, id *uuid.UUID) error {
	// Start a transaction
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	// Get the question's quiz ID and arrangement number
	var quizID uuid.UUID
	var arrangement uint
	row := tx.QueryRow(
		`SELECT quiz_id, arrangement
		FROM questions
		WHERE id = $1;`,
		id,
	)
	err = row.Scan(&quizID, &arrangement)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Delete the question from the database
	result, err := tx.Exec(
		`DELETE FROM questions
		WHERE id = $1;`,
		id,
	)
	if err != nil {
		tx.Rollback()
		return err
	}

	if rowsAffected, _ := result.RowsAffected(); rowsAffected == 0 {
		tx.Rollback()
		return ErrNoQuestionDeleted
	}

	// Delete the alternatives from the database
	_, err = tx.Exec(
		`DELETE FROM answer_alternatives
		WHERE question_id = $1;`,
		id,
	)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Update the arrangement numbers of the remaining questions
	// (decrement all questions' arrangement with higher arrangement number)
	_, err = tx.Exec(
		`UPDATE questions
		SET arrangement = arrangement - 1
		WHERE quiz_id = $1 AND arrangement > $2;`,
		quizID, arrangement,
	)
	if err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

// Update the image URL for a question by question ID.
func SetImageByQuestionID(db *sql.DB, id *uuid.UUID, imageURL *url.URL) error {
	result, err := db.Exec(
		`UPDATE
			questions
		SET
			image_url = $1
		WHERE
			id = $2;`,
		imageURL.String(), id,
	)

	if rowsAffected, _ := result.RowsAffected(); rowsAffected == 0 {
		return ErrNoImageUpdated
	}

	return err
}

// Remove the image URL for a question by question ID.
func RemoveImageByQuestionID(db *sql.DB, id *uuid.UUID) error {
	result, err := db.Exec(
		`UPDATE
			questions
		SET
			image_url = NULL
		WHERE
			id = $1;`,
		id,
	)

	if rowsAffected, _ := result.RowsAffected(); rowsAffected == 0 {
		return ErrNoImageRemoved
	}

	return err
}

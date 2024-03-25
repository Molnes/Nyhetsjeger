package questions

import (
	"context"
	"database/sql"
	"errors"
	"net/url"
	"strconv"
	"time"

	"github.com/Molnes/Nyhetsjeger/internal/models/articles"
	data_handling "github.com/Molnes/Nyhetsjeger/internal/utils/data"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

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
				q.id, q.question, q.image_url, q.arrangement, q.article_id, q.quiz_id, q.time_limit_seconds, q.points,
				array_agg(a.id)
			FROM
				questions q
			LEFT JOIN
				answer_alternatives a ON q.id = a.question_id
			WHERE
				quiz_id = $1
			GROUP BY
				q.id`,
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
	var alternativeIDs []uuid.UUID
	var imageURL sql.NullString
	err := row.Scan(
		&q.ID, &q.Text, &imageURL, &q.Arrangement, &articleID, &q.QuizID, &q.Points,
		pq.Array(&alternativeIDs),
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
	for _, id := range alternativeIDs {
		alternative, _ := GetAlternativeByID(db, id)
		if alternative != nil {
			q.Alternatives = append(q.Alternatives, *alternative)
		}
	}

	return &q, nil
}

// Converts a row from the database to a list of questions
func scanQuestionsFromFullRows(db *sql.DB, rows *sql.Rows) (*[]Question, error) {
	questions := []Question{}

	for rows.Next() {
		var q Question
		var articleID uuid.UUID
		var alternativeIDs []uuid.UUID
		var imageURL sql.NullString
		err := rows.Scan(
			&q.ID, &q.Text, &imageURL, &q.Arrangement, &articleID, &q.QuizID, &q.TimeLimitSeconds, &q.Points,
			pq.Array(&alternativeIDs),
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
		for _, id := range alternativeIDs {
			alternative, _ := GetAlternativeByID(db, id)
			if alternative != nil {
				q.Alternatives = append(q.Alternatives, *alternative)
			}
		}

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
func GetQuestionByID(db *sql.DB, id uuid.UUID) (*Question, error) {
	var q Question
	var imageUrlString sql.NullString
	row := db.QueryRow(
		`
		SELECT
		id, question, image_url, arrangement, article_id, quiz_id, time_limit_seconds, points
		FROM questions
		WHERE id = $1;
		`, id)
	err := row.Scan(
		&q.ID, &q.Text, &imageUrlString, &q.Arrangement, &q.Article.ID, &q.QuizID, &q.TimeLimitSeconds, &q.Points,
	)
	if err != nil {
		return nil, err
	}

	imageUrl, err := data_handling.ConvertNullStringToURL(&imageUrlString)
	if err != nil {
		return nil, err
	} else {
		q.ImageURL = *imageUrl
	}

	answerRows, err := db.Query(
		`SELECT aa.id, aa.text, aa.correct, aa.question_id, COUNT(ua.chosen_answer_alternative_id)
		FROM answer_alternatives aa
		LEFT JOIN user_answers ua ON aa.id = ua.chosen_answer_alternative_id
		WHERE aa.question_id = $1
		GROUP BY aa.id;
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
	Alternative1        string
	Alternative2        string
	Alternative3        string
	Alternative4        string
	CorrectAnswerNumber string
}

// Parse and validate question data.
// If the data is invalid, return an error describing the problem.
// Returns the points, article URL, image URL and error text.
func ParseAndValidateQuestionData(questionText string, questionPoints string, articleURLString string, imageURL string) (uint, *url.URL, *url.URL, string) {
	if questionText == "" {
		return 0, nil, nil, "Missing question text"
	}

	points, err := strconv.ParseInt(questionPoints, 10, 64)
	if err != nil {
		return 0, nil, nil, "Failed to parse points"
	}
	if points < 0 {
		return 0, nil, nil, "Points must be positive"
	}

	articleURL, err := url.Parse(articleURLString)
	if err != nil {
		return 0, nil, nil, "Failed to parse article URL"
	}

	image, err := url.Parse(imageURL)
	if err != nil {
		return 0, nil, nil, "Failed to parse image URL"
	}

	return uint(points), articleURL, image, ""
}

// Create a question object from a form.
// Returns the created question and an error message.
func CreateQuestionFromForm(form QuestionForm) (Question, string) {
	question := Question{
		ID:           form.ID,
		Text:         form.Text,
		ImageURL:     *form.ImageURL,
		Article:      *form.Article,
		QuizID:       *form.QuizID,
		Points:       form.Points,
		Alternatives: []Alternative{},
	}

	hasCorrectAlternative := false

	// Only add alternatives that are not empty
	for index, alt := range []string{form.Alternative1, form.Alternative2, form.Alternative3, form.Alternative4} {
		if alt != "" {
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
		return question, "Question has no correct alternative"
	}

	// Check that there are two to four alternatives
	if len(question.Alternatives) < 2 {
		return question, "There must be at least two alternatives"
	}
	if len(question.Alternatives) > 4 {
		return question, "There can be at most four alternatives"
	}

	return question, ""
}

// Add a new question to the database.
// Adds the question alternatives to the database.
// Returns the ID of the new question.
func AddNewQuestion(db *sql.DB, question Question) error {
	// Start a transaction
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	// Insert the question into the database
	_, err = tx.Exec(
		`INSERT INTO questions (id, question, image_url, article_id, quiz_id, points)
		VALUES ($1, $2, $3, $4, $5, $6);`,
		question.ID, question.Text, question.ImageURL.String(), question.Article.ID, question.QuizID, question.Points,
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
func UpdateQuestion(db *sql.DB, question *Question) error {
	// Start a transaction
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	_, err = tx.Exec(
		`UPDATE questions
		SET question = $1, image_url = $2, article_id = $3, quiz_id = $4, points = $5
		WHERE id = $6;`,
		question.Text, question.ImageURL.String(), question.Article.ID, question.QuizID, question.Points, question.ID,
	)

	if err != nil {
		tx.Rollback()
		return err
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
func DeleteQuestionByID(db *sql.DB, id uuid.UUID) error {
	// Start a transaction
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	// Delete the question from the database
	_, err = tx.Exec(
		`DELETE FROM questions
		WHERE id = $1;`,
		id,
	)

	if err != nil {
		tx.Rollback()
		return err
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

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

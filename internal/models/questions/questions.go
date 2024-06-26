package questions

import (
	"context"
	"database/sql"
	"errors"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Molnes/Nyhetsjeger/internal/models/ai"
	data_handling "github.com/Molnes/Nyhetsjeger/internal/utils/data"
	"github.com/google/uuid"
)

var ErrNoQuestionDeleted = errors.New("questions: no question deleted")
var ErrNoQuestionUpdated = errors.New("questions: no question updated")
var ErrNoImageRemoved = errors.New("questions: no image removed")
var ErrNoImageUpdated = errors.New("questions: no image updated")
var ErrLastQuestion = errors.New("questions: cannot delete the last question in a published quiz")
var ErrNonSequentialQuestions = errors.New("questions: question arrangement is not sequential")

type Question struct {
	ID               uuid.UUID
	Text             string
	ImageURL         url.URL
	Arrangement      uint
	ArticleID        uuid.NullUUID
	QuizID           uuid.UUID
	TimeLimitSeconds uint
	Points           uint
	Alternatives     []Alternative
}

type Alternative struct {
	ID            uuid.UUID
	Text          string
	IsCorrect     bool
	Arrangement   uint
	QuestionID    uuid.UUID
	chosenBy      uint
	PercentChosen float64
}

type PartialAlternative struct {
	ID        uuid.UUID
	Text      string
	IsCorrect bool
}

// Create a default question with an ID, quiz ID and 100 points.
// All other fields are "empty".
func GetDefaultQuestion(quizId uuid.UUID) Question {
	return Question{
		ID:           uuid.New(),
		Text:         "",
		ImageURL:     url.URL{},
		ArticleID:    uuid.NullUUID{},
		QuizID:       quizId,
		Points:       100,
		Alternatives: []Alternative{},
	}
}

// Convert an AI question to a normal question
func ConvertAiQuestionToQuestion(quizId uuid.UUID, articleId uuid.UUID, aiQuestion ai.Question) *Question {
	newQuestion := Question{
		ID:               uuid.New(),
		QuizID:           quizId,
		Text:             aiQuestion.Question,
		Points:           100,
		TimeLimitSeconds: 30,
		ArticleID: uuid.NullUUID{
			UUID:  articleId,
			Valid: true,
		},
		ImageURL:     url.URL{},
		Alternatives: []Alternative{},
	}

	for index, alt := range aiQuestion.Alternatives {
		newQuestion.Alternatives = append(newQuestion.Alternatives, Alternative{
			ID:            uuid.New(),
			QuestionID:    newQuestion.ID,
			Text:          alt.AlternativeText,
			IsCorrect:     alt.Correct,
			PercentChosen: 0,
			Arrangement:   uint(index + 1),
		})
	}

	return &newQuestion
}

// Checks if the provided answer id is correct for this question.
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

// Gets the text of the alternative with given ID within this quesiton. If answerId is not found in this question, default empty string ("") is returned.
func (q *Question) GetAnswerTextById(answerID uuid.UUID) string {
	var text string
	for _, a := range q.Alternatives {
		if a.ID == answerID {
			text = a.Text
			break
		}
	}
	return text
}

// Returns the seconds left after substracting given duration from the question's time limit.
func (q *Question) GetRemainingTimeSeconds(duration time.Duration) uint {
	diffSeconds := duration.Seconds()
	var timeLeft uint
	if diffSeconds < float64(q.TimeLimitSeconds) {
		timeLeft = q.TimeLimitSeconds - uint(diffSeconds)
	}
	return timeLeft
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
// Includes the alternatives for each question.
func GetQuestionsByQuizID(db *sql.DB, id *uuid.UUID) (*[]Question, error) {
	rows, err := db.Query(
		`SELECT
				q.id, q.question, q.image_url, q.arrangement, q.article_id, q.quiz_id, q.time_limit_seconds, q.points
			FROM
				questions q
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

// Gets n-th question in a given quiz, indexing from 1.
func GetNthQuestionByQuizId(db *sql.DB, quizId uuid.UUID, questionNumber uint) (*Question, error) {
	row := db.QueryRow(
		`SELECT
				q.id, q.question, q.image_url, q.arrangement, q.article_id, q.quiz_id, q.time_limit_seconds, q.points
			FROM
				questions q
			WHERE
				quiz_id = $1
			GROUP BY
				q.id
			ORDER BY
				q.arrangement ASC
			LIMIT 1
			OFFSET $2;`, quizId, questionNumber-1)

	return scanQuestionFromFullRow(db, row)
}

// Gets the next question's id in a quiz, next meaning after the question with id provided.
// If no more questions, sql.ErrNoRows  will be returned
func GetNextQuestionInQuizByQuestionId(db *sql.DB, questionId uuid.UUID) (uuid.UUID, error) {
	row := db.QueryRow(
		`WITH questions_in_quiz AS (
			SELECT q.id, q.arrangement,
			ROW_NUMBER () OVER (ORDER BY q.arrangement) 
			FROM questions q, 
			  (SELECT quiz_id as id
				FROM questions 
				WHERE id = $1
			  ) quiz 
			WHERE q.quiz_id = quiz.id
		  ) 
		  SELECT questions_in_quiz.id 
		  FROM questions_in_quiz, 
			( SELECT row_number 
			  FROM questions_in_quiz 
			  WHERE id = $1
			) current 
		  WHERE questions_in_quiz.row_number = current.row_number + 1;`, questionId)
	var id uuid.UUID
	err := row.Scan(&id)
	return id, err
}

// Convert a row from the database to a Question.
// Also adds the alternatives to the question. (separate query)
func scanQuestionFromFullRow(db *sql.DB, row *sql.Row) (*Question, error) {
	var q Question
	var imageURL sql.NullString
	err := row.Scan(
		&q.ID, &q.Text, &imageURL, &q.Arrangement, &q.ArticleID, &q.QuizID, &q.TimeLimitSeconds, &q.Points,
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

	// Add the alternatives to the question
	altneratives, err := GetAlternativesByQuestionID(db, q.ID)
	if err != nil {
		return nil, err
	}
	q.Alternatives = *altneratives

	return &q, nil
}

// Converts a row from the database to a list of questions
// Also adds the alternatives to the questions. (separate query)
func scanQuestionsFromFullRows(db *sql.DB, rows *sql.Rows) (*[]Question, error) {
	questions := []Question{}

	for rows.Next() {
		var q Question
		var imageURL sql.NullString
		err := rows.Scan(
			&q.ID, &q.Text, &imageURL, &q.Arrangement, &q.ArticleID, &q.QuizID, &q.TimeLimitSeconds, &q.Points,
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
// Includes the answer altneratives for the question.
func GetQuestionByID(db *sql.DB, id uuid.UUID) (*Question, error) {
	var q Question
	var imageUrlString sql.NullString
	row := db.QueryRow(
		`
		SELECT
			id, question, image_url AS quiz_image, arrangement, article_id, quiz_id, time_limit_seconds, points
		FROM
			questions
		WHERE
			id = $1;
		`, id)
	err := row.Scan(
		&q.ID, &q.Text, &imageUrlString, &q.Arrangement, &q.ArticleID, &q.QuizID, &q.TimeLimitSeconds, &q.Points,
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

	answerRows, err := db.Query(
		`SELECT
			aa.id, aa.text, aa.correct, aa.arrangement, aa.question_id, COUNT(ua.chosen_answer_alternative_id)
		FROM
			answer_alternatives aa
		LEFT JOIN
			user_answers ua ON aa.id = ua.chosen_answer_alternative_id
		WHERE
			aa.question_id = $1
		GROUP BY
			aa.id
		ORDER BY aa.arrangement;`, id)
	if err != nil {
		return nil, err
	}
	defer answerRows.Close()

	var alternatives []Alternative
	for answerRows.Next() {
		var a Alternative
		err := answerRows.Scan(
			&a.ID, &a.Text, &a.IsCorrect, &a.Arrangement, &a.QuestionID, &a.chosenBy,
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
      id, text, correct, arrangement, question_id
    FROM
      answer_alternatives
    WHERE
      question_id = $1
	ORDER BY arrangement `, id)
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
			&a.ID, &a.Text, &a.IsCorrect, &a.Arrangement, &a.QuestionID,
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
	ID               uuid.UUID
	Text             string
	ImageURL         *url.URL
	ArticleID        *uuid.NullUUID
	QuizID           *uuid.UUID
	Points           uint
	TimeLimitSeconds uint
	Alternatives     [4]PartialAlternative
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
		ArticleID:        *form.ArticleID,
		QuizID:           *form.QuizID,
		Points:           form.Points,
		TimeLimitSeconds: form.TimeLimitSeconds,
		Alternatives:     []Alternative{},
	}

	hasCorrectAlternative := false

	// Only add alternatives that are not empty
	for i, alt := range form.Alternatives {
		// Do not count empty white space as an alternative
		if (strings.TrimSpace(alt.Text) != "") || (strings.TrimSpace(alt.Text) == "" && alt.ID != uuid.Nil) {
			question.Alternatives = append(question.Alternatives, Alternative{
				ID:          alt.ID,
				Text:        alt.Text,
				IsCorrect:   alt.IsCorrect,
				Arrangement: uint(i + 1),
			})

			if alt.IsCorrect {
				hasCorrectAlternative = true
			}
		}
	}

	// Check that there is a correct alternative
	if !hasCorrectAlternative {
		return question, "Spørsmålet har ingen korrekt svaralternativ"
	}

	// Check that there are two to four alternatives
	if len(question.Alternatives) < 2 {
		return question, "Spørsmålet må ha minst 2 svaralternativer. Tomme alternativer teller ikke"
	}
	if len(question.Alternatives) > 4 {
		return question, "Spørsmålet kan ha maks 4 svaralternativer"
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
		question.ID, question.Text, question.ImageURL.String(), question.ArticleID, question.QuizID, question.Points, question.TimeLimitSeconds,
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
		question.Text, question.ImageURL.String(), question.ArticleID, question.QuizID, question.Points, question.TimeLimitSeconds, question.ID,
	)

	if err != nil {
		tx.Rollback()
		return err
	}

	if rowsAffected, _ := result.RowsAffected(); rowsAffected == 0 {
		tx.Rollback()
		return ErrNoQuestionUpdated
	}

	// Update the alternatives
	for _, a := range question.Alternatives {
		// If the alternative ID exists, but the text is empty, delete the alternative
		if a.Text == "" {
			_, err := tx.Exec(
				`DELETE FROM answer_alternatives
				WHERE id = $1;`,
				a.ID,
			)
			if err != nil {
				tx.Rollback()
				return err
			}
		} else {
			// Either insert or update, based on if it exists already
			_, err := tx.Exec(
				`INSERT INTO answer_alternatives (id, text, correct, question_id)
				VALUES ($1, $2, $3, $4)
				ON CONFLICT (id)
				DO UPDATE SET text = $2, correct = $3, arrangement = $5
				WHERE answer_alternatives.id = $1;`,
				a.ID, a.Text, a.IsCorrect, question.ID, a.Arrangement,
			)

			if err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

// Delete a question from the database.
// Deletes the question alternatives from the database.
// If the question is the last one in a published quiz, return an error.
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

	res := tx.QueryRow(
		`SELECT COUNT(qe.id)
		FROM questions qe
		JOIN quizzes qu ON qu.id = qe.quiz_id
		WHERE qu.id = $1
		AND qu.published = true`,
		quizID,
	)

	// Number of questions in published quiz
	var questionsInPublishedQuiz int

	if res.Scan(&questionsInPublishedQuiz); questionsInPublishedQuiz == 1 {
		tx.Rollback()
		return ErrLastQuestion
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

// Rearrange the question arrangement in a quiz.
func RearrangeQuestions(db *sql.DB, ctx context.Context, quizID uuid.UUID, questionArrangement map[int]uuid.UUID) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	// Get a list of the keys in the map.
	arrangements := []int{}
	for k := range questionArrangement {
		arrangements = append(arrangements, k)
	}

	// Check that the arrangement of all questions in a quiz is perfectly sequential.
	numberOfSequence := 0
	for index := range arrangements {
		if _, ok := questionArrangement[index+1]; ok {
			numberOfSequence++
		} else {
			break
		}
	}

	if numberOfSequence != len(questionArrangement) {
		tx.Rollback()
		return ErrNonSequentialQuestions
	}

	// Update the arrangement of the questions in the quiz.
	for _, arrangement := range arrangements {
		_, err = tx.Exec(
			`UPDATE questions
			SET arrangement = $1
			WHERE id = $2 AND quiz_id = $3;`,
			arrangement,
			questionArrangement[arrangement],
			quizID)
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

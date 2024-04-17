package quizzes

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"net/url"
	"time"

	data_handling "github.com/Molnes/Nyhetsjeger/internal/utils/data"
	"github.com/google/uuid"
)

var ErrNoQuestions = errors.New("quizzes: no questions in quiz")
var ErrNonSequentialQuestions = errors.New("quizzes: question arrangement is not sequential")

// Quiz represents a quiz in the database.
type Quiz struct {
	ID             uuid.UUID
	Title          string
	ImageURL       url.URL
	ActiveFrom     time.Time
	ActiveTo       time.Time
	CreatedAt      time.Time
	LastModifiedAt time.Time
	Published      bool
	IsDeleted      bool
}

// PartialQuiz represents a quiz in the database with fewer fields.
type PartialQuiz struct {
	ID             uuid.UUID
	Title          string
	ImageURL       url.URL
	ActiveFrom     time.Time
	ActiveTo       time.Time
	Published      bool
	QuestionNumber uint
	MaxScore       uint
}

// Create a default quiz.
// This function is used to create a new quiz with default values.
func CreateDefaultQuiz() Quiz {
	tn := time.Now().Local()
	_, week := tn.ISOWeek()

	return Quiz{
		ID:    uuid.New(),
		Title: fmt.Sprintf("Quiz: Uke %d", week),
		ImageURL: url.URL{
			Scheme: "https",
			Host:   "upload.wikimedia.org",
			Path:   "/wikipedia/commons/5/59/Question_mark_choice.jpg",
		},
		ActiveFrom:     time.Now(),
		ActiveTo:       time.Now().Add(24 * 7 * time.Hour),
		CreatedAt:      time.Now(),
		LastModifiedAt: time.Now(),
		Published:      false,
		IsDeleted:      false,
	}
}

// Retrieves a quiz from the database by its ID.
func GetQuizByID(db *sql.DB, id uuid.UUID) (*Quiz, error) {
	row := db.QueryRow(
		`SELECT
			id, title, image_url, active_from, active_to, created_at, last_modified_at, published, is_deleted
    FROM
			quizzes
		WHERE
			id = $1 AND
			is_deleted = false`,
		id)

	quiz, err := scanQuizFromFullRow(row)
	if err != nil {
		return nil, err
	}

	return quiz, nil
}

// Update the image URL for a quiz by its ID.
func UpdateImageByQuizID(db *sql.DB, id uuid.UUID, imageURL url.URL) error {
	_, err := db.Exec(
		`UPDATE quizzes
		SET image_url = $1
		WHERE id = $2`,
		imageURL.String(),
		id)
	return err
}

// Remove the image URL for a quiz by its ID.
func RemoveImageByQuizID(db *sql.DB, id uuid.UUID) error {
	_, err := db.Exec(
		`UPDATE quizzes
		SET image_url = NULL
		WHERE id = $1`,
		id)
	return err
}

// Update the title for a quiz by its ID.
func UpdateTitleByQuizID(db *sql.DB, id uuid.UUID, title string) error {
	_, err := db.Exec(
		`UPDATE quizzes
		SET title = $1
		WHERE id = $2`,
		title,
		id)
	return err
}

// Get all quizzes in the database.
func GetQuizzes(db *sql.DB) ([]Quiz, error) {
	rows, err := db.Query(
		`SELECT
			id, title, image_url, active_from, active_to, created_at, last_modified_at, published, is_deleted
    FROM
			quizzes
		WHERE
			is_deleted = false
		ORDER BY
			active_from DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanQuizzesFromFullRows(rows)
}

// Get all quizzes in the database by the user ID that are not finished.
func GetIsQuizzesByUserIDAndNotFinished(db *sql.DB, userID uuid.UUID) ([]Quiz, error) {
	rows, err := db.Query(
		`SELECT q.id, q.title, q.image_url, q.active_from, q.active_to
FROM quizzes q
LEFT JOIN user_quizzes cq ON q.id = cq.quiz_id AND cq.user_id = $1 
WHERE (cq.user_id IS NULL OR cq.is_completed = 'f')
AND q.active_from <= NOW() 
AND q.active_to >= NOW() 
AND q.published = 't'
AND q.is_deleted = 'f'; `, userID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	quizzes := []Quiz{}
	var imageURL sql.NullString
	for rows.Next() {
		var quiz Quiz
		err := rows.Scan(
			&quiz.ID,
			&quiz.Title,
			&imageURL,
			&quiz.ActiveFrom,
			&quiz.ActiveTo,
		)
		if err != nil {
			return nil, err
		}

		tempURL, err := data_handling.ConvertNullStringToURL(&imageURL)
		if err != nil {
			return nil, err
		}
		quiz.ImageURL = *tempURL
		quizzes = append(quizzes, quiz)
	}
	return quizzes, nil
}

// Get all quizzes in the database by the user ID that are not finished and not active.
func GetIsQuizzesByUserIDNotFinishedAndNotActive(db *sql.DB, userID uuid.UUID) ([]Quiz, error) {
	rows, err := db.Query(
		`SELECT q.id, q.title, q.image_url, q.active_from, q.active_to
FROM quizzes q
LEFT JOIN user_quizzes cq ON q.id = cq.quiz_id AND cq.user_id = $1
WHERE cq.user_id IS NULL
AND q.active_from > NOW()   
OR q.active_to < NOW()
AND q.published = 't'
AND q.is_deleted = 'f'; `, userID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	quizzes := []Quiz{}
	var imageURL sql.NullString
	for rows.Next() {
		var quiz Quiz
		err := rows.Scan(
			&quiz.ID,
			&quiz.Title,
			&imageURL,
			&quiz.ActiveFrom,
			&quiz.ActiveTo,
		)
		if err != nil {
			return nil, err
		}

		tempURL, err := data_handling.ConvertNullStringToURL(&imageURL)
		if err != nil {
			return nil, err
		}
		quiz.ImageURL = *tempURL

		quizzes = append(quizzes, quiz)
	}
	return quizzes, nil
}

// Get all quizzes in the database by the user ID that are finished.
func GetIsQuizzesByUserIDAndFinished(db *sql.DB, userID uuid.UUID) ([]Quiz, error) {
	rows, err := db.Query(
		`SELECT q.id, q.title, q.image_url, q.active_from, q.active_to
FROM quizzes q
LEFT JOIN user_quizzes cq ON q.id = cq.quiz_id AND cq.user_id = $1
WHERE cq.user_id IS NOT NULL
AND q.published = 't'
AND cq.is_completed = 't'
; `, userID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	quizzes := []Quiz{}
	var imageURL sql.NullString
	for rows.Next() {
		var quiz Quiz
		err := rows.Scan(
			&quiz.ID,
			&quiz.Title,
			&imageURL,
			&quiz.ActiveFrom,
			&quiz.ActiveTo,
		)
		if err != nil {
			return nil, err
		}

		tempURL, err := data_handling.ConvertNullStringToURL(&imageURL)
		if err != nil {
			return nil, err
		}
		quiz.ImageURL = *tempURL

		quizzes = append(quizzes, quiz)
	}
	return quizzes, nil
}

// Get quizzes that a user has finished or not. Quiz has to be published and not deleted.
func GetQuizzesByUserIDAndFinishedOrNot(db *sql.DB, userID uuid.UUID, isFinished bool) ([]Quiz, error) {

	if isFinished {
		quizzes, err := GetIsQuizzesByUserIDAndFinished(db, userID)
		if err != nil {
			return nil, err
		}
		return quizzes, nil
	} else {
		quizzes, err := GetIsQuizzesByUserIDAndNotFinished(db, userID)
		if err != nil {
			return nil, err
		}
		return quizzes, nil
	}
}

// Get quizzes that a user has finished or not. Quiz has to be published and not deleted. Gets quizzez that are not active.
func GetQuizzesByUserIDAndFinishedOrNotAndNotActive(db *sql.DB, userID uuid.UUID, isFinished bool) ([]Quiz, error) {

	if isFinished {
		quizzes, err := GetIsQuizzesByUserIDAndFinished(db, userID)
		if err != nil {
			return nil, err
		}
		return quizzes, nil
	} else {
		quizzes, err := GetIsQuizzesByUserIDNotFinishedAndNotActive(db, userID)
		if err != nil {
			return nil, err
		}
		return quizzes, nil
	}

}

// Get all the quizzes that are not published and not deleted.
func GetQuizzesByPublishStatus(db *sql.DB, published bool) ([]Quiz, error) {
	rows, err := db.Query(
		`SELECT
			id, title, image_url, active_from, active_to, created_at, last_modified_at, published, is_deleted
		FROM
			quizzes
		WHERE
			published = $1 AND
			is_deleted = false
		ORDER BY
			active_from DESC`,
		published)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Scan the quizzes from the database.
	return scanQuizzesFromFullRows(rows)
}

// Converts a row from the database to a Quiz.
// It expects the row to contain ID, Title, ImageURL, ActiveFrom, ActiveTo, CreatedAt, LastModifiedAt, Published, IsDeleted.
// It will return a Quiz with these values.
func scanQuizFromFullRow(row *sql.Row) (*Quiz, error) {
	var quiz Quiz
	var imageURL sql.NullString
	err := row.Scan(
		&quiz.ID,
		&quiz.Title,
		&imageURL,
		&quiz.ActiveFrom,
		&quiz.ActiveTo,
		&quiz.CreatedAt,
		&quiz.LastModifiedAt,
		&quiz.Published,
		&quiz.IsDeleted,
	)
	if err != nil {
		return nil, err
	}

	// Set image URL
	tempURL, err := data_handling.ConvertNullStringToURL(&imageURL)
	if err != nil {
		return nil, err
	}
	quiz.ImageURL = *tempURL

	return &quiz, err
}

// Converts rows from the database to a list of Quizzes.
// It expects the row to contain ID, Title, ImageURL, ActiveFrom, ActiveTo, CreatedAt, LastModifiedAt, Published, IsDeleted.
// It will return a Quiz with these values.
func scanQuizzesFromFullRows(rows *sql.Rows) ([]Quiz, error) {
	quizzes := []Quiz{}
	for rows.Next() {
		var quiz Quiz
		var imageURL sql.NullString
		err := rows.Scan(
			&quiz.ID,
			&quiz.Title,
			&imageURL,
			&quiz.ActiveFrom,
			&quiz.ActiveTo,
			&quiz.CreatedAt,
			&quiz.LastModifiedAt,
			&quiz.Published,
			&quiz.IsDeleted,
		)
		if err != nil {
			return nil, err
		}

		// Set image URL
		tempURL, err := data_handling.ConvertNullStringToURL(&imageURL)
		if err != nil {
			return nil, err
		}
		quiz.ImageURL = *tempURL

		quizzes = append(quizzes, quiz)
	}
	return quizzes, nil
}

// Create a Quiz in the DB.
func CreateQuiz(db *sql.DB, quiz Quiz) (*uuid.UUID, error) {
	_, err := db.Exec(
		`INSERT INTO quizzes
			(id, title, image_url, active_from, active_to, created_at, last_modified_at, published, is_deleted)
		VALUES
			($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		quiz.ID,
		quiz.Title,
		quiz.ImageURL.String(),
		quiz.ActiveFrom,
		quiz.ActiveTo,
		quiz.CreatedAt,
		quiz.LastModifiedAt,
		quiz.Published,
		quiz.IsDeleted,
	)

	return &quiz.ID, err
}

// Set a Quiz to deleted in the DB by its ID.
func DeleteQuizByID(db *sql.DB, id uuid.UUID) error {
	_, err := db.Exec(
		`UPDATE quizzes
		SET is_deleted = true
		WHERE id = $1`,
		id)
	return err
}

// Update the published status of a quiz by its ID.
func UpdatePublishedStatusByQuizID(db *sql.DB, ctx context.Context, id uuid.UUID, published bool) error {
	// Start a transaction
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	// If quiz is published, but the quiz has no questions, then return an error.
	if published {
		result := tx.QueryRow(
			`SELECT COUNT(*)
			FROM questions q
			WHERE q.quiz_id = $1`,
			id)

		var questionsInQuiz int
		result.Scan(&questionsInQuiz)
		if questionsInQuiz == 0 {
			tx.Rollback()
			return ErrNoQuestions
		}
	}

	_, err = tx.Exec(
		`UPDATE quizzes
		SET published = $1
		WHERE id = $2`,
		published,
		id)

	if err != nil {
		tx.Rollback()
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return err
}

// Retrieves a partial quiz from the database by a quiz ID.
func GetPartialQuizByID(db *sql.DB, quizid uuid.UUID) (*PartialQuiz, error) {
	row := db.QueryRow(
		`SELECT qz.id, qz.title, qz.image_url, qz.active_from, qz.active_to, qz.published, count(q.id), sum(q.points)
		FROM quizzes qz 
		LEFT JOIN questions q ON q.quiz_id = qz.id
		WHERE qz.id = $1 AND qz.is_deleted = false
		GROUP BY qz.id;`, quizid)

	var pq PartialQuiz
	var imageURLStr sql.NullString
	err := row.Scan(
		&pq.ID,
		&pq.Title,
		&imageURLStr,
		&pq.ActiveFrom,
		&pq.ActiveTo,
		&pq.Published,
		&pq.QuestionNumber,
		&pq.MaxScore,
	)
	if err != nil {
		return nil, err
	}

	tempURL, err := data_handling.ConvertNullStringToURL(&imageURLStr)
	if err != nil {
		return nil, err
	}
	pq.ImageURL = *tempURL

	return &pq, nil
}

// Update the quiz's 'active' start time by its ID.
func UpdateActiveStartByQuizID(db *sql.DB, id uuid.UUID, activeStart time.Time) error {
	_, err := db.Exec(
		`UPDATE quizzes
		SET active_from = $1
		WHERE id = $2`,
		activeStart,
		id)
	return err
}

// Update the quiz's 'active' end time by its ID.
func UpdateActiveEndByQuizID(db *sql.DB, id uuid.UUID, activeEnd time.Time) error {
	_, err := db.Exec(
		`UPDATE quizzes
		SET active_to = $1
		WHERE id = $2`,
		activeEnd,
		id)
	return err
}

// Rearrange the question arrangement in a quiz.
func RearrangeQuestions(db *sql.DB, ctx context.Context, quizID uuid.UUID, questionArrangement map[int]uuid.UUID) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	arrangements := []int{}
	// Get a list of the keys in the map.
	for k := range questionArrangement {
		arrangements = append(arrangements, k)
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

	// Check that the arrangement of all questions in a quiz is perfectly sequential.
	numberOfSequence := 0
	for key := range arrangements {
		if _, ok := questionArrangement[key]; ok {
			numberOfSequence++
		} else {
			break
		}
	}

	if numberOfSequence != len(questionArrangement) {
		tx.Rollback()
		return ErrNonSequentialQuestions
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

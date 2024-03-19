package quizzes

import (
	"database/sql"
	"fmt"

	"net/url"
	"time"

	"github.com/Molnes/Nyhetsjeger/internal/models/questions"
	data_handling "github.com/Molnes/Nyhetsjeger/internal/utils/data"
	"github.com/google/uuid"
)

type Quiz struct {
	ID             uuid.UUID
	Title          string
	ImageURL       url.URL
	AvailableFrom  time.Time
	AvailableTo    time.Time
	CreatedAt      time.Time
	LastModifiedAt time.Time
	Published      bool
	Questions      []questions.Question
}

type PartialQuiz struct {
	ID             uuid.UUID
	Title          string
	ImageURL       url.URL
	AvailableFrom  time.Time
	AvailableTo    time.Time
	Published      bool
	QuestionNumber uint
	MaxScore       uint
}

func GetQuiz(quizID uuid.UUID) (Quiz, error) {
	return SampleQuiz, nil
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
			Host:   "unsplash.it",
			Path:   "/200/200",
		},
		AvailableFrom:  time.Now(),
		AvailableTo:    time.Now().Add(24 * 7 * time.Hour),
		CreatedAt:      time.Now(),
		LastModifiedAt: time.Now(),
		Published:      false,
		Questions:      []questions.Question{},
	}
}

var SampleQuiz Quiz = Quiz{
	ID:        uuid.New(),
	Title:     "Eksempel quiz",
	Questions: questions.SampleQuestions,
}

// Retrieves a quiz from the database by its ID.
// Includes the questions for the quiz.
// Includes the articles for each question.
// Includes the alternatives for each question.
func GetQuizByID(db *sql.DB, id uuid.UUID) (*Quiz, error) {
	row := db.QueryRow(
		`SELECT
			id, title, image_url, available_from, available_to, created_at, last_modified_at, published
    FROM
			quizzes
		WHERE
			id = $1`,
		id)

	quiz, err := scanQuizFromFullRow(row)
	if err != nil {
		return nil, err
	}

	tempQuestions, err := questions.GetQuestionsByQuizID(db, id)
	if err != nil {
		return nil, err
	}
	quiz.Questions = *tempQuestions

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
			id, title, image_url, available_from, available_to, created_at, last_modified_at, published 
    FROM
			quizzes`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	quizzes := []Quiz{}
	for rows.Next() {
		var quiz Quiz
		var imageURL sql.NullString
		err := rows.Scan(
			&quiz.ID,
			&quiz.Title,
			&imageURL,
			&quiz.AvailableFrom,
			&quiz.AvailableTo,
			&quiz.CreatedAt,
			&quiz.LastModifiedAt,
			&quiz.Published,
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

func GetNonPublishedQuizzes(db *sql.DB) ([]Quiz, error) {
	return GetQuizzes(db)
}

func GetAllPublishedQuizzes(db *sql.DB) ([]Quiz, error) {
	quizzes, err := GetQuizzes(db)
	if err != nil {
		return nil, err
	}
	quizzes = append(quizzes, quizzes...)
	quizzes = append(quizzes, quizzes...)
	return quizzes, nil
}

// Converts a row from the database to a Quiz.
func scanQuizFromFullRow(row *sql.Row) (*Quiz, error) {
	var quiz Quiz
	var imageURL sql.NullString
	err := row.Scan(
		&quiz.ID,
		&quiz.Title,
		&imageURL,
		&quiz.AvailableFrom,
		&quiz.AvailableTo,
		&quiz.CreatedAt,
		&quiz.LastModifiedAt,
		&quiz.Published,
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

	if err == sql.ErrNoRows {
		return nil, err
	}
	return &quiz, err
}

// Create a Quiz in the DB.
func CreateQuiz(db *sql.DB, quiz Quiz) (*uuid.UUID, error) {
	_, err := db.Exec(
		`INSERT INTO quizzes
			(id, title, image_url, available_from, available_to, created_at, last_modified_at, published)
		VALUES
			($1, $2, $3, $4, $5, $6, $7, $8)`,
		quiz.ID,
		quiz.Title,
		quiz.ImageURL.String(),
		quiz.AvailableFrom,
		quiz.AvailableTo,
		quiz.CreatedAt,
		quiz.LastModifiedAt,
		quiz.Published,
	)

	return &quiz.ID, err
}

// Delete a Quiz from the DB by its ID.
func DeleteQuizByID(db *sql.DB, id uuid.UUID) error {
	_, err := db.Exec(
		`DELETE FROM quizzes
		WHERE id = $1`,
		id)
	return err
}

// Update the published status of a quiz by its ID.
func UpdatePublishedStatusByQuizID(db *sql.DB, id uuid.UUID, published bool) error {
	_, err := db.Exec(
		`UPDATE quizzes
		SET published = $1
		WHERE id = $2`,
		published,
		id)
	return err
}

// Retrieves a partial quiz from the database by a quiz ID.
func GetPartialQuizByID(db *sql.DB, quizid uuid.UUID) (*PartialQuiz, error) {
	row := db.QueryRow(
		`SELECT qz.id, qz.title, qz.image_url, qz.available_from, qz.available_to, qz.published, count(q.id), sum(q.points)
		FROM quizzes qz 
		LEFT JOIN questions q ON q.quiz_id = qz.id
		WHERE qz.id = $1
		GROUP BY qz.id;`, quizid)

	var pq PartialQuiz
	var imageURLStr sql.NullString
	err := row.Scan(
		&pq.ID,
		&pq.Title,
		&imageURLStr,
		&pq.AvailableFrom,
		&pq.AvailableTo,
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

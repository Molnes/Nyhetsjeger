package quizzes

import (
	"database/sql"

	"net/url"
	"time"

	"github.com/Molnes/Nyhetsjeger/internal/data/questions"
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

func GetQuiz(quizID uuid.UUID) (Quiz, error) {
	return SampleQuiz, nil
}

// Create a default quiz.
// This function is used to create a new quiz with default values.
func CreateDefaultQuiz() (Quiz, error) {
	return Quiz{
		ID:    uuid.New(),
		Title: "Daglig Quiz: " + time.Now().Format("01 Jan 2006"),
		ImageURL: url.URL{
			Scheme: "https",
			Host:   "unsplash.it",
			Path:   "/200/200",
		},
		AvailableFrom:  time.Now(),
		AvailableTo:    time.Now().Add(24 * time.Hour),
		CreatedAt:      time.Now(),
		LastModifiedAt: time.Now(),
		Published:      false,
		Questions:      []questions.Question{},
	}, nil
}

var SampleQuiz Quiz = Quiz{
	ID:        uuid.New(),
	Title:     "Sample quiz",
	Questions: questions.SampleQuestions,
}

// Retrieves a quiz from the database by its ID.
// Includes the questions for the quiz.
// Includes the articles for each question.
// Includes the alternatives for each question.
func GetFullQuizByID(db *sql.DB, id uuid.UUID) (*Quiz, error) {
	row := db.QueryRow(
		`SELECT
			id, title, image_url, available_from, available_to, created_at, last_modified_at, published
    FROM
			quizzes
		WHERE
			id = $1`,
		id)

	quiz, err := scanQuizFromFullRow(row)

	tempQuestions, err := questions.GetQuestionsByQuizID(db, id)
	quiz.Questions = *tempQuestions

	return quiz, err
}

func GetQuizzes(db *sql.DB) ([]Quiz, error) {
	rows, err := db.Query(
		`SELECT id, title 
        FROM quizzes`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	quizzes := []Quiz{}
	for rows.Next() {
		quiz := Quiz{}
		err := rows.Scan(
			&quiz.ID,
			&quiz.Title,
		)
		if err != nil {
			return nil, err
		}
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
	var imageURL string
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
	tempURL, err := url.Parse(imageURL)
	quiz.ImageURL = *tempURL

	if err == sql.ErrNoRows {
		return nil, err
	}
	return &quiz, err
}

// Create a Quiz in the DB.
func CreateQuiz(db *sql.DB, quiz Quiz) (*uuid.UUID, error) {
	db.QueryRow(
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

	return &quiz.ID, nil
}

// Delete a Quiz from the DB by its ID.
func DeleteQuizByID(db *sql.DB, id uuid.UUID) error {
	_, err := db.Exec(
		`DELETE FROM quizzes
		WHERE id = $1`,
		id)
	return err
}

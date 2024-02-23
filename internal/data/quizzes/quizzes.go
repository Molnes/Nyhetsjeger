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
		Questions:      []questions.Question{},
	}, nil
}

var SampleQuiz Quiz = Quiz{
	ID:        uuid.New(),
	Title:     "Sample quiz",
	Questions: questions.SampleQuestions,
}

func GetQuizByID(db *sql.DB, id uuid.UUID) (*Quiz, error) {
	row := db.QueryRow(
		`SELECT id, title 
        FROM quizzes
		WHERE id = $1`,
		id)
	return scanQuizzesromFullRow(row)
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

func scanQuizzesromFullRow(row *sql.Row) (*Quiz, error) {
	quiz := Quiz{}
	err := row.Scan(
		&quiz.ID,
		&quiz.Title,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &quiz, err
}

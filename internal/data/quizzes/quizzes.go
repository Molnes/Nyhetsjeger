package quizzes

import (
	"database/sql"

	"github.com/Molnes/Nyhetsjeger/internal/data/questions"
	"github.com/google/uuid"
)

type Quiz struct {
	ID        uuid.UUID
	Title     string
	Questions []questions.Question
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

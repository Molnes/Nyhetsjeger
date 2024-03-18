package user_quiz

import (
	"database/sql"
	"errors"
	"time"

	"github.com/Molnes/Nyhetsjeger/internal/models/questions"
	"github.com/google/uuid"
)

var ErrNoSuchQuiz = errors.New("user_quiz: no such quiz")
var ErrNoMoreQuestions = errors.New("user_quiz: no more unanswered questions in quiz")
var ErrNoSuchAnswer = errors.New("user_quiz: no such answer")

func getNextUnansweredQuestion(db *sql.DB, userID uuid.UUID, quizID uuid.UUID) (*questions.Question, error) {
	row := db.QueryRow(
		`SELECT id
		FROM questions
		WHERE quiz_id = $1
		AND id NOT IN (
			SELECT question_id
			FROM user_answers
			WHERE user_id = $2
		)
		ORDER BY arrangement
		LIMIT 1;`, quizID, userID)

	var id uuid.UUID
	err := row.Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNoMoreQuestions
		}
		return nil, err
	}
	return questions.GetQuestionByID(db, id)
}

func startQuestion(db *sql.DB, userId uuid.UUID, questionId uuid.UUID) error {
	_, err := db.Exec(
		`INSERT INTO user_answers
		(user_id, question_id, question_presented_at)
		VALUES ($1, $2, $3)`, userId, questionId, time.Now())
	return err
}

// Returns the next question in the quiz for the user and saves the time it was presented.
//
// May return:
//
// ErrNoMoreQuestions if there are no more unanswered questions for the user in the quiz.
// ErrNoSuchQuiz if the quiz does not exist.
func StartNextQuestion(db *sql.DB, userID uuid.UUID, quizID uuid.UUID) (*questions.Question, error) {
	question, err := getNextUnansweredQuestion(db, userID, quizID)
	if err != nil {
		return nil, err
	}
	err = startQuestion(db, userID, question.ID)
	if err != nil {
		return nil, err
	}
	return question, nil
}

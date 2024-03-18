package users

import (
	"database/sql"
	"errors"

	"github.com/google/uuid"
)

type UserQuizSummary struct {
	QuizID            uuid.UUID
	QuizTitle         string
	MaxScore          uint
	AchievedScore     uint
	AnsweredQuestions []AnsweredQuestion
}

type AnsweredQuestion struct {
	ChosenAlternativeID   uuid.UUID
	QuestionText          string
	ChosenAlternativeText string
	IsCorrect             bool
	PointsAwarded         uint
}

var ErrNoSuchQuiz = errors.New("quiz_summary: no such quiz")
var ErrQuizNotCompleted = errors.New("quiz_summary: quiz not completed")

func GetQuizSummary(db *sql.DB, userID uuid.UUID, quizID uuid.UUID) (*UserQuizSummary, error) {
	var summary UserQuizSummary

	var questionNumber uint

	quizRow := db.QueryRow(
		`SELECT quizzes.id, title, count(questions.id) as num_questions, sum(questions.points) as max_score
		FROM quizzes, questions
		WHERE quizzes.id = questions.quiz_id
		AND quizzes.id = $1
		GROUP BY quizzes.id, title;`, quizID)

	err := quizRow.Scan(&summary.QuizID, &summary.QuizTitle, &questionNumber, &summary.MaxScore)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNoSuchQuiz
		}
	}

	answeredQuestions, err := getAnsweredQuestions(db, userID, quizID)
	if err != nil {
		return nil, err
	}

	summary.AnsweredQuestions = answeredQuestions

	return &summary, nil
}

func getAnsweredQuestions(db *sql.DB, userID uuid.UUID, quizID uuid.UUID) ([]AnsweredQuestion, error) {
	var answeredQuestions []AnsweredQuestion

	rows, err := db.Query(
		`SELECT user_answers.question_id, points_awarded,
	questions.question, answer_alternatives.text, answer_alternatives.id, answer_alternatives.correct
	FROM user_answers LEFT JOIN answer_alternatives ON user_answers.chosen_answer_alternative_id = answer_alternatives.id
	JOIN questions ON user_answers.question_id = questions.id
	JOIN quizzes ON questions.quiz_id = quizzes.id
	JOIN users ON user_answers.user_id = users.id
	WHERE quizzes.id = $1 AND users.id = $2;`, quizID, userID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {

	}

	return answeredQuestions, nil
}

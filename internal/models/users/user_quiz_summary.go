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
	QuestionID            uuid.UUID
	QuestionText          string
	ChosenAlternativeID   uuid.UUID
	ChosenAlternativeText string
	IsCorrect             bool
	PointsAwarded         int
}

func (aq *AnsweredQuestion) IsComplete() bool {
	return aq.ChosenAlternativeID != uuid.Nil
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

	if len(answeredQuestions) < int(questionNumber) {
		return nil, ErrQuizNotCompleted
	}

	return &summary, nil
}

func getAnsweredQuestions(db *sql.DB, userID uuid.UUID, quizID uuid.UUID) ([]AnsweredQuestion, error) {
	rows, err := db.Query(
		`SELECT questions.id, questions.question, answer_alternatives.id, answer_alternatives.text,
		answer_alternatives.correct, points_awarded
		FROM user_answers
		LEFT JOIN answer_alternatives ON user_answers.chosen_answer_alternative_id = answer_alternatives.id
		JOIN questions ON user_answers.question_id = questions.id
		JOIN quizzes ON questions.quiz_id = quizzes.id
		JOIN users ON user_answers.user_id = users.id
	WHERE quizzes.id = $1 AND users.id = $2;`, quizID, userID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var answeredQuestions []AnsweredQuestion
	for rows.Next() {
		var aq AnsweredQuestion
		var nullableALternativeText sql.NullString
		var nullablePointsAwarded sql.NullInt64
		var nullableIsCorrect sql.NullBool

		err := rows.Scan(
			&aq.QuestionID,
			&aq.QuestionText,
			&aq.ChosenAlternativeID,
			&nullableALternativeText,
			&nullableIsCorrect,
			&nullablePointsAwarded,
		)
		if err != nil {
			return nil, err
		}

		if nullableALternativeText.Valid {
			aq.ChosenAlternativeText = nullableALternativeText.String
		}
		if nullableIsCorrect.Valid {
			aq.IsCorrect = nullableIsCorrect.Bool
		}
		if nullablePointsAwarded.Valid {
			aq.PointsAwarded = int(nullablePointsAwarded.Int64)
		}

		answeredQuestions = append(answeredQuestions, aq)
	}

	return answeredQuestions, nil
}

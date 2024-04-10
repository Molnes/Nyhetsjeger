package user_quiz_summary

import (
	"database/sql"
	"errors"
	"time"

	"github.com/Molnes/Nyhetsjeger/internal/models/articles"
	"github.com/google/uuid"
)

type UserQuizSummary struct {
	QuizID            uuid.UUID
	QuizTitle         string
	QuizActiveTo      time.Time
	MaxScore          uint
	AchievedScore     uint
	AnsweredQuestions []AnsweredQuestion
	HasArticlesToShow bool
}

type AnsweredQuestion struct {
	QuestionID            uuid.UUID
	QuestionText          string
	MaxPoints             uint
	ChosenAlternativeID   uuid.UUID
	ChosenAlternativeText string
	IsCorrect             bool
	PointsAwarded         uint
}

var ErrNoSuchQuiz = errors.New("quiz_summary: no such quiz")
var ErrQuizNotCompleted = errors.New("quiz_summary: quiz not completed")

func GetQuizSummary(db *sql.DB, userID uuid.UUID, quizID uuid.UUID) (*UserQuizSummary, error) {
	quizRow := db.QueryRow(
		`SELECT qz.id, COALESCE(uq.total_points_awarded, 0), COALESCE(uq.is_completed, false),
		qz.title, qz.active_to, sum(q.points) as max_score
		FROM quizzes qz
		LEFT JOIN user_quizzes uq ON uq.quiz_id = qz.id AND uq.user_id = $1
		LEFT JOIN questions q ON qz.id= q.quiz_id
		WHERE qz.id= $2
		GROUP BY qz.id, uq.total_points_awarded, uq.is_completed, qz.title, qz.active_to;
		`, userID, quizID)

	var summary UserQuizSummary
	var isQuizComplete bool
	err := quizRow.Scan(&summary.QuizID, &summary.AchievedScore, &isQuizComplete, &summary.QuizTitle, &summary.QuizActiveTo, &summary.MaxScore)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNoSuchQuiz
		}
	}
	if !isQuizComplete {
		return nil, ErrQuizNotCompleted
	}

	answeredQuestions, err := getAnsweredQuestions(db, userID, quizID)
	if err != nil {
		return nil, err
	}
	summary.AnsweredQuestions = answeredQuestions

	articles, err := articles.GetUsedArticlesByQuizID(db, quizID)
	if err != nil {
		return nil, err
	}
	summary.HasArticlesToShow = len(*articles) > 0

	return &summary, nil
}

func getAnsweredQuestions(db *sql.DB, userID uuid.UUID, quizID uuid.UUID) ([]AnsweredQuestion, error) {
	rows, err := db.Query(
		`SELECT uqp.question_id, q.question, q.points, uqp.chosen_answer_alternative_id, a.text, a.correct, uqp.points_awarded
		FROM user_question_points uqp
		LEFT JOIN questions q ON uqp.question_id = q.id
		LEFT JOIN answer_alternatives a ON uqp.chosen_answer_alternative_id = a.id
		WHERE uqp.quiz_id = $1
		AND uqp.user_id = $2;`, quizID, userID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var answeredQuestions []AnsweredQuestion
	for rows.Next() {
		var aq AnsweredQuestion
		err := rows.Scan(
			&aq.QuestionID,
			&aq.QuestionText,
			&aq.MaxPoints,
			&aq.ChosenAlternativeID,
			&aq.ChosenAlternativeText,
			&aq.IsCorrect,
			&aq.PointsAwarded,
		)
		if err != nil {
			return nil, err
		}

		answeredQuestions = append(answeredQuestions, aq)
	}

	return answeredQuestions, nil
}

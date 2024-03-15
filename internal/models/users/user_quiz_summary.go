package users

import (
	"database/sql"

	"github.com/google/uuid"
)

type UserQuizSummary struct {
	quizID            uuid.UUID
	QuizTitle         string
	MaxScore          uint
	AchievedScore     uint
	AnsweredQuestions []UserAnsweredQuestion
}

func GetQuizSummary(db *sql.DB, userID uuid.UUID, quizID uuid.UUID) (UserQuizSummary, error) {
	return UserQuizSummary{}, nil
}

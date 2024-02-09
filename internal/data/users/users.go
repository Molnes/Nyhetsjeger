package users

import (
	"github.com/Molnes/Nyhetsjeger/internal/data/quizzes"
	"github.com/google/uuid"
)

type User struct {
	ID              uuid.UUID
	Username        string
	TotalScore      uint
	CompletedQuizzes []quizzes.Quiz
}

package users

import (
	"github.com/Molnes/Nyhetsjeger/internal/data/quizes"
	"github.com/google/uuid"
)

type User struct {
	ID              uuid.UUID
	Username        string
	TotalScore      uint
	CompletedQuizes []quizes.Quiz
}

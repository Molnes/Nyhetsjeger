package users

import (
	"github.com/Molnes/Nyhetsjeger/internal/data/questions"
	"github.com/google/uuid"
)

type UserAnsweredQUestion struct {
	UserID            uuid.UUID
	Question          questions.Question
	ChosenAlternative questions.Alternative
}

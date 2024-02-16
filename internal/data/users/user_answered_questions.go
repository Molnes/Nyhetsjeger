package users

import (
	"github.com/Molnes/Nyhetsjeger/internal/data/questions"
	"github.com/google/uuid"
)

type UserAnsweredQuestion struct {
	UserID            uuid.UUID
	QuestionID        uuid.UUID
	QuestionText      string
	ChosenAlternative questions.Alternative
}

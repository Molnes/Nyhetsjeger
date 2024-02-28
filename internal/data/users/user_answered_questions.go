package users

import (
	"time"

	"github.com/Molnes/Nyhetsjeger/internal/data/questions"
	"github.com/google/uuid"
)

type UserAnsweredQuestion struct {
	UserID              uuid.UUID
	QuestionID          uuid.UUID
	QuestionText        string
	QuestionPresentedAt time.Time
	ChosenAlternative   questions.Alternative
	AnsweredAt          time.Time
	PointsAwarded       uint
}

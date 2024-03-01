package users

import (
	"database/sql"
	"time"

	"github.com/Molnes/Nyhetsjeger/internal/data/questions"
	"github.com/google/uuid"
)

type UserAnsweredQuestion struct {
	UserID              uuid.UUID
	QuestionID          uuid.UUID
	QuestionPresentedAt time.Time
	ChosenAlternative   questions.Alternative
	AnsweredAt          time.Time
	PointsAwarded       uint
}


func StartQuestion(db *sql.DB, userId uuid.UUID, questionId uuid.UUID) error {
        db.QueryRow("INSERT INTO user_answers (user_id, question_id, question_presented_at) VALUES ($1, $2, $3) RETURNING user_id, question_id, question_presented_at", userId, questionId, time.Now())
        return nil
}

func  AnswerQuestion(db *sql.DB, userId uuid.UUID, questionId uuid.UUID, chosenAlternative uuid.UUID) error {

       db.QueryRow("UPDATE user_answers SET chosen_alternative = $1, answered_at = $2 WHERE user_id = $3 AND question_id = $4 RETURNING chosen_alternative, answered_at", chosenAlternative, time.Now(), userId, questionId)
        return nil
}

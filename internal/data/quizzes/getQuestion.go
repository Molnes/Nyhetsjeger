package quizzes

import (
	"github.com/Molnes/Nyhetsjeger/internal/data/questions"
	"github.com/google/uuid"
)

// Gets the question from the id
func GetQuestionFromId(id uuid.UUID) *questions.Question {
	for _, question := range SampleQuiz.Questions {
		if question.ID == id {
			return &question
		}
	}
	return nil
}
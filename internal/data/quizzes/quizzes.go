package quizzes

import (
	"github.com/Molnes/Nyhetsjeger/internal/data/questions"
	"github.com/google/uuid"
)

type Quiz struct {
	ID        uuid.UUID
	Title     string
	Questions []questions.Question
}

func GetQuiz(quizID uuid.UUID) (Quiz, error) {
	return SampleQuiz, nil
}

var SampleQuiz Quiz = Quiz{
	ID:        uuid.New(),
	Title:     "Sample quiz",
	Questions: questions.SampleQuestions,
}

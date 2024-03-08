package users

import (
	"github.com/Molnes/Nyhetsjeger/internal/data/questions"
	"github.com/google/uuid"
)

type UserQuizSummary struct {
	quizID            uuid.UUID
	QuizTitle         string
	MaxScore          uint
	AchievedScore     uint
	AnsweredQuestions []UserAnsweredQuestion
}

var SampleUserQuizSummary UserQuizSummary = UserQuizSummary{
	QuizTitle:     "Example quiz",
	MaxScore:      10,
	AchievedScore: 5,
	AnsweredQuestions: []UserAnsweredQuestion{
		{
			QuestionID:   uuid.New(),
			ChosenAlternative: questions.Alternative{
				ID:        uuid.New(),
				Text:      "alt1 123",
				IsCorrect: true,
			},
		}, {
			QuestionID:   uuid.New(),
			ChosenAlternative: questions.Alternative{
				ID:        uuid.New(),
				Text:      "alt2 lorem ipsum",
				IsCorrect: false,
			},
		},
	},
}

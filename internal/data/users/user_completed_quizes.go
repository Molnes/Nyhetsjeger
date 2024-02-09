package users

import (
	quizes "github.com/Molnes/Nyhetsjeger/internal/data/quizzes"
	"github.com/google/uuid"
)

type UserCompletedQuiz struct {
	Quiz              quizes.Quiz
	AnsweredQuestions []UserAnsweredQUestion
}

func GetCompletedQuizes(userID uuid.UUID) ([]UserCompletedQuiz, error) {
	return SampleUserCompletedQuizes, nil
}

var SampleUserCompletedQuizes []UserCompletedQuiz = []UserCompletedQuiz{
	{
		Quiz: quizes.SampleQuiz,
		AnsweredQuestions: []UserAnsweredQUestion{
			{
				UserID:            uuid.New(),
				QuestionID:        uuid.New(),
				ChosenAlternative: quizes.SampleQuiz.Questions[0].Alternatives[0],
			},
		},
	},
}

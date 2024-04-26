package data_converting

import (
	"database/sql"

	"github.com/Molnes/Nyhetsjeger/internal/models/quizzes"
)

// Simple function to convert quizzes to partial quizzes
func ConvertQuizzesToPartial(quizList []quizzes.Quiz, db *sql.DB) ([]quizzes.PartialQuiz, error) {
	partialQuizzes := []quizzes.PartialQuiz{}

	for _, quiz := range quizList {
		partialQuiz, err := quizzes.GetPartialQuizByID(db, quiz.ID)
		if err != nil {
			return []quizzes.PartialQuiz{}, err
		}
		partialQuizzes = append(partialQuizzes, *partialQuiz)
	}
	return partialQuizzes, nil
}

// Simple function to convert a quiz to a partial quiz
func ConvertQuizToPartial(quiz quizzes.Quiz, db *sql.DB) (*quizzes.PartialQuiz, error) {
	partialQuiz, err := quizzes.GetPartialQuizByID(db, quiz.ID)
	if err != nil {
		return nil, err
	}
	return partialQuiz, nil
}

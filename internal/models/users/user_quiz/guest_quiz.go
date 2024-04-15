package user_quiz

import (
	"database/sql"

	"github.com/Molnes/Nyhetsjeger/internal/models/questions"
	"github.com/Molnes/Nyhetsjeger/internal/models/quizzes"
	"github.com/google/uuid"
)

func GetQuestionByNumberInQuiz(db *sql.DB, quizID uuid.UUID, questionNumber uint) (*QuizData, error) {
	partialQuiz, err := quizzes.GetPartialQuizByID(db, quizID)
	if err != nil || !partialQuiz.Published || partialQuiz.QuestionNumber == 0 {
		return nil, ErrNoSuchQuiz
	}
	// nextQuestion, secondsLeft, err := startNextQuestion(db, userID, quizID)
	// if err != nil {
	// 	return nil, err
	// }

	question, err := questions.GetNthQuestionByQuizId(db, quizID, questionNumber)

	if err != nil {
		return nil, err
	}

	// pointsSoFar, err := getPointsGatheredInQuiz(db, quizID, userID)

	return &QuizData{
		*partialQuiz,
		*question,
		0,
		0,
	}, nil
}

func GetOpenQuizId(db *sql.DB) (uuid.UUID, error) {
	var id uuid.UUID
	err := db.QueryRow(`
	SELECT id from quizzes limit 1;
	`).Scan(&id)

	return id, err
}

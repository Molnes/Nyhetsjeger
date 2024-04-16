package user_quiz

import (
	"database/sql"
	"time"

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
		question.TimeLimitSeconds,
	}, nil
}

func GetOpenQuizId(db *sql.DB) (uuid.UUID, error) {
	var id uuid.UUID
	err := db.QueryRow(`
	SELECT id from quizzes limit 1;
	`).Scan(&id)

	return id, err
}

// Gets UserAnsweredQuestion data for the answer without saving dany data in the database.
func AnswerQuestionGuest(db *sql.DB, questionId uuid.UUID, chosenAnswerId uuid.UUID, questionPresentedAt time.Time) (*UserAnsweredQuestion, error) {
	answeredQuestion := UserAnsweredQuestion{
		ChosenAnswerID: chosenAnswerId,
	}

	question, err := questions.GetQuestionByID(db, questionId)
	if err != nil {
		return nil, err
	}
	answeredQuestion.Question = *question

	if question.IsAnswerCorrect(chosenAnswerId) {
		points, err := calculatePointsWithSqlFunction(db, question, questionPresentedAt)
		if err != nil {
			return nil, err
		}
		answeredQuestion.PointsAwarded = points
	}

	nextQuestionId, err := questions.GetNextQuestionInQuizByQuestionId(db, questionId)
	if err != nil {
		if err != sql.ErrNoRows {
			return nil, err
		}
		nextQuestionId = uuid.Nil
	}
	answeredQuestion.NextQuestionID = nextQuestionId

	return &answeredQuestion, nil

}

// Uses the function defined in the database to calculatem how many points would be awarded. No data saved in the database.
func calculatePointsWithSqlFunction(db *sql.DB, question *questions.Question, questionPresentedAt time.Time) (uint, error) {
	row := db.QueryRow(`
	SELECT calculate_points_awarded($1, $2, $3, $4);
	`, questionPresentedAt, time.Now(), question.TimeLimitSeconds, question.Points)
	var points uint
	err := row.Scan(&points)
	return points, err
}

package user_quiz

import (
	"database/sql"
	"errors"
	"math"
	"time"

	"github.com/Molnes/Nyhetsjeger/internal/models/questions"
	"github.com/Molnes/Nyhetsjeger/internal/models/quizzes"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

var ErrNoSuchQuiz = errors.New("user_quiz: no such quiz")
var ErrNoMoreQuestions = errors.New("user_quiz: no more unanswered questions in quiz")
var ErrNoSuchAnswer = errors.New("user_quiz: no such answer")

// Returns the ID of the next question the provided user has not answered in the provided quiz.
//
// If there are no more questions, returns ErrNoMoreQuestions.
func getNextUnansweredQuestionID(db *sql.DB, userID uuid.UUID, quizID uuid.UUID) (uuid.UUID, error) {
	row := db.QueryRow(
		`SELECT id
		FROM questions
		WHERE quiz_id = $1
		AND id NOT IN (
			SELECT question_id
			FROM user_answers
			WHERE chosen_answer_alternative_id IS NOT NULL
			AND user_id = $2
		)
		ORDER BY arrangement
		LIMIT 1;`, quizID, userID)

	var id uuid.UUID
	err := row.Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return uuid.Nil, ErrNoMoreQuestions
		}
		return uuid.Nil, err
	}
	return id, nil
}

var errQuestionAlreadyStarted = errors.New("user quiz: question already started")

// Initiates the answering process for a question.
//
// Saves that user was presented with the question (question presented at). If the user has already started answering the question, returns errQuestionAlreadyStarted.
func startQuestion(db *sql.DB, userId uuid.UUID, questionId uuid.UUID) error {
	_, err := db.Exec(
		`INSERT INTO user_answers
		(user_id, question_id, question_presented_at)
		VALUES ($1, $2, $3)`, userId, questionId, time.Now().UTC())

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return errQuestionAlreadyStarted
		}
	}

	return err
}

// Returns the time the use was presented with the question.
// If the user has not been presented with the question, returns an error.
func getQuestionPresentedAtTime(db *sql.DB, userId uuid.UUID, questionId uuid.UUID) (time.Time, error) {
	var questionPresentedAt time.Time
	err := db.QueryRow(
		`SELECT question_presented_at
		FROM user_answers
		WHERE user_id = $1 AND question_id = $2;`, userId, questionId,
	).Scan(&questionPresentedAt)
	if err != nil {
		return time.Time{}, err
	}
	return questionPresentedAt, nil
}

type QuizData struct {
	PartialQuiz     quizzes.PartialQuiz
	CurrentQuestion questions.Question
}

// Returns the next question in the quiz for the user and saves the time it was presented.
//
// May return:
//
// ErrNoSuchQuiz if the quiz does not exist.
// ErrNoMoreQuestions if there are no more unanswered questions for the user.
func NextQuestionInQuiz(db *sql.DB, userID uuid.UUID, quizID uuid.UUID) (*QuizData, error) {
	partialQuiz, err := quizzes.GetPartialQuizByID(db, quizID)
	if err != nil || !partialQuiz.Published || partialQuiz.QuestionNumber == 0 {
		return nil, ErrNoSuchQuiz
	}
	nextQuestion, err := StartNextQuestion(db, userID, quizID)
	if err != nil {
		return nil, err
	}
	return &QuizData{
		*partialQuiz,
		*nextQuestion,
	}, nil

}

// Returns the next question in the quiz for the user and saves the time it was presented.
//
// May return:
//
// ErrNoMoreQuestions if there are no more unanswered questions for the user in the quiz.
// ErrNoSuchQuiz if the quiz does not exist.
func StartNextQuestion(db *sql.DB, userID uuid.UUID, quizID uuid.UUID) (*questions.Question, error) {
	questionID, err := getNextUnansweredQuestionID(db, userID, quizID)
	if err != nil {
		return nil, err
	}
	question, err := questions.GetQuestionByID(db, questionID)
	if err != nil {
		return nil, err
	}

	err = startQuestion(db, userID, question.ID)
	if err != nil {
		if err == errQuestionAlreadyStarted {
			timePresented, err := getQuestionPresentedAtTime(db, userID, question.ID)
			if err != nil {
				return nil, err
			}
			question.SubtractFromTimeLimit(time.Since(timePresented))
		} else {
			return nil, err
		}
	}
	return question, nil
}

type UserAnsweredQuestion struct {
	Question       questions.Question
	ChosenAnswerID uuid.UUID
	PointsAwarded  uint
	NextQuestionID uuid.UUID
}

var ErrQuestionAlreadyAnswered = errors.New("user quiz: question already answered")

// Saves the user's answer to a question and returns the result as a UserAnsweredQuestion.
//
// May return:
//
// ErrQuestionAlreadyAnswered if the user has already answered the question.
func AnswerQuestion(db *sql.DB, userId uuid.UUID, questionId uuid.UUID, chosenAlternative uuid.UUID) (*UserAnsweredQuestion, error) {
	var questionPresentedAt time.Time
	var chosenAnswerIdNull uuid.UUID
	var maxPoints uint
	var timeLimit uint
	var quizID uuid.UUID
	var quizOpenTime time.Time
	err := db.QueryRow(
		`SELECT question_presented_at, questions.points, questions.time_limit_seconds, chosen_answer_alternative_id, questions.quiz_id
		FROM user_answers JOIN questions ON user_answers.question_id = questions.id
		WHERE user_id = $1 AND question_id = $2;`, userId, questionId,
	).Scan(&questionPresentedAt, &maxPoints, &timeLimit, &chosenAnswerIdNull, &quizID)
	if err != nil {
		return nil, err
	}
	if chosenAnswerIdNull != uuid.Nil {
		return nil, ErrQuestionAlreadyAnswered
	}

	isCorrect, err := questions.IsCorrectAnswer(db, questionId, chosenAlternative)
	if err != nil {
		return nil, err
	}

	// Check if the quiz is open.
	err = db.QueryRow(
		`SELECT available_to FROM quizzes WHERE id = $1;`, quizID,
	).Scan(&quizOpenTime)
	if err != nil {
		return nil, err
	}

	nowTime := time.Now().UTC()
	var pointsAwarded uint
	if isCorrect {
		pointsAwarded = calculatePoints(questionPresentedAt, nowTime, timeLimit, maxPoints, nowTime.After(quizOpenTime))
	}
	_, err = db.Exec(
		`UPDATE user_answers
		SET chosen_answer_alternative_id = $1, answered_at = $2, points_awarded = $3
		WHERE user_id = $4 AND question_id = $5;`,
		chosenAlternative, nowTime, pointsAwarded, userId, questionId)

	if err != nil {
		return nil, err
	}

	question, err := questions.GetQuestionByID(db, questionId)
	if err != nil {
		return nil, err
	}
	nextQuestionID, err := getNextUnansweredQuestionID(db, userId, question.QuizID)
	if err != nil {
		if err != ErrNoMoreQuestions {
			return nil, err
		}
		nextQuestionID = uuid.Nil
	}

	return &UserAnsweredQuestion{
		Question:       *question,
		ChosenAnswerID: chosenAlternative,
		PointsAwarded:  pointsAwarded,
		NextQuestionID: nextQuestionID,
	}, nil
}

// Calculates the points awarded for answering a question. The points are based on the time taken to answer the question.
// As of now there are 3 possible outcomes: 100%, 50% or 25% of the max points.
// These are based on the thresholds: 0-25%, 25-50% and 50+% of the time limit used.
func calculatePoints(questionPresentadAt time.Time, answeredAt time.Time, timeLimit uint, maxPoints uint, pastQuizOpenTime bool) uint {

	diff := answeredAt.Sub(questionPresentadAt)

	secondsTaken := diff.Seconds()

	threshholdMax := 0.25
	threshholdMid := 0.5

	var pointsAwarded float64

	timeLimitFloat := float64(timeLimit)
	if secondsTaken < threshholdMax*timeLimitFloat {
		pointsAwarded = float64(maxPoints)
	} else if secondsTaken < threshholdMid*timeLimitFloat {
		pointsAwarded = float64(maxPoints) / 2
	} else {
		pointsAwarded = float64(maxPoints) / 4
	}
	// If the quiz is not open, the points awarded are halved.
	if pastQuizOpenTime {
		pointsAwarded = pointsAwarded * 0.5
	}

	rounded := math.RoundToEven(pointsAwarded)
	return uint(rounded)
}

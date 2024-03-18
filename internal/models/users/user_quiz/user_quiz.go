package user_quiz

import (
	"database/sql"
	"errors"
	"math"
	"time"

	"github.com/Molnes/Nyhetsjeger/internal/models/questions"
	"github.com/google/uuid"
)

var ErrNoSuchQuiz = errors.New("user_quiz: no such quiz")
var ErrNoMoreQuestions = errors.New("user_quiz: no more unanswered questions in quiz")
var ErrNoSuchAnswer = errors.New("user_quiz: no such answer")

func getNextUnansweredQuestion(db *sql.DB, userID uuid.UUID, quizID uuid.UUID) (*questions.Question, error) {
	row := db.QueryRow(
		`SELECT id
		FROM questions
		WHERE quiz_id = $1
		AND id NOT IN (
			SELECT question_id
			FROM user_answers
			WHERE user_id = $2
		)
		ORDER BY arrangement
		LIMIT 1;`, quizID, userID)

	var id uuid.UUID
	err := row.Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNoMoreQuestions
		}
		return nil, err
	}
	return questions.GetQuestionByID(db, id)
}

func startQuestion(db *sql.DB, userId uuid.UUID, questionId uuid.UUID) error {
	_, err := db.Exec(
		`INSERT INTO user_answers
		(user_id, question_id, question_presented_at)
		VALUES ($1, $2, $3)`, userId, questionId, time.Now())
	return err
}

// Returns the next question in the quiz for the user and saves the time it was presented.
//
// May return:
//
// ErrNoMoreQuestions if there are no more unanswered questions for the user in the quiz.
// ErrNoSuchQuiz if the quiz does not exist.
func StartNextQuestion(db *sql.DB, userID uuid.UUID, quizID uuid.UUID) (*questions.Question, error) {
	question, err := getNextUnansweredQuestion(db, userID, quizID)
	if err != nil {
		return nil, err
	}
	err = startQuestion(db, userID, question.ID)
	if err != nil {
		return nil, err
	}
	return question, nil
}


type UserAnsweredQuestion struct {
	Question       questions.Question
	ChosenAnswerID uuid.UUID
	PointsAwarded  uint
}

var ErrQuestionAlreadyAnswered = errors.New("user quiz: question already answered")

func AnswerQuestion(db *sql.DB, userId uuid.UUID, questionId uuid.UUID, chosenAlternative uuid.UUID) (*UserAnsweredQuestion, error) {
	var questionPresentedAt time.Time
	var chosenAnsweIdNull uuid.UUID
	var maxPoints uint
	var timeLimit time.Time
	err := db.QueryRow(
		`SELECT question_presented_at, questions.points, questions.time_limit_seconds, chosen_answer_alternative_id
		FROM user_answers JOIN questions ON user_answers.question_id = questions.id
		WHERE user_id = $1 AND question_id = $2;`, userId, questionId,
	).Scan(&questionPresentedAt, &maxPoints, &timeLimit, &chosenAnsweIdNull)
	if err != nil {
		return nil, err
	}
	if chosenAnsweIdNull != uuid.Nil {
		return nil, ErrQuestionAlreadyAnswered
	}

	nowTime := time.Now()
	pointsAwarded := calculatePoints(questionPresentedAt, nowTime, uint(timeLimit.Second()), maxPoints)
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

	return &UserAnsweredQuestion{
		Question:       *question,
		ChosenAnswerID: chosenAlternative,
		PointsAwarded:  pointsAwarded,
	}, nil
}

func calculatePoints(questionPresentadAt time.Time, answeredAt time.Time, timeLimit uint, maxPoints uint) uint {

	diff := answeredAt.Sub(questionPresentadAt)
	secondsTaken := diff.Seconds()

	var pointsAwarded float64

	timeLimitFloat := float64(timeLimit)
	if secondsTaken < 0.75*timeLimitFloat {
		pointsAwarded = float64(maxPoints)
	} else if secondsTaken < 0.5*timeLimitFloat {
		pointsAwarded = float64(maxPoints) / 2
	} else {
		pointsAwarded = float64(maxPoints) / 4
	}

	rounded := math.RoundToEven(pointsAwarded)
	return uint(rounded)

}
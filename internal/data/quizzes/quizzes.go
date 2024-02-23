package quizzes

import (
	"net/url"
	"time"

	"github.com/Molnes/Nyhetsjeger/internal/data/questions"
	"github.com/google/uuid"
)

type Quiz struct {
	ID             uuid.UUID
	Title          string
	ImageURL       url.URL
	AvailableFrom  time.Time
	AvailableTo    time.Time
	CreatedAt      time.Time
	LastModifiedAt time.Time
	Questions      []questions.Question
}

func GetQuiz(quizID uuid.UUID) (Quiz, error) {
	return SampleQuiz, nil
}

// Create a default quiz.
// This function is used to create a new quiz with default values.
func CreateDefaultQuiz() (Quiz, error) {
	return Quiz{
		ID:    uuid.New(),
		Title: "Daglig Quiz: " + time.Now().Format("01 Jan 2006"),
		ImageURL: url.URL{
			Scheme: "https",
			Host:   "unsplash.it",
			Path:   "/200/200",
		},
		AvailableFrom:  time.Now(),
		AvailableTo:    time.Now().Add(24 * time.Hour),
		CreatedAt:      time.Now(),
		LastModifiedAt: time.Now(),
		Questions:      []questions.Question{},
	}, nil
}

var SampleQuiz Quiz = Quiz{
	ID:        uuid.New(),
	Title:     "Sample quiz",
	Questions: questions.SampleQuestions,
}

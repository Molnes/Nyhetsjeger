package main

import (
	"context"
	"log"
	"time"

	"github.com/Molnes/Nyhetsjeger/internal/database"
	"github.com/google/uuid"
)

func main() {
	defer database.DB.Close()

	log.Println("----- Populating database -----")
	defer log.Println("----- Database populated -----")

	createSampleQuiz("Daglig quiz 17/02/24")
	createSampleQuiz("Daglig quiz 01/03/24")
}

func createSampleQuiz(title string) {
	var quiz_id uuid.UUID
	row := database.DB.QueryRow(
		`INSERT INTO quizzes (title, available_from, available_to)
		values ($1, $2, $3)
		RETURNING id;`,
		title, time.Now(), time.Now().Add(time.Hour*24))

	err := row.Scan(&quiz_id)
	if err != nil {
		log.Println(err)
	}

	createQuestion(quiz_id, "What is the capital of Norway?", 1,
		[]answerAlt{
			{"Oslo", true},
			{"Bergen", false},
			{"Trondheim", false},
			{"Stavanger", false},
		})
	createQuestion(quiz_id, "What is the capital of Sweden?", 2,
		[]answerAlt{
			{"Stockholm", true},
			{"Gothenburg", false},
			{"Malmö", false},
			{"Uppsala", false},
		})

}

type answerAlt struct {
	answer  string
	correct bool
}

func createQuestion(quiz_id uuid.UUID, question string, arrangement int, answers []answerAlt) {
	question_id := uuid.New()

	ctx := context.Background()
	tx, err := database.DB.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
	}
	tx.Exec(
		`INSERT INTO questions (id, question, arrangement, quiz_id)
		VALUES ($1, $2, $3, $4);`,
		question_id, question, arrangement, quiz_id)

	for _, a := range answers {
		tx.Exec(
			`INSERT INTO answer_alternatives (question_id, text, correct)
			VALUES ($1, $2, $3);`,
			question_id, a.answer, a.correct)
	}
	if err := tx.Commit(); err != nil {
		log.Println(err)
	}

}

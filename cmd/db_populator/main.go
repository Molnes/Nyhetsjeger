package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"time"

	"github.com/Molnes/Nyhetsjeger/internal/database"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("DB Populator: Error loading .env")
	}

	dburl, ok := os.LookupEnv("POSTGRESQL_URL_APP")
	if !ok {
		log.Fatal("DB Populator: No database url provided. Expected POSTGRESQL_URL_APP")
	}

	db, err := database.NewDatabaseConnection(dburl)
	if err != nil {
		log.Fatal("DB Populator: Error connecting to database: ", err)
	}

	defer db.Close()

	log.Println("----- Populating database -----")
	defer log.Println("----- Database populated -----")

	createSampleQuiz(db, "Daglig quiz 17/02/24")
	createSampleQuiz(db, "Daglig quiz 01/03/24")
}

func createSampleQuiz(db *sql.DB, title string) {
	var quiz_id uuid.UUID
	row := db.QueryRow(
		`INSERT INTO quizzes (title, available_from, available_to)
		values ($1, $2, $3)
		RETURNING id;`,
		title, time.Now(), time.Now().Add(time.Hour*24))

	err := row.Scan(&quiz_id)
	if err != nil {
		log.Println(err)
	}

	for range 3 {
		createQuestion(db, quiz_id, sampleQuestion1)
		createQuestion(db, quiz_id, sampleQuestion2)
		createQuestion(db, quiz_id, sampleQuestion3)
	}
}

type answerAlt struct {
	answer  string
	correct bool
}

type question struct {
	text        string
	answer_alts []answerAlt
}

func createQuestion(db *sql.DB, quiz_id uuid.UUID, question question) {
	question_id := uuid.New()

	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
	}
	tx.Exec(
		`INSERT INTO questions (id, question, quiz_id)
		VALUES ($1, $2, $3);`,
		question_id, question.text, quiz_id)

	for _, a := range question.answer_alts {
		tx.Exec(
			`INSERT INTO answer_alternatives (question_id, text, correct)
			VALUES ($1, $2, $3);`,
			question_id, a.answer, a.correct)
	}
	if err := tx.Commit(); err != nil {
		log.Println(err)
	}

}

var sampleQuestion1 = question{
	"What is the capital of Norway?",
	[]answerAlt{
		{"Oslo", true},
		{"Bergen", false},
		{"Trondheim", false},
		{"Stavanger", false},
	},
}

var sampleQuestion2 = question{
	"What is the capital of Sweden? This is a longer question text.",
	[]answerAlt{
		{"Stockholm", true},
		{"Gothenburg", false},
		{"Malmö", false},
		{"Bergen", false},
	},
}

var sampleQuestion3 = question{
	"What is the capital of Denmark? Long alternatives",
	[]answerAlt{
		{"Copenhagen is the capital", true},
		{"Warsaw is the capital of Denmark", false},
		{"Oslo", false},
		{"Copenhagen again to test 2 correct", true},
	},
}

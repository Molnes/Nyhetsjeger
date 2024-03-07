package main

import (
	"context"
	"database/sql"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/Molnes/Nyhetsjeger/internal/data/articles"
	"github.com/Molnes/Nyhetsjeger/internal/database"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("DB Populator: Error loading .env")
	}

	dburl, ok := os.LookupEnv("POSTGRESQL_URL_DEV")
	if !ok {
		log.Fatal("DB Populator: No database url provided. Expected POSTGRESQL_URL_DEV")
	}

	db, err := database.NewDatabaseConnection(dburl)
	if err != nil {
		log.Fatal("DB Populator: Error connecting to database: ", err)
	}

	defer db.Close()

	log.Println("----- Populating database -----")
	defer log.Println("----- Database populated -----")

	createSampleQuiz(db, "Ukentlig quiz 1")
	createSampleQuiz(db, "Ukentlig quiz 2")
}

func createSampleQuizArticle(db *sql.DB, quizID uuid.UUID) {
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
	}

	article := articles.Article{
		ID: uuid.NullUUID{
			UUID:  uuid.New(),
			Valid: true,
		},
		Title: "Sample article",
		ArticleURL: url.URL{
			Scheme: "https",
			Host:   "www.example.com",
		},
		ImgURL: url.URL{
			Scheme: "https",
			Host:   "www.unsplash.it",
			Path:   "/200/200",
		},
	}

	tx.Exec(
		`INSERT INTO 
			articles (id, title, url, image_url)
		VALUES 
			($1, $2, $3, $4);`,
		article.ID, article.Title, article.ArticleURL.String(), article.ImgURL.String())

	tx.Exec(
		`INSERT INTO 
			quiz_articles (quiz_id, article_id)
		VALUES 
			($1, $2);`,
		quizID, article.ID)

	if err := tx.Commit(); err != nil {
		log.Println(err)
	}
}

func createSampleQuiz(db *sql.DB, title string) {
	var quiz_id uuid.UUID
	row := db.QueryRow(
		`INSERT INTO quizzes (title, available_from, available_to, image_url)
		values ($1, $2, $3, $4)
		RETURNING id;`,
		title, time.Now(), time.Now().Add(time.Hour*24*7), "https://www.unsplash.it/200/200")

	err := row.Scan(&quiz_id)
	if err != nil {
		log.Println(err)
	}

	createSampleQuizArticle(db, quiz_id)

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
		`INSERT INTO questions (id, question, image_url, article_id, quiz_id, points)
		VALUES ($1, $2, $3, $4, $5, $6);`,
		question_id, question.text, "https://unsplash.it/200/200", nil, quiz_id, 10)

	for _, a := range question.answer_alts {
		alternative_id := uuid.New()
		tx.Exec(
			`INSERT INTO answer_alternatives (id, question_id, text, correct)
			VALUES ($1, $2, $3, $4);`,
			alternative_id, question_id, a.answer, a.correct)
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

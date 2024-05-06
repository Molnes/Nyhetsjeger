package db_populator

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/Molnes/Nyhetsjeger/internal/models/articles"
	"github.com/Molnes/Nyhetsjeger/internal/models/users"
	"github.com/google/uuid"
)

func PopulateDbWithTestData(db *sql.DB) {
	createSampleQuiz(db, "Ukentlig quiz 1")
	createSampleQuiz(db, "Ukentlig quiz 2")
	createTestUser(db)
}

func createSampleQuizArticle(db *sql.DB, quizID uuid.UUID) {
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
	}

	article := getSampleAricle()

	article.ArticleURL.Path = article.ArticleURL.Path + article.ID.UUID.String()

	_, err = tx.Exec(
		`INSERT INTO 
			articles (id, title, url, image_url)
		VALUES 
			($1, $2, $3, $4);`,
		article.ID, article.Title, article.ArticleURL.String(), article.ImgURL.String())
	if err != nil {
		log.Println(err)
	}

	_, err = tx.Exec(
		`INSERT INTO 
			quiz_articles (quiz_id, article_id)
		VALUES 
			($1, $2);`,
		quizID, article.ID)
	if err != nil {
		log.Println(err)
	}

	if err := tx.Commit(); err != nil {
		log.Println(err)
	}
}

func createSampleQuiz(db *sql.DB, title string) {
	var quizID uuid.UUID
	row := db.QueryRow(
		`INSERT INTO quizzes (title, active_from, active_to, image_url, published)
		values ($1, $2, $3, $4, true)
		RETURNING id;`,
		title, time.Now(), time.Now().Add(time.Hour*24*7), "https://picsum.photos/id/1062/500/300")

	err := row.Scan(&quizID)
	if err != nil {
		log.Println(err)
	}

	createSampleQuizArticle(db, quizID)

	for range 3 {
		createQuestion(db, quizID, sampleQuestion1)
		createQuestion(db, quizID, sampleQuestion2)
		createQuestion(db, quizID, sampleQuestion3)
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

func createQuestion(db *sql.DB, quizID uuid.UUID, question question) {
	questionID := uuid.New()

	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
	}
	tx.Exec(
		`INSERT INTO questions (id, question, image_url, article_id, quiz_id, points)
		VALUES ($1, $2, $3, $4, $5, $6);`,
		questionID, question.text, "https://picsum.photos/id/1062/500/300", nil, quizID, 10)

	for _, a := range question.answer_alts {
		alternativeID := uuid.New()
		tx.Exec(
			`INSERT INTO answer_alternatives (id, question_id, text, correct)
			VALUES ($1, $2, $3, $4);`,
			alternativeID, questionID, a.answer, a.correct)
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

func getSampleAricle() *articles.Article {
	id := uuid.New()
	article := articles.Article{
		ID: uuid.NullUUID{
			UUID:  id,
			Valid: true,
		},
		Title: "Sample article",
		ArticleURL: url.URL{
			Scheme: "https",
			Host:   "www.example.com",
			Path:   fmt.Sprint("/", id.String()),
		},
		ImgURL: url.URL{
			Scheme: "https",
			Host:   "www.picsum.photos",
			Path:   "/id/1062/500/300",
		},
	}

	return &article
}

func createTestUser(db *sql.DB) error {
	_, err := db.Exec(`INSERT INTO adjectives VALUES ('adj1'), ('adj2');`)
	if err != nil {
		return err
	}
	_, err = db.Exec(`INSERT INTO nouns VALUES ('noun1'), ('noun2');`)
	if err != nil {
		return err
	}
	ctx := context.Background()
	_, err = users.CreateUser(db, ctx, &users.PartialUser{
		SsoID:        "test_user_sso_id",
		Email:        "test_user@email.com",
		AccessToken:  "",
		RefreshToken: "",
		TokenExpire:  time.Now().Add(time.Hour),
	})
	return err
}

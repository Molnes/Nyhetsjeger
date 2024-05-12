package db_populator

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"time"

	"github.com/Molnes/Nyhetsjeger/internal/models/articles"
	// "github.com/Molnes/Nyhetsjeger/internal/models/users"
	"github.com/google/uuid"
)

type knownValues struct {
	UserId    uuid.UUID
	UserSsoId string
	UserEmail string
	QuizId1   uuid.UUID
	Quiz2Id2  uuid.UUID
}

func PopulateDbWithTestData(db *sql.DB) (*knownValues, error) {
	quizId1, err := createSampleQuiz(db, "Ukentlig quiz 1")
	if err != nil {
		return nil, err
	}
	quizId2, err := createSampleQuiz(db, "Ukentlig quiz 2")
	if err != nil {
		return nil, err
	}
	userdata, err := createTestUser(db)
	if err != nil {
		return nil, err
	}

	err = createTestUsernames(db)
	if err != nil {
		return nil, err
	}

	return &knownValues{
		userdata.userId,
		userdata.ssoId,
		userdata.email,
		quizId1,
		quizId2,
	}, nil
}

func createSampleQuizArticle(db *sql.DB, quizID uuid.UUID) error {
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
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
		return err
	}

	_, err = tx.Exec(
		`INSERT INTO 
			quiz_articles (quiz_id, article_id)
		VALUES 
			($1, $2);`,
		quizID, article.ID)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

func createSampleQuiz(db *sql.DB, title string) (uuid.UUID, error) {
	var quizID uuid.UUID
	row := db.QueryRow(
		`INSERT INTO quizzes (title, active_from, active_to, image_url, published)
		values ($1, $2, $3, $4, true)
		RETURNING id;`,
		title, time.Now(), time.Now().Add(time.Hour*24*7), "https://picsum.photos/id/1062/500/300")

	err := row.Scan(&quizID)
	if err != nil {
		return uuid.Nil, err
	}

	err = createSampleQuizArticle(db, quizID)
	if err != nil {
		return uuid.Nil, err
	}

	for range 3 {
		err = createQuestion(db, quizID, sampleQuestion1)
		if err != nil {
			return uuid.Nil, err
		}
		err = createQuestion(db, quizID, sampleQuestion2)
		if err != nil {
			return uuid.Nil, err
		}
		err = createQuestion(db, quizID, sampleQuestion3)
		if err != nil {
			return uuid.Nil, err
		}
	}
	return quizID, nil
}

type answerAlt struct {
	answer  string
	correct bool
}

type question struct {
	text        string
	answer_alts []answerAlt
}

func createQuestion(db *sql.DB, quizID uuid.UUID, question question) error {
	questionID := uuid.New()

	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
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
		return err
	}

	return nil
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

type testUserData struct {
	userId uuid.UUID
	ssoId  string
	email  string
}

func createTestUser(db *sql.DB) (*testUserData, error) {
	_, err := db.Exec(`INSERT INTO adjectives VALUES ('test');`)
	if err != nil {
		return nil, err
	}
	_, err = db.Exec(`INSERT INTO nouns VALUES ('user');`)
	if err != nil {
		return nil, err
	}

	userData := testUserData{
		uuid.New(),
		"test_user_sso_id",
		"test_user@email.com",
	}

	_, err = db.Exec(
		`INSERT INTO users
		(id, sso_user_id, email, phone, opt_in_ranking, accepted_terms, role, access_token, token_expires_at, refresh_token, username_adjective, username_noun)
		SELECT $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, random_username.adjective, random_username.noun
		FROM (
			SELECT adjective, noun
			FROM available_usernames 
			OFFSET floor(random() * (SELECT COUNT(*) FROM available_usernames)) 
		LIMIT 1) AS random_username;`,
		userData.userId, userData.ssoId, userData.email, "no phone", true, true, "user",
		"", time.Now().Add(time.Hour), "")

	if err != nil {
		return nil, err
	}
	return &userData, nil
}

func createTestUsernames(db *sql.DB) error {
	_, err := db.Exec(`INSERT INTO adjectives VALUES ('adj1'), ('adj2');`)
	if err != nil {
		return err
	}
	_, err = db.Exec(`INSERT INTO nouns VALUES ('noun1'), ('noun2');`)
	if err != nil {
		return err
	}
	return nil
}

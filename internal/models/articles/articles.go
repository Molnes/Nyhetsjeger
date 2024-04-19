package articles

import (
	"context"
	"database/sql"
	"net/url"

	data_handling "github.com/Molnes/Nyhetsjeger/internal/utils/data"
	"github.com/google/uuid"
)

type Article struct {
	ID         uuid.NullUUID // The ID of the article.
	Title      string        // The title of the article.
	ArticleURL url.URL       // The URL of the article.
	ImgURL     url.URL       // The URL of the main image of the article.
}

func newArticleNoID(title string, articleURL url.URL, imgURL url.URL) Article {

	return Article{
		ID: uuid.NullUUID{
			UUID:  uuid.New(),
			Valid: true,
		},
		Title:      title,
		ArticleURL: articleURL,
		ImgURL:     imgURL,
	}
}
func newArticle(ID uuid.NullUUID, title string, articleURL url.URL, imgURL url.URL) Article {
	return Article{
		ID:         ID,
		Title:      title,
		ArticleURL: articleURL,
		ImgURL:     imgURL,
	}
}

// Get the articles by Quiz ID.
// These are all the articles attached to a quiz.
// This does not guarantee they are used in the questions.
func GetArticlesByQuizID(db *sql.DB, quizID uuid.UUID) (*[]Article, error) {
	rows, err := db.Query(
		`SELECT
				a.id, a.title, a.url, a.image_url
			FROM
				articles a
			LEFT JOIN
				quiz_articles q ON q.article_id = a.id
			WHERE
				q.quiz_id = $1`,
		quizID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanArticlesFromFullRows(rows)
}

// Get the articles used in a quiz.
// These articles are guaranteed to be used in the questions of the quiz.
func GetUsedArticlesByQuizID(db *sql.DB, quizID uuid.UUID) (*[]Article, error) {
	rows, err := db.Query(
		`SELECT
			DISTINCT(a.id), a.title, a.url, a.image_url
		FROM
			questions q
		JOIN
			articles a ON q.article_id = a.id
		WHERE
			q.quiz_id = $1`,
		quizID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanArticlesFromFullRows(rows)
}

// Get an Article by its ID.
func GetArticleByID(db *sql.DB, id uuid.UUID) (*Article, error) {
	row := db.QueryRow(
		`SELECT
			id, title, url, image_url
		FROM
			articles
		WHERE
			id = $1`,
		id)

	return scanArticleFromFullRow(row)
}

// Convert a row from the database to an Article.
func scanArticleFromFullRow(row *sql.Row) (*Article, error) {
	var article Article
	var articleURL string
	var imageURL sql.NullString
	err := row.Scan(
		&article.ID,
		&article.Title,
		&articleURL,
		&imageURL,
	)

	if err != nil {
		return nil, err
	}

	// Parse the article URL.
	tempArticleURL, err := url.Parse(articleURL)
	article.ArticleURL = *tempArticleURL
	if err != nil {
		return nil, err
	}

	// Set image URL
	tempURL, err := data_handling.ConvertNullStringToURL(&imageURL)
	if err != nil {
		return nil, err
	}
	article.ImgURL = *tempURL

	return &article, nil
}

// Convert a set of rows from the database to a list of Articles.
func scanArticlesFromFullRows(rows *sql.Rows) (*[]Article, error) {
	var articles []Article

	for rows.Next() {
		var article Article
		var articleURL string
		var imageURL sql.NullString

		err := rows.Scan(
			&article.ID,
			&article.Title,
			&articleURL,
			&imageURL,
		)
		if err != nil {
			return nil, err
		}

		// Parse the article URL.
		tempArticleURL, err := url.Parse(articleURL)
		article.ArticleURL = *tempArticleURL
		if err != nil {
			return nil, err
		}

		// Set image URL
		tempURL, err := data_handling.ConvertNullStringToURL(&imageURL)
		if err != nil {
			return nil, err
		}
		article.ImgURL = *tempURL

		articles = append(articles, article)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &articles, nil
}

// Get an Article by its URL from database.
func GetArticleByURL(db *sql.DB, articleURL *url.URL) (*Article, error) {
	row := db.QueryRow(
		`SELECT
		id, title, url, image_url
		FROM
			articles
		WHERE
			url = $1`,
		articleURL.String())

	return scanArticleFromFullRow(row)
}

// Add an Article to the database.
func AddArticle(db *sql.DB, article *Article) error {
	_, err := db.Exec(
		`INSERT INTO
			articles (id, title, url, image_url)
		VALUES
			($1, $2, $3, $4)`,
		article.ID, article.Title, article.ArticleURL.String(), article.ImgURL.String())

	return err
}

// Add an Article to a Quiz by IDs.
func AddArticleToQuizByID(db *sql.DB, articleID *uuid.UUID, quizID *uuid.UUID) error {
	_, err := db.Exec(
		`INSERT INTO
			quiz_articles (quiz_id, article_id)
		VALUES
			($1, $2)`,
		quizID, articleID)

	return err
}

// Returns 'true' if it finds the article in the quiz in the DB.
// Returns 'false' if it does not find it.
func IsArticleInQuiz(db *sql.DB, articleID *uuid.UUID, quizID *uuid.UUID) (bool, error) {
	row := db.QueryRow(
		`SELECT
			article_id
		FROM
			quiz_articles
		WHERE
			article_id = $1 AND quiz_id = $2`,
		articleID, quizID)

	var tempID uuid.UUID
	err := row.Scan(&tempID)

	if err == sql.ErrNoRows {
		return false, nil
	}

	return true, err
}

func DeleteArticleFromQuiz(db *sql.DB, ctx context.Context, quizID *uuid.UUID, articleID *uuid.UUID) error {
	// Open a transaction.
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	// Remove the article from the quiz.
	_, err = tx.Exec(
		`DELETE FROM
			quiz_articles
		WHERE
			quiz_id = $1 AND
			article_id = $2`,
		quizID,
		articleID)

	if err != nil {
		tx.Rollback()
		return err
	}

	// If this is the last quiz that uses the article, delete the article.
	_, err = tx.Exec(
		`DELETE FROM
			articles
		WHERE
			id = $1 AND
			NOT EXISTS (
				SELECT
					article_id
				FROM
					quiz_articles
				WHERE
					article_id = $1
			)`,
		articleID)

	if err != nil {
		tx.Rollback()
		return err
	}

	// Commit the transaction.
	if err := tx.Commit(); err != nil {
		return err
	}

	return err
}

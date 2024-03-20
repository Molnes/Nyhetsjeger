package articles

import (
	"database/sql"
	"net/url"

	data_handling "github.com/Molnes/Nyhetsjeger/internal/utils/data"
	"github.com/google/uuid"
)

type Article struct {
	ID         uuid.NullUUID
	Title      string
	ArticleURL url.URL
	ImgURL     url.URL
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

func GetAllArticles() ([]Article, error) {
	return SampleArticles, nil
}

func GetArticle(articleID uuid.UUID) (Article, error) {
	return SampleArticles[0], nil
}

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
	if err == sql.ErrNoRows {
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
		if err == sql.ErrNoRows {
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

// Get an Article ID by its URL.
func GetArticleIDByURL(db *sql.DB, articleURL *url.URL) (*uuid.NullUUID, error) {
	row := db.QueryRow(
		`SELECT
			id
		FROM
			articles
		WHERE
			url = $1`,
		articleURL.String())

	var articleID uuid.NullUUID
	err := row.Scan(&articleID)

	return &articleID, err
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

// Add an Article to a Quiz.
func AddArticleToQuiz(db *sql.DB, articleID *uuid.UUID, quizID *uuid.UUID) error {
	_, err := db.Exec(
		`INSERT INTO
			quiz_articles (quiz_id, article_id)
		VALUES
			($1, $2)`,
		quizID, articleID)

	return err
}

// Get an Article's data by its URL.
func GetArticleDataByURL(articleURL url.URL) (Article, error) {
	tempID := uuid.NullUUID{
		UUID:  uuid.New(),
		Valid: true,
	}

	article := Article{
		ID:         tempID,
		Title:      "Test article",
		ArticleURL: articleURL,
		ImgURL:     url.URL{},
	}

	return article, nil
}

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

var SampleArticles []Article = []Article{
	{
		ID: uuid.NullUUID{
			UUID:  uuid.New(),
			Valid: true,
		},
		Title: "Test 1",
		ArticleURL: url.URL{
			Scheme: "https",
			Host:   "www.smp.no",
			Path:   "/nyheter/i/Q7QX6Q/konkursnettverket-dette-gjer-regjeringa",
		},
		ImgURL: url.URL{
			Scheme: "https",
			Host:   "unsplash.it",
			Path:   "/200/200",
		},
	},
	{
		ID: uuid.NullUUID{
			UUID:  uuid.New(),
			Valid: true,
		},
		Title: "Test article: this is a very long title in order to check how titles look when they are long",
		ArticleURL: url.URL{
			Scheme: "https",
			Host:   "www.smp.no",
			Path:   "/nyheter/i/0QPq3A/tilraar-nye-ferjekailoeysingar",
		},
		ImgURL: url.URL{
			Scheme: "https",
			Host:   "unsplash.it",
			Path:   "/200/200",
		},
	},
}

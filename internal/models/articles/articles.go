package articles

import (
	"database/sql"
	"net/url"

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

	// Convert the image URL from DB to a URL type.
	if imageURL.Valid {
		tempURL, err := url.Parse(imageURL.String)
		if err != nil {
			return nil, err
		} else {
			article.ImgURL = *tempURL
		}
	} else {
		article.ImgURL = url.URL{}
	}

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

		// Convert the image URL from DB to a URL type.
		if imageURL.Valid {
			tempURL, err := url.Parse(imageURL.String)
			if err != nil {
				return nil, err
			} else {
				article.ImgURL = *tempURL
			}
		} else {
			article.ImgURL = url.URL{}
		}

		articles = append(articles, article)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &articles, nil
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

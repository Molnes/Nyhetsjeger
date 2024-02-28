package articles

import (
	"database/sql"
	"net/url"

	"github.com/google/uuid"
)

type Article struct {
	ID         uuid.UUID
	Title      string
	ArticleURL url.URL
	ImgURL     url.URL
}

func newArticleNoID(title string, articleURL url.URL, imgURL url.URL) Article {
	return Article{
		ID:         uuid.New(),
		Title:      title,
		ArticleURL: articleURL,
		ImgURL:     imgURL,
	}
}
func newArticle(ID uuid.UUID, title string, articleURL url.URL, imgURL url.URL) Article {
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

// Get an Article by its ID
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

// Convert a row from the database to an Article
func scanArticleFromFullRow(row *sql.Row) (*Article, error) {
	var article Article
	var articleURL string
	var imageURL string
	err := row.Scan(
		&article.ID,
		&article.Title,
		&articleURL,
		&imageURL,
	)

	if err != nil {
		return nil, err
	}

	// Parse the article URL
	tempArticleURL, err := url.Parse(articleURL)
	article.ArticleURL = *tempArticleURL
	if err == sql.ErrNoRows {
		return nil, err
	}

	// Parse the image URL
	tempImageURL, err := url.Parse(imageURL)
	article.ImgURL = *tempImageURL
	if err == sql.ErrNoRows {
		return nil, err
	}

	return &article, nil
}

var SampleArticles []Article = []Article{
	{
		ID:    uuid.New(),
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
		ID:    uuid.New(),
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

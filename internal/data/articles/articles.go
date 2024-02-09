package articles

import (
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

var SampleArticles []Article = []Article{
	{
		ID:    uuid.New(),
		Title: "Test 1",
		ArticleURL: url.URL{
			Scheme: "https",
			Host:   "www.example.com",
			Path:   "/test1",
		},
		ImgURL: url.URL{},
	},
	{
		ID:    uuid.New(),
		Title: "Test article2",
		ArticleURL: url.URL{
			Scheme: "https",
			Host:   "www.example.com",
			Path:   "/test1",
		},
		ImgURL: url.URL{},
	},
}

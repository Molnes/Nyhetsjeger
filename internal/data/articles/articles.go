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

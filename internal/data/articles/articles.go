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

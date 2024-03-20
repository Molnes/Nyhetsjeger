package articles

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/google/uuid"
)

// The data structure for a third party article (Sunnmørsposten).
type ArticleSMP struct {
	SchemaVersion int    `json:"schemaVersion"`
	SchemaType    string `json:"schemaType"`
	ID            string `json:"id"`
	Section       struct {
		ID      string `json:"id"`
		Title   string `json:"title"`
		Enabled bool   `json:"enabled"`
	} `json:"section"`
	Title struct {
		Value string `json:"value"`
	} `json:"title"`
	Components []struct {
		Caption struct {
			Value string `json:"value"`
		} `json:"caption,omitempty"`
		Byline struct {
			Title string `json:"title"`
		} `json:"byline,omitempty"`
		ImageAsset struct {
			ID   string `json:"id"`
			Size struct {
				Width  int `json:"width"`
				Height int `json:"height"`
			} `json:"size"`
		} `json:"imageAsset,omitempty"`
		Characteristics struct {
			Figure    bool `json:"figure"`
			Sensitive bool `json:"sensitive"`
		} `json:"characteristics,omitempty"`
		Type       string `json:"type"`
		Paragraphs []struct {
			Text struct {
				Value string `json:"value"`
			} `json:"text"`
			BlockType string `json:"blockType"`
		} `json:"paragraphs,omitempty"`
		Subtype string `json:"subtype,omitempty"`
	} `json:"components"`
}

// Get an ArticleSMP from a JSON file.
func readJSONtoArticleSMP(filename string) (ArticleSMP, error) {
	// Open file
	file, err := os.Open(filename)

	if err != nil {
		log.Println("Error opening file: ", err)
		return ArticleSMP{}, err
	}
	defer file.Close()

	// Read file
	byteValue, err := io.ReadAll(file)

	// Parse JSON to ArticleSMP
	var article ArticleSMP
	err = json.Unmarshal(byteValue, &article)
	if err != nil {
		log.Println("Error parsing JSON: ", err)
		return ArticleSMP{}, err
	}

	return article, nil
}

// Get the ID for ArticleSMP from the article URL.
func getSmpIdFromString(url string) (string, error) {
	// Split the URL by "/"
	parts := strings.Split(url, "/")

	var index int
	for i, part := range parts {
		if part == "i" {
			index = i
			break
		}
	}

	// Get the article ID
	if index+1 < len(parts) {
		return parts[index+1], nil
	}

	return "", nil
}

// Get the main image of the article (the first image).
func getMainImageOfArticle(article ArticleSMP) (*url.URL, error) {
	for _, component := range article.Components {
		if component.Type == "image" {
			return url.Parse(fmt.Sprintf("https://vcdn.polarismedia.no/%s?fit=clip&h=700&q=80&tight=false&w=1000", component.ImageAsset.ID))
		}
	}

	return &url.URL{}, nil
}

// Get an Article by its URL.
func GetArticleByURL(articleUrl string) (Article, error) {
	// TODO: Update this to fetch the article from the web instead of reading it from JSON.

	// Get article's SMP ID
	articleID, err := getSmpIdFromString(articleUrl)
	if err != nil {
		return Article{}, err
	}

	// Get the article data from the JSON file
	articleSMP, err := readJSONtoArticleSMP(fmt.Sprintf("data/articles/%s.json", articleID))
	if err != nil {
		return Article{}, err
	}

	// Parse the article URL
	tempURL, err := url.Parse(articleUrl)
	if err != nil {
		return Article{}, err
	}

	mainImage, err := getMainImageOfArticle(articleSMP)
	if err != nil {
		return Article{}, err
	}

	// Create the article object
	article := Article{
		ID: uuid.NullUUID{
			UUID:  uuid.New(),
			Valid: true,
		},
		Title:      articleSMP.Title.Value,
		ArticleURL: *tempURL,
		ImgURL:     *mainImage,
	}

	return article, nil
}

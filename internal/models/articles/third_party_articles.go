package articles

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/google/uuid"
)

var ErrInvalidArticleID = errors.New("third_party_articles: invalid article ID")
var ErrInvalidArticleURL = errors.New("third_party_articles: invalid article URL")
var ErrArticleNotFound = errors.New("third_party_articles: could not find article")
var ErrUnableToFetchData = errors.New("third_party_articles: unable to fetch article data")

// The data structure for a third party article (Sunnmørsposten).
/* This struct was automatically generated from a JSON file using https://transform.tools/json-to-go */
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
		log.Println("Error opening file: ", err, filename)
		return ArticleSMP{}, ErrArticleNotFound
	}
	defer file.Close()

	// Read file
	byteValue, err := io.ReadAll(file)

	// Parse JSON to ArticleSMP
	var article ArticleSMP
	err = json.Unmarshal(byteValue, &article)
	if err != nil {
		log.Println("Error parsing JSON: ", err)
		return ArticleSMP{}, ErrUnableToFetchData
	}

	return article, nil
}

// Get an ArticleSMP by its URL.
func GetSmpArticleByURL(articleUrl *url.URL) (ArticleSMP, error) {
	articleId, err := GetSmpIdFromString(articleUrl.String())
	if err != nil {
		return ArticleSMP{}, err
	}

	articleSMP, err := readJSONtoArticleSMP(fmt.Sprintf("data/articles/%s.json", articleId))
	if err != nil {
		log.Println("Error getting article: ", err)
		return ArticleSMP{}, err
	}
	return articleSMP, nil
}

// Get the ID for ArticleSMP from the article URL.
func GetSmpIdFromString(url string) (string, error) {
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

// Generate a URL based on ArticleSMP ID.
func GetSmpURLFromID(articleID string) *url.URL {
	return &url.URL{
		Scheme: "https",
		Host:   "www.smp.no",
		Path:   fmt.Sprintf("/i/%s", articleID),
	}
}

// Get the main image of the article (the first image).
func getMainImageOfArticle(article ArticleSMP) (*url.URL, error) {
	for _, component := range article.Components {
		if component.Type == "image" {
			return url.Parse(fmt.Sprintf("https://vcdn.polarismedia.no/%s?fit=clip&h=600&w=900&q=80&tight=true", component.ImageAsset.ID))
		}
	}

	return &url.URL{}, nil
}

// Get all the image components of an article.
func getImagesOfArticle(article ArticleSMP) ([]url.URL, error) {
	var images []url.URL

	for _, component := range article.Components {
		if component.Type == "image" {
			imageURL, err := url.Parse(fmt.Sprintf("https://vcdn.polarismedia.no/%s?fit=clip&h=600&w=900&q=80&tight=true", component.ImageAsset.ID))
			if err != nil {
				return nil, err
			}
			images = append(images, *imageURL)
		}
	}

	return images, nil
}

// Get an Article by its SMP URL.
func GetArticleBySmpUrl(articleUrl string) (Article, error) {
	// TODO: Update this to fetch the article from the web instead of reading it from JSON.

	// Get article's SMP ID
	articleID, err := GetSmpIdFromString(articleUrl)
	if err != nil {
		return Article{}, ErrInvalidArticleID
	}

	// Parse the article URL
	tempURL, err := url.Parse(articleUrl)
	if err != nil {
		return Article{}, ErrInvalidArticleURL
	}

	// Get the article data from the JSON file
	articleSMP, err := readJSONtoArticleSMP(fmt.Sprintf("data/articles/%s.json", articleID))
	if err != nil {
		return Article{}, err
	}

	// Get the main image of the article
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

// Get all the images of a list of articles.
func GetImagesFromArticles(articles *[]Article) ([]url.URL, error) {
	// Read the file using the article ID as filename
	var images []url.URL

	for _, article := range *articles {
		// Get article's SMP ID
		articleID, err := GetSmpIdFromString(article.ArticleURL.String())
		if err != nil {
			return nil, ErrInvalidArticleID
		}

		// Get the article data from the JSON file
		articleSMP, err := readJSONtoArticleSMP(fmt.Sprintf("data/articles/%s.json", articleID))
		if err != nil {
			return nil, err
		}

		// Get all the images of an article
		imageList, err := getImagesOfArticle(articleSMP)
		if err != nil {
			return nil, err
		}

		// Append the images to the list
		images = append(images, imageList...)
	}

	return images, nil
}

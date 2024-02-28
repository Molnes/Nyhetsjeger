package utils

import (
	"encoding/xml"
	"net/http"
	"os"
)

type Item struct {
	Title       string `xml:"title"`
	Description string `xml:"description"`
	Link        string `xml:"link"`
}

type Channel struct {
	Title       string `xml:"title"`
	Description string `xml:"description"`
	Link        string `xml:"link"`
	Items       []Item `xml:"item"`
}

type RSS struct {
	XMLName xml.Name `xml:"rss"`
	Channel Channel  `xml:"channel"`
}

func GetRssFeed() (RSS, error) {
	//get the root url of the rss feed from .env
	rssRootUrl := os.Getenv("ARTICLE_ROOT_URL")
	if rssRootUrl == "" {
		panic("ARTICLE_ROOT_URL not set")
	}

	resp, err := http.Get(rssRootUrl)
	if err != nil {
		return RSS{}, err
	}
	defer resp.Body.Close()

	rss := RSS{}
	err = xml.NewDecoder(resp.Body).Decode(&rss)
	if err != nil {
		return RSS{}, err
	}

	return rss, nil

}

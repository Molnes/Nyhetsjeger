package articles

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

// Some sample articles for testing purposes.

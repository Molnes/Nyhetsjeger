package ai

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/Molnes/Nyhetsjeger/internal/models/articles"
	openai "github.com/sashabaranov/go-openai"
)

// GetJsonQuestions takes an article and an API key and returns a question and error.
func GetJsonQuestions(c context.Context, article articles.ArticleSMP, apiKey string) (Question, error) {

	client := openai.NewClient(apiKey)
	resp, err := client.CreateChatCompletion(c,
		openai.ChatCompletionRequest{
			Model: openai.GPT4Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: combinePromptAndArticle(article),
				},
			},
		},
	)
	if err != nil {
		return Question{}, err
	}

	question, err := ParseJsonQuestion(c, resp.Choices[0].Message.Content)
	if err != nil {
		return Question{}, err
	}

	return question, nil

}

// combinePromptAndArticle combines the prompt with the article text.
func combinePromptAndArticle(article articles.ArticleSMP) string {
	prompt := `make a quiz question out of the provided article.
It should be outputted as json in this format:
{
question: "",
alternatives: [
    {alternative_text:"",
     correct: bool 
} 
] 
}

There should be at the minimum 2 to maximum 4 alternatives. 
No other text should be returned. 
Both the question and the alternatives should be based on the provided article, but do not mention that the question is based on an article.
Make sure the question contains enough context that anyone who has previously read the article knows to which article the question refers.
The alternatives should not be longer than 50 characters.
And it should be in norwegian, if the article is bokmål, the question should be in bokmål.
If the article is nynorsk, the question should be in nynorsk.
Article:`

	articleText := ""
	for _, component := range article.Components {
		for _, paragraph := range component.Paragraphs {
			articleText += paragraph.Text.Value + ""
		}
	}
	return prompt + articleText
}

// Question is a struct that represents a question and its alternatives.
type Question struct {
	Question     string        `json:"question"`
	Alternatives []Alternative `json:"alternatives"`
}

// Alternative is a struct that represents an alternative to a question.
type Alternative struct {
	AlternativeText string `json:"alternative_text"`
	Correct         bool   `json:"correct"`
}

// ParseJsonQuestions parses the JSON string and returns a Question struct.
func ParseJsonQuestion(c context.Context, jsonMessage string) (Question, error) {
	question := Question{}

	// remove all ` from the json string
	jsonMessage = strings.ReplaceAll(jsonMessage, "`", "")

	// remove all "json" from the json string
	jsonMessage = strings.ReplaceAll(jsonMessage, "json", "")

	err := json.Unmarshal([]byte(jsonMessage), &question)
	if err != nil {
		return Question{}, err
	}
	return question, nil

}

package quizes

import (
	"github.com/Molnes/Nyhetsjeger/internal/data/articles"
	"github.com/Molnes/Nyhetsjeger/internal/data/questions"
	"github.com/google/uuid"
)

type Quiz struct {
	ID        uuid.UUID
	Title     string
	Questions []questions.Question
	Articles  []articles.Article
}

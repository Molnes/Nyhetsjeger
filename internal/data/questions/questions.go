package questions

import (
	"database/sql"

	"github.com/Molnes/Nyhetsjeger/internal/data/articles"
	"github.com/google/uuid"
)

type Question struct {
	ID           uuid.UUID
	Text         string
	Arrangement  uint
	ArticleID    articles.Article // The article this question is based on.
	QuizID       uuid.UUID
	Points       uint
	Alternatives []Alternative
}

type Alternative struct {
	ID         uuid.UUID
	Text       string
	IsCorrect  bool
	QuestionID uuid.UUID
}

// Returns all questions for a given quiz.
func GetAllQuestions(quizID uuid.UUID) ([]Question, error) {
	return SampleQuestions, nil
}

// Returns a question by its ID.
func GetQuestion(quizID uuid.UUID) (Question, error) {
	return SampleQuestions[0], nil
}

// Returns all questions for a given quiz.
// Includes the alternatives for each question.
func GetQuestionsByQuizID(db *sql.DB, id uuid.UUID) (*[]Question, error) {
	rows, err := db.Query(
		`SELECT
				q.id, q.question, q.arrangement, q.article_id, q.quiz_id, q.points,
				a.id, a.text, a.correct, a.question_id
			FROM
				questions q
			LEFT JOIN
				alternatives a ON q.id = a.question_id
			WHERE
				quiz_id = $1`,
		id)
	if err != nil {
		return nil, err
	}
	defer rows.Close() // Close the rows when the function returns

	return scanQuestionsFromFullRows(rows)
}

// Scans questions from full rows.
func scanQuestionsFromFullRows(rows *sql.Rows) (*[]Question, error) {
	questionsMap := make(map[uuid.UUID]*Question)

	for rows.Next() {
		var q Question
		var a Alternative
		err := rows.Scan(
			&q.ID, &q.Text, &q.Arrangement, &q.ArticleID, &q.QuizID, &q.Points,
			&a.ID, &a.Text, &a.IsCorrect, &a.QuestionID,
		)
		if err != nil {
			return nil, err
		}

		if question, ok := questionsMap[q.ID]; ok {
			question.Alternatives = append(question.Alternatives, a)
		} else {
			q.Alternatives = append(q.Alternatives, a)
			questionsMap[q.ID] = &q
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	questions := []Question{}
	for _, question := range questionsMap {
		questions = append(questions, *question)
	}

	return &questions, nil
}

var SampleQuestions []Question = []Question{
	{
		ID:   uuid.New(),
		Text: "What is the capital of Norway?",
		Alternatives: []Alternative{
			{
				ID:        uuid.New(),
				Text:      "Oslo",
				IsCorrect: true,
			},
			{
				ID:        uuid.New(),
				Text:      "Stockholm",
				IsCorrect: false,
			},
			{
				ID:        uuid.New(),
				Text:      "Copenhagen",
				IsCorrect: false,
			},
			{
				ID:        uuid.New(),
				Text:      "Helsinki",
				IsCorrect: false,
			},
		},
		Points:    10,
		ArticleID: articles.SampleArticles[0],
	},
	{
		ID:   uuid.New(),
		Text: "What is the capital of Sweden?",
		Alternatives: []Alternative{
			{
				ID:        uuid.New(),
				Text:      "Oslo",
				IsCorrect: false,
			},
			{
				ID:        uuid.New(),
				Text:      "Stockholm",
				IsCorrect: true,
			},
			{
				ID:        uuid.New(),
				Text:      "Copenhagen",
				IsCorrect: false,
			},
			{
				ID:        uuid.New(),
				Text:      "Helsinki",
				IsCorrect: false,
			},
		},
		Points:    10,
		ArticleID: articles.SampleArticles[0],
	},
}

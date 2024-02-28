package questions

import (
	"database/sql"

	"github.com/Molnes/Nyhetsjeger/internal/data/articles"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type Question struct {
	ID           uuid.UUID
	Text         string
	Arrangement  uint
	Article      articles.Article // The article this question is based on.
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
// Includes the article for each question.
// Includes the alternatives for each question.
func GetQuestionsByQuizID(db *sql.DB, id uuid.UUID) (*[]Question, error) {
	rows, err := db.Query(
		`SELECT
				q.id, q.question, q.arrangement, q.article_id, q.quiz_id, q.points,
				array_agg(a.id)
			FROM
				questions q
			LEFT JOIN
				answer_alternatives a ON q.id = a.question_id
			WHERE
				quiz_id = $1
			GROUP BY
				q.id`,
		id)
	if err != nil {
		return nil, err
	}
	defer rows.Close() // Close the rows when the function returns

	return scanQuestionsFromFullRows(db, rows)
}

// Scans questions from full rows.
func scanQuestionsFromFullRows(db *sql.DB, rows *sql.Rows) (*[]Question, error) {
	questions := []Question{}

	for rows.Next() {
		var q Question
		var articleID uuid.UUID
		var alternativeIDs []uuid.UUID
		err := rows.Scan(
			&q.ID, &q.Text, &q.Arrangement, &articleID, &q.QuizID, &q.Points,
			pq.Array(&alternativeIDs),
		)
		if err != nil {
			return nil, err
		}

		// Print the question
		println("----------- NEW LINE -----------")

		println("Question ID: " + q.ID.String())
		println("Text: " + q.Text)
		println("Points: " + string(q.Points))
		println("Article ID: " + articleID.String())

		// Map the article to the question
		article, _ := articles.GetArticleByID(db, articleID)
		q.Article = *article

	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	println("Questions: " + questions[0].Text)

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
		Points:  10,
		Article: articles.SampleArticles[0],
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
		Points:  10,
		Article: articles.SampleArticles[0],
	},
}

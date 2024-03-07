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



func GetFirstQuestion(db *sql.DB, quizID uuid.UUID) (*Question, error) {
        rows:= db.QueryRow(
                `SELECT
                                id, question, arrangement, article_id, quiz_id, points
                        FROM
                                questions
                        WHERE
                                quiz_id = $1 AND arrangement = 1`,
                quizID)

        return scanQuestionFromFullRow(db, rows)
}

// function for getting the next question in a quiz if there is one based on the arrangement, returns nil if there is no next question or an error
// takes *sql.DB 
// takes the current question UUID to get the next question
func GetNextQuestion(db *sql.DB,  currentQuestionID uuid.UUID) (*Question, error) {
        currentQuestion, err := GetQuestionByID(db, currentQuestionID)
        if err != nil {
                return nil, err
        }
        rows:= db.QueryRow(
                `SELECT
                        id, question, arrangement, article_id, quiz_id, points
                FROM
                        questions
                WHERE
                        quiz_id = $1 AND arrangement = $2`,
                currentQuestion.QuizID, currentQuestion.Arrangement + 1)

        return scanQuestionFromFullRow(db, rows)
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

// Converts a row from the database to a list of questions
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

		// Add the article to the question
		article, _ := articles.GetArticleByID(db, articleID)
		if article != nil {
			q.Article = *article
		}

		// Add the alternatives to the question
		for _, id := range alternativeIDs {
			alternative, _ := GetAlternativeByID(db, id)
			if alternative != nil {
				q.Alternatives = append(q.Alternatives, *alternative)
			}
		}

		// Add the question to the list of questions
		questions = append(questions, q)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &questions, nil
}

// Get specific question by ID.
func GetQuestionByID(db *sql.DB, id uuid.UUID) (*Question, error) {
        row := db.QueryRow(
                `SELECT
                                id, question, arrangement, article_id, quiz_id, points
                        FROM
                                questions
                        WHERE
                                id = $1`,
                id)

        return scanQuestionFromFullRow(db, row)
}

// Convert a row from the database to a Question.
func scanQuestionFromFullRow(db *sql.DB, row *sql.Row) (*Question, error) {
        var q Question
        var articleID uuid.UUID
        err := row.Scan(
                &q.ID, &q.Text, &q.Arrangement, &articleID, &q.QuizID, &q.Points,
        )

        if err != nil {
                return nil, err
        }

        // Add the article to the question
        article, _ := articles.GetArticleByID(db, articleID)
        if article != nil {
                q.Article = *article
        }

        // Add the alternatives to the question
        alternatives, _ := GetAlternativesByQuestionID(db, q.ID)
        q.Alternatives = *alternatives

        return &q, nil
}

// Get all alternatives for a given question.
func GetAlternativesByQuestionID(db *sql.DB, id uuid.UUID) (*[]Alternative, error) {
        rows, err := db.Query(
                `SELECT
                                id, text, correct, question_id
                        FROM
                                answer_alternatives
                        WHERE
                                question_id = $1`,
                id)
        if err != nil {
                return nil, err
        }
        defer rows.Close() // Close the rows when the function returns

        return scanAlternativesFromFullRows(rows)
}

// Convert a row from the database to a list of alternatives
func scanAlternativesFromFullRows(rows *sql.Rows) (*[]Alternative, error) {
        alternatives := []Alternative{}

        for rows.Next() {
                var a Alternative
                err := rows.Scan(
                        &a.ID, &a.Text, &a.IsCorrect, &a.QuestionID,
                )
                if err != nil {
                        return nil, err
                }

                // Add the alternative to the list of alternatives
                alternatives = append(alternatives, a)
        }

        if err := rows.Err(); err != nil {
                return nil, err
        }

        return &alternatives, nil
}

// Return alternative for a given question.
func GetAlternativeByID(db *sql.DB, id uuid.UUID) (*Alternative, error) {
	row := db.QueryRow(
		`SELECT
				id, text, correct, question_id
			FROM
				answer_alternatives
			WHERE
				id = $1`,
		id)

	return scanAlternativeFromFullRow(row)
}

// Convert a row from the database to an Alternative.
func scanAlternativeFromFullRow(row *sql.Row) (*Alternative, error) {
	var a Alternative
	err := row.Scan(
		&a.ID,
		&a.Text,
		&a.IsCorrect,
		&a.QuestionID,
	)

	if err != nil {
		return nil, err
	}
	return &a, nil
}

// Post a new question to the database.
// Returns the ID of the new question.
func PostNewQuestion(db *sql.DB, question Question) (uuid.UUID, error) {
	// Insert the question into the database
	db.QueryRow(
		`INSERT INTO questions (id, question, article_id, quiz_id, points)
		VALUES ($1, $2, $3, $4, $5);`,
		question.ID, question.Text, question.Article.ID, question.QuizID, question.Points,
	)

	return question.ID, nil
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
		Arrangement: 0,
		QuizID: uuid.New(),
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
		Arrangement: 1,
		QuizID: uuid.New(),
	},
}

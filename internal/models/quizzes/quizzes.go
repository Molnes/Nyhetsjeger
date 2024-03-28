package quizzes

import (
	"database/sql"
	"fmt"

	"net/url"
	"time"

	"github.com/Molnes/Nyhetsjeger/internal/models/questions"
	data_handling "github.com/Molnes/Nyhetsjeger/internal/utils/data"
	"github.com/google/uuid"
)

type Quiz struct {
	ID             uuid.UUID
	Title          string
	ImageURL       url.URL
	AvailableFrom  time.Time
	AvailableTo    time.Time
	CreatedAt      time.Time
	LastModifiedAt time.Time
	Published      bool
	IsDeleted      bool
	Questions      []questions.Question
}

type PartialQuiz struct {
	ID             uuid.UUID
	Title          string
	ImageURL       url.URL
	AvailableFrom  time.Time
	AvailableTo    time.Time
	Published      bool
	QuestionNumber uint
	MaxScore       uint
}

// Create a default quiz.
// This function is used to create a new quiz with default values.
func CreateDefaultQuiz() Quiz {
	tn := time.Now().Local()
	_, week := tn.ISOWeek()

	return Quiz{
		ID:    uuid.New(),
		Title: fmt.Sprintf("Quiz: Uke %d", week),
		ImageURL: url.URL{
			Scheme: "https",
			Host:   "upload.wikimedia.org",
			Path:   "/wikipedia/commons/5/59/Question_mark_choice.jpg",
		},
		AvailableFrom:  time.Now(),
		AvailableTo:    time.Now().Add(24 * 7 * time.Hour),
		CreatedAt:      time.Now(),
		LastModifiedAt: time.Now(),
		Published:      false,
		IsDeleted:      false,
		Questions:      []questions.Question{},
	}
}

// Retrieves a quiz from the database by its ID.
// Includes the questions for the quiz.
// Includes the articles for each question.
// Includes the alternatives for each question.
func GetQuizByID(db *sql.DB, id uuid.UUID) (*Quiz, error) {
	row := db.QueryRow(
		`SELECT
			id, title, image_url, available_from, available_to, created_at, last_modified_at, published, is_deleted
    FROM
			quizzes
		WHERE
			id = $1 AND
			is_deleted = false`,
		id)

	quiz, err := scanQuizFromFullRow(row)
	if err != nil {
		return nil, err
	}

	tempQuestions, err := questions.GetQuestionsByQuizID(db, id)
	if err != nil {
		return nil, err
	}
	quiz.Questions = *tempQuestions

	return quiz, nil
}

// Update the image URL for a quiz by its ID.
func UpdateImageByQuizID(db *sql.DB, id uuid.UUID, imageURL url.URL) error {
	_, err := db.Exec(
		`UPDATE quizzes
		SET image_url = $1
		WHERE id = $2`,
		imageURL.String(),
		id)
	return err
}

// Remove the image URL for a quiz by its ID.
func RemoveImageByQuizID(db *sql.DB, id uuid.UUID) error {
	_, err := db.Exec(
		`UPDATE quizzes
		SET image_url = NULL
		WHERE id = $1`,
		id)
	return err
}

// Update the title for a quiz by its ID.
func UpdateTitleByQuizID(db *sql.DB, id uuid.UUID, title string) error {
	_, err := db.Exec(
		`UPDATE quizzes
		SET title = $1
		WHERE id = $2`,
		title,
		id)
	return err
}

// Get all quizzes in the database.
func GetQuizzes(db *sql.DB) ([]Quiz, error) {
	rows, err := db.Query(
		`SELECT
			id, title, image_url, available_from, available_to, created_at, last_modified_at, published, is_deleted
    FROM
			quizzes
		WHERE
			is_deleted = false`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	quizzes := []Quiz{}
	for rows.Next() {
		var quiz Quiz
		var imageURL sql.NullString
		err := rows.Scan(
			&quiz.ID,
			&quiz.Title,
			&imageURL,
			&quiz.AvailableFrom,
			&quiz.AvailableTo,
			&quiz.CreatedAt,
			&quiz.LastModifiedAt,
			&quiz.Published,
			&quiz.IsDeleted,
		)
		if err != nil {
			return nil, err
		}

		// Set image URL
		tempURL, err := data_handling.ConvertNullStringToURL(&imageURL)
		if err != nil {
			return nil, err
		}
		quiz.ImageURL = *tempURL

		quizzes = append(quizzes, quiz)
	}
	return quizzes, nil
}

func GetUnfinishedQuizzesByUserID(db *sql.DB, userID uuid.UUID) ([]Quiz, error) {
	rows, err := db.Query(
		`SELECT 

        quizzes.id,
        quizzes.title,
        quizzes.image_url,
        quizzes.available_from,
        quizzes.available_to,
        quizzes.created_at,
        quizzes.last_modified_at,
        quizzes.published,
        quizzes.is_deleted



         FROM quizzes
        JOIN (


        SELECT *
        FROM quizzes
        JOIN
          (SELECT COUNT(id) AS existing_questions,
                  quiz_id
           FROM questions
           GROUP BY quiz_id) AS q ON quizzes.id = q.quiz_id
        JOIN
          (
        SELECT user_id,
        quiz_id,
        COUNT(answered_at) AS answered_questions
           FROM user_answers
           JOIN
             (SELECT *
              FROM questions) AS questions ON questions.id = user_answers.question_id
        WHERE chosen_answer_alternative_id IS NOT NULL
        GROUP BY user_id, quiz_id
        
        ) AS a ON a.quiz_id = quizzes.id
        
        ) AS counting ON counting.id = quizzes.id
        WHERE answered_questions < existing_questions
        AND user_id = $1
        `, userID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	quizzes := []Quiz{}
	for rows.Next() {
		var quiz Quiz
		var imageURL sql.NullString
		err := rows.Scan(
			&quiz.ID,
			&quiz.Title,
			&imageURL,
			&quiz.AvailableFrom,
			&quiz.AvailableTo,
			&quiz.CreatedAt,
			&quiz.LastModifiedAt,
			&quiz.Published,
			&quiz.IsDeleted,
		)
		if err != nil {
			return nil, err
		}

		// Set image URL
		tempURL, err := data_handling.ConvertNullStringToURL(&imageURL)
		if err != nil {
			return nil, err
		}
		quiz.ImageURL = *tempURL

		quizzes = append(quizzes, quiz)
	}
	return quizzes, nil

}


func GetFinishedQuizzesByUserID(db *sql.DB, userID uuid.UUID) ([]Quiz, error) {
        rows, err := db.Query(
                 
		`SELECT 

        quizzes.id,
        quizzes.title,
        quizzes.image_url,
        quizzes.available_from,
        quizzes.available_to,
        quizzes.created_at,
        quizzes.last_modified_at,
        quizzes.published,
        quizzes.is_deleted



         FROM quizzes
        JOIN (


        SELECT *
        FROM quizzes
        JOIN
          (SELECT COUNT(id) AS existing_questions,
                  quiz_id
           FROM questions
           GROUP BY quiz_id) AS q ON quizzes.id = q.quiz_id
        JOIN
          (
        SELECT user_id,
        quiz_id,
        COUNT(answered_at) AS answered_questions
           FROM user_answers
           JOIN
             (SELECT *
              FROM questions) AS questions ON questions.id = user_answers.question_id
        WHERE chosen_answer_alternative_id IS NOT NULL
        GROUP BY user_id, quiz_id
        
        ) AS a ON a.quiz_id = quizzes.id
        
        ) AS counting ON counting.id = quizzes.id
        WHERE answered_questions = existing_questions
        AND user_id = $1
        `, userID)

        
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	quizzes := []Quiz{}
	for rows.Next() {
		var quiz Quiz
		var imageURL sql.NullString
		err := rows.Scan(
			&quiz.ID,
			&quiz.Title,
			&imageURL,
			&quiz.AvailableFrom,
			&quiz.AvailableTo,
			&quiz.CreatedAt,
			&quiz.LastModifiedAt,
			&quiz.Published,
			&quiz.IsDeleted,
		)
		if err != nil {
			return nil, err
		}

		// Set image URL
		tempURL, err := data_handling.ConvertNullStringToURL(&imageURL)
		if err != nil {
			return nil, err
		}
		quiz.ImageURL = *tempURL

		quizzes = append(quizzes, quiz)
	}
	return quizzes, nil

}


func GetNonPublishedQuizzes(db *sql.DB) ([]Quiz, error) {
	return GetQuizzes(db)
}

func GetAllPublishedQuizzes(db *sql.DB) ([]Quiz, error) {
	quizzes, err := GetQuizzes(db)
	if err != nil {
		return nil, err
	}
	quizzes = append(quizzes, quizzes...)
	quizzes = append(quizzes, quizzes...)
	return quizzes, nil
}

// Converts a row from the database to a Quiz.
func scanQuizFromFullRow(row *sql.Row) (*Quiz, error) {
	var quiz Quiz
	var imageURL sql.NullString
	err := row.Scan(
		&quiz.ID,
		&quiz.Title,
		&imageURL,
		&quiz.AvailableFrom,
		&quiz.AvailableTo,
		&quiz.CreatedAt,
		&quiz.LastModifiedAt,
		&quiz.Published,
		&quiz.IsDeleted,
	)
	if err != nil {
		return nil, err
	}

	// Set image URL
	tempURL, err := data_handling.ConvertNullStringToURL(&imageURL)
	if err != nil {
		return nil, err
	}
	quiz.ImageURL = *tempURL

	if err == sql.ErrNoRows {
		return nil, err
	}
	return &quiz, err
}

// Create a Quiz in the DB.
func CreateQuiz(db *sql.DB, quiz Quiz) (*uuid.UUID, error) {
	_, err := db.Exec(
		`INSERT INTO quizzes
			(id, title, image_url, available_from, available_to, created_at, last_modified_at, published, is_deleted)
		VALUES
			($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		quiz.ID,
		quiz.Title,
		quiz.ImageURL.String(),
		quiz.AvailableFrom,
		quiz.AvailableTo,
		quiz.CreatedAt,
		quiz.LastModifiedAt,
		quiz.Published,
		quiz.IsDeleted,
	)

	return &quiz.ID, err
}

// Set a Quiz to deleted in the DB by its ID.
func DeleteQuizByID(db *sql.DB, id uuid.UUID) error {
	_, err := db.Exec(
		`UPDATE quizzes
		SET is_deleted = true
		WHERE id = $1`,
		id)
	return err
}

// Update the published status of a quiz by its ID.
func UpdatePublishedStatusByQuizID(db *sql.DB, id uuid.UUID, published bool) error {
	_, err := db.Exec(
		`UPDATE quizzes
		SET published = $1
		WHERE id = $2`,
		published,
		id)
	return err
}

// Retrieves a partial quiz from the database by a quiz ID.
func GetPartialQuizByID(db *sql.DB, quizid uuid.UUID) (*PartialQuiz, error) {
	row := db.QueryRow(
		`SELECT qz.id, qz.title, qz.image_url, qz.available_from, qz.available_to, qz.published, count(q.id), sum(q.points)
		FROM quizzes qz 
		LEFT JOIN questions q ON q.quiz_id = qz.id
		WHERE qz.id = $1
		GROUP BY qz.id;`, quizid)

	var pq PartialQuiz
	var imageURLStr sql.NullString
	err := row.Scan(
		&pq.ID,
		&pq.Title,
		&imageURLStr,
		&pq.AvailableFrom,
		&pq.AvailableTo,
		&pq.Published,
		&pq.QuestionNumber,
		&pq.MaxScore,
	)
	if err != nil {
		return nil, err
	}

	tempURL, err := data_handling.ConvertNullStringToURL(&imageURLStr)
	if err != nil {
		return nil, err
	}
	pq.ImageURL = *tempURL

	return &pq, nil
}

// Update the quiz's 'active' start time by its ID.
func UpdateActiveStartByQuizID(db *sql.DB, id uuid.UUID, activeStart time.Time) error {
	_, err := db.Exec(
		`UPDATE quizzes
		SET available_from = $1
		WHERE id = $2`,
		activeStart,
		id)
	return err
}

// Update the quiz's 'active' end time by its ID.
func UpdateActiveEndByQuizID(db *sql.DB, id uuid.UUID, activeEnd time.Time) error {
	_, err := db.Exec(
		`UPDATE quizzes
		SET available_to = $1
		WHERE id = $2`,
		activeEnd,
		id)
	return err
}

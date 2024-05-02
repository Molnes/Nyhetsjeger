package labels

import (
	"database/sql"

	"time"

	"github.com/google/uuid"
)

// Label struct
type Label struct {
        ID        uuid.UUID `json:"id"`
        Name      string    `json:"name"`
        CreatedAt time.Time `json:"created_at"`
        Active    bool      `json:"active"`
}

// GetLabels returns a list of all labels in the database.
func GetLabels(db *sql.DB) ([]Label, error) {
        rows, err := db.Query("SELECT id, name, created_at, is_active FROM labels")
        if err != nil {
                return nil, err
        }
        defer rows.Close()

        labels := []Label{}
        for rows.Next() {
                var label Label
                if err := rows.Scan(&label.ID, &label.Name, &label.CreatedAt, &label.Active); err != nil {
                        return nil, err
                }
                labels = append(labels, label)
        }
        return labels, nil
}

// GetLabelByID returns a label by its ID.
func GetLabelByID(db *sql.DB, id uuid.UUID) (Label, error) {
        var label Label
        err := db.QueryRow("SELECT id, name, created_at, is_active FROM labels WHERE id=$1", id).Scan(&label.ID, &label.Name, &label.CreatedAt, &label.Active)
        if err != nil {
                return label, err
        }
        return label, nil
}

// CreateLabel creates a new label in the database.
func CreateLabel(db *sql.DB, name string) (uuid.UUID, error) {
        id := uuid.New()
        _, err := db.Exec("INSERT INTO labels (id, name, created_at, is_active) VALUES ($1, $2, $3, $4)", id, name, time.Now(), true)
        if err != nil {
                return uuid.Nil, err
        }
        return id, nil
}

// GetLabelByQuizzID returns a list of labels that are associated with the given quizz.
// It will return an empty list if the quizz is not associated with any label.
func GetLabelByQuizzID(db *sql.DB, quizzID uuid.UUID) ([]Label, error) {
        rows, err := db.Query("SELECT l.id, l.name, l.created_at, l.is_active FROM labels l JOIN quiz_labels ql ON l.id = ql.label_id WHERE ql.quiz_id=$1", quizzID)
        if err != nil {
                return nil, err
        }
        defer rows.Close()

        labels := []Label{}
        for rows.Next() {
                var label Label
                if err := rows.Scan(&label.ID, &label.Name, &label.CreatedAt, &label.Active); err != nil {
                        return nil, err
                }
                labels = append(labels, label)
        }
        return labels, nil
}

// GetQuizzesByLabelID returns a list of quizz IDs that are associated with the given label.
// It will return an empty list if the label is not associated with any quizz.
func GetQuizzesByLabelID(db *sql.DB, labelID uuid.UUID) ([]uuid.UUID, error) {
        rows, err := db.Query("SELECT quiz_id FROM quiz_labels WHERE label_id=$1", labelID)
        if err != nil {
                return nil, err
        }
        defer rows.Close()

        quizzIDs := []uuid.UUID{}
        for rows.Next() {
                var quizzID uuid.UUID
                if err := rows.Scan(&quizzID); err != nil {
                        return nil, err
                }
                quizzIDs = append(quizzIDs, quizzID)
        }
        return quizzIDs, nil
}

// AddLabelToQuizz adds a label to a quizz.
// It will fail if the label is already associated with the quizz.
func AddLabelToQuizz(db *sql.DB, quizzID, labelID uuid.UUID) error {
        _, err := db.Exec("INSERT INTO quiz_labels (quiz_id, label_id) VALUES ($1, $2)", quizzID, labelID)
        if err != nil {
                return err
        }
        return nil
}

// RemoveLabelFromQuizz removes a label from a quizz.
// It will fail if the label is not associated with the quizz.
func RemoveLabelFromQuizz(db *sql.DB, quizzID, labelID uuid.UUID) error {
        _, err := db.Exec("DELETE FROM quiz_labels WHERE quiz_id=$1 AND label_id=$2", quizzID, labelID)
        if err != nil {
                return err
        }
        return nil
}

// RemoveLabel removes a label from the database.
// It will fail if the label does not exist or if the label is still in use.
func RemoveLabel(db *sql.DB, id uuid.UUID) error {
        _, err := db.Exec("DELETE FROM labels WHERE id=$1", id)
        if err != nil {
                return err
        }
        return nil
}

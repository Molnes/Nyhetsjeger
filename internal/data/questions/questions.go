package questions

import "github.com/google/uuid"

type Question struct {
	ID           uuid.UUID
	Text         string
	Alternatives []Alternative
	Points       uint
}

type Alternative struct {
	ID        uuid.UUID
	Text      string
	IsCorrect bool
}

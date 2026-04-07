package utils

import (
	"github.com/google/uuid"
)

// NewID generates a time-ordered UUID v7.
// It returns a 16-byte uuid.UUID instead of a string to achieve zero-allocation
// in Go and maximum index performance in the database.
func NewID() uuid.UUID {
	return uuid.Must(uuid.NewV7())
}

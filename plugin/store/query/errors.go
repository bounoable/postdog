package query

import (
	"fmt"

	"github.com/google/uuid"
)

// LetterNotFoundError is returned from a `query.Find()` if the underlying repository can't retrieve a letter by it's ID.
type LetterNotFoundError struct {
	ID  uuid.UUID
	Err error
}

func (err LetterNotFoundError) Unwrap() error {
	return err.Err
}

func (err LetterNotFoundError) Error() string {
	return fmt.Sprintf("letter not found: %s: %v", err.ID, err.Err)
}

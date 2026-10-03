package listing

import (
	"time"

	"github.com/google/uuid"
)

type Listing struct {
	ID          uuid.UUID
	Title       string
	Description string
	Price       int64
	City        string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type CreateInput struct {
	Title       string
	Description string
	Price       int64
	City        string
}

// UpdateInput replaces all editable fields, using the same fields as creation.
type UpdateInput CreateInput

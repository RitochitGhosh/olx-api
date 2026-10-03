package listing

import (
	"context"

	"github.com/google/uuid"
)

// Repository defines storage operations without tying the service to PostgreSQL.
type Repository interface {
	Create(ctx context.Context, input CreateInput) (Listing, error)
	Update(ctx context.Context, id uuid.UUID, input UpdateInput) (Listing, error)
	List(ctx context.Context, limit int) ([]Listing, error)
	FindByID(ctx context.Context, id uuid.UUID) (Listing, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

package listing

import (
	"context"
	"strings"

	"github.com/google/uuid"
)

// ValidationError separates invalid input from unexpected database failures.
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

// Service validates input and applies business rules before calling the repository.
type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (Listing, error) {
	input, err := validateInput(input)
	if err != nil {
		return Listing{}, err
	}
	return s.repo.Create(ctx, input)
}

// Update validates a complete replacement before writing to storage.
func (s *Service) Update(ctx context.Context, id uuid.UUID, input UpdateInput) (Listing, error) {
	if id == uuid.Nil {
		return Listing{}, &ValidationError{Message: "invalid listing ID"}
	}
	validated, err := validateInput(CreateInput(input))
	if err != nil {
		return Listing{}, err
	}
	return s.repo.Update(ctx, id, UpdateInput(validated))
}

// Share normalization and validation so create and update follow the same rules.
func validateInput(input CreateInput) (CreateInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)
	input.City = strings.TrimSpace(input.City)

	if input.Title == "" {
		return CreateInput{}, &ValidationError{
			Message: "title is required",
		}
	}

	if input.Description == "" {
		return CreateInput{}, &ValidationError{
			Message: "description is required",
		}
	}

	if input.City == "" {
		return CreateInput{}, &ValidationError{
			Message: "city is required",
		}
	}

	if input.Price <= 0 {
		return CreateInput{}, &ValidationError{
			Message: "price must be greater than zero",
		}
	}

	return input, nil
}

func (s *Service) List(ctx context.Context) ([]Listing, error) {
	const limit = 100

	return s.repo.List(ctx, limit)
}

func (s *Service) FindByID(ctx context.Context, id uuid.UUID) (Listing, error) {
	// Reject the zero UUID before querying the database.
	if id == uuid.Nil {
		return Listing{}, &ValidationError{Message: "invalid listing ID"}
	}

	return s.repo.FindByID(ctx, id)
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return &ValidationError{Message: "invalid listing ID"}
	}

	return s.repo.Delete(ctx, id)
}

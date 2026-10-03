package listing

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
)

// PostgresRepository implements listing storage using parameterized SQL.
type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{
		db: db,
	}
}

func (r *PostgresRepository) Create(ctx context.Context, input CreateInput) (Listing, error) {
	var l Listing

	// New listings start as active; status is required by the database schema.
	if err := r.db.QueryRowContext(ctx, `
		INSERT INTO listings (title, description, price, city, status)
		VALUES ($1, $2, $3, $4, 'active')
		RETURNING id, title, description, price, city, created_at, updated_at
	`, input.Title, input.Description, input.Price, input.City).Scan(
		&l.ID,
		&l.Title,
		&l.Description,
		&l.Price,
		&l.City,
		&l.CreatedAt,
		&l.UpdatedAt,
	); err != nil {
		return Listing{}, err
	}

	return l, nil
}

// Update returns the saved listing in one query and preserves its creation time and status.
func (r *PostgresRepository) Update(ctx context.Context, id uuid.UUID, input UpdateInput) (Listing, error) {
	var l Listing
	err := r.db.QueryRowContext(ctx, `
		UPDATE listings
		SET title = $1, description = $2, price = $3, city = $4, updated_at = now()
		WHERE id = $5
		RETURNING id, title, description, price, city, created_at, updated_at
	`, input.Title, input.Description, input.Price, input.City, id).Scan(
		&l.ID, &l.Title, &l.Description, &l.Price, &l.City, &l.CreatedAt, &l.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return Listing{}, ErrNotFound
	}
	if err != nil {
		return Listing{}, err
	}
	return l, nil
}

func (r *PostgresRepository) List(ctx context.Context, limit int) ([]Listing, error) {
	// ID breaks timestamp ties so the limited results have a stable order.
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, title, description, price, city, created_at, updated_at
		FROM listings
		ORDER BY created_at DESC, id DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	listings := make([]Listing, 0)
	// Check rows.Err after iteration because Next also returns false on errors.
	for rows.Next() {
		var l Listing
		if err := rows.Scan(
			&l.ID,
			&l.Title,
			&l.Description,
			&l.Price,
			&l.City,
			&l.CreatedAt,
			&l.UpdatedAt,
		); err != nil {
			return nil, err
		}
		listings = append(listings, l)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return listings, nil
}

func (r *PostgresRepository) FindByID(ctx context.Context, id uuid.UUID) (Listing, error) {
	var l Listing

	err := r.db.QueryRowContext(ctx, `
		SELECT id, title, description, price, city, created_at, updated_at
		FROM listings
		WHERE id = $1
	`, id).Scan(&l.ID, &l.Title, &l.Description, &l.Price, &l.City, &l.CreatedAt, &l.UpdatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return Listing{}, ErrNotFound
	}

	if err != nil {
		return Listing{}, err
	}

	return l, nil
}

func (r *PostgresRepository) Delete(
	ctx context.Context,
	id uuid.UUID,
) error {
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM listings
		WHERE id = $1
	`, id)

	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return ErrNotFound
	}

	return nil
}

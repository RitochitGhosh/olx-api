package listing

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

// A small SQL driver exercises database/sql scanning without a running database.
type listingTestDriver struct{}

type listingTestConn struct{}

type listingTestRows struct {
	done bool
}

func (listingTestDriver) Open(string) (driver.Conn, error) { return listingTestConn{}, nil }
func (listingTestConn) Close() error                       { return nil }
func (listingTestConn) Begin() (driver.Tx, error)          { return nil, errors.New("not supported") }
func (listingTestConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("not supported")
}

func (listingTestConn) QueryContext(_ context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	if strings.Contains(query, "UPDATE listings") {
		if len(args) != 5 || !strings.Contains(query, "updated_at = now()") || !strings.Contains(query, "WHERE id = $5") {
			return nil, errors.New("update must bind its fields and ID and refresh updated_at")
		}
		return &listingTestRows{done: args[4].Value == uuid.Nil.String()}, nil
	}
	if strings.Contains(query, "INSERT INTO") {
		if !strings.Contains(query, "city, status)") || !strings.Contains(query, "'active'") {
			return nil, errors.New("insert must supply the required status")
		}
		if len(args) != 4 {
			return nil, errors.New("insert must supply four input arguments")
		}
		return &listingTestRows{}, nil
	}
	return &listingTestRows{done: true}, nil
}

func (listingTestConn) ExecContext(context.Context, string, []driver.NamedValue) (driver.Result, error) {
	return driver.RowsAffected(0), nil
}

func (*listingTestRows) Columns() []string {
	return []string{"id", "title", "description", "price", "city", "created_at", "updated_at"}
}

func (*listingTestRows) Close() error { return nil }

func (r *listingTestRows) Next(dest []driver.Value) error {
	if r.done {
		return io.EOF
	}
	r.done = true
	values := []driver.Value{
		"00000000-0000-0000-0000-000000000001", "Bike", "Used bike", int64(100), "Pune",
		time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC),
	}
	copy(dest, values)
	return nil
}

func init() {
	sql.Register("listing-test", listingTestDriver{})
}

func TestPostgresRepository(t *testing.T) {
	db, err := sql.Open("listing-test", "")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	repo := NewPostgresRepository(db)
	ctx := context.Background()

	created, err := repo.Create(ctx, CreateInput{Title: "Bike", Description: "Used bike", Price: 100, City: "Pune"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if created.ID == uuid.Nil || created.Title != "Bike" || created.Price != 100 || !created.UpdatedAt.After(created.CreatedAt) {
		t.Fatalf("incorrectly scanned listing: %+v", created)
	}
	if _, err := repo.FindByID(ctx, uuid.New()); !errors.Is(err, ErrNotFound) {
		t.Fatalf("lookup: expected ErrNotFound, got %v", err)
	}
	if err := repo.Delete(ctx, uuid.New()); !errors.Is(err, ErrNotFound) {
		t.Fatalf("delete: expected ErrNotFound, got %v", err)
	}
	updated, err := repo.Update(ctx, created.ID, UpdateInput{Title: "Bike", Description: "Used bike", Price: 100, City: "Pune"})
	if err != nil || updated.ID != created.ID || updated.Title != "Bike" || !updated.UpdatedAt.After(updated.CreatedAt) {
		t.Fatalf("update: got %+v, error %v", updated, err)
	}
	if _, err := repo.Update(ctx, uuid.Nil, UpdateInput{}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("missing update: expected ErrNotFound, got %v", err)
	}
}

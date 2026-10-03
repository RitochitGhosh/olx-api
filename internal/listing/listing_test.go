package listing

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/RitochitGhosh/olx-api/internal/middleware"
	"github.com/google/uuid"
)

// stubRepository keeps service and HTTP tests independent of PostgreSQL.
type stubRepository struct {
	Repository
	created   CreateInput
	updated   UpdateInput
	updatedID uuid.UUID
	err       error
}

func (r *stubRepository) Create(_ context.Context, input CreateInput) (Listing, error) {
	r.created = input
	return Listing{ID: uuid.New(), Title: input.Title}, r.err
}

func (r *stubRepository) List(context.Context, int) ([]Listing, error) {
	return nil, r.err
}

func (r *stubRepository) Update(_ context.Context, id uuid.UUID, input UpdateInput) (Listing, error) {
	r.updated, r.updatedID = input, id
	return Listing{ID: id, Title: input.Title, Description: input.Description, Price: input.Price, City: input.City}, r.err
}

func TestUpdateValidation(t *testing.T) {
	id := uuid.New()
	valid := UpdateInput{Title: " Bike ", Description: " Used bike ", City: " Pune ", Price: 100}
	for _, field := range []string{"title", "description", "city", "price", "id"} {
		t.Run(field, func(t *testing.T) {
			input, listingID := valid, id
			switch field {
			case "title":
				input.Title = " "
			case "description":
				input.Description = ""
			case "city":
				input.City = ""
			case "price":
				input.Price = 0
			case "id":
				listingID = uuid.Nil
			}
			repo := &stubRepository{}
			_, err := NewService(repo).Update(context.Background(), listingID, input)
			var validationErr *ValidationError
			if !errors.As(err, &validationErr) || repo.updatedID != uuid.Nil {
				t.Fatalf("invalid update reached storage or returned wrong error: %v", err)
			}
		})
	}
	repo := &stubRepository{}
	if _, err := NewService(repo).Update(context.Background(), id, valid); err != nil {
		t.Fatal(err)
	}
	want := UpdateInput{Title: "Bike", Description: "Used bike", City: "Pune", Price: 100}
	if repo.updated != want || repo.updatedID != id {
		t.Fatalf("unexpected update: %+v, ID %s", repo.updated, repo.updatedID)
	}
}

func (r *stubRepository) FindByID(context.Context, uuid.UUID) (Listing, error) {
	return Listing{}, r.err
}

func (r *stubRepository) Delete(context.Context, uuid.UUID) error {
	return r.err
}

func TestCreateValidation(t *testing.T) {
	valid := CreateInput{Title: " Bike ", Description: " Used bike ", City: " Pune ", Price: 100}
	tests := []struct {
		name  string
		input CreateInput
	}{
		{"empty title", CreateInput{Description: "bike", City: "Pune", Price: 100}},
		{"blank description", CreateInput{Title: "Bike", Description: " \t ", City: "Pune", Price: 100}},
		{"empty city", CreateInput{Title: "Bike", Description: "bike", Price: 100}},
		{"zero price", CreateInput{Title: "Bike", Description: "bike", City: "Pune"}},
		{"negative price", CreateInput{Title: "Bike", Description: "bike", City: "Pune", Price: -1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &stubRepository{}
			_, err := NewService(repo).Create(context.Background(), tt.input)
			var validationErr *ValidationError
			if !errors.As(err, &validationErr) {
				t.Fatalf("expected ValidationError, got %v", err)
			}
			if repo.created != (CreateInput{}) {
				t.Fatal("invalid input reached repository")
			}
		})
	}

	repo := &stubRepository{}
	if _, err := NewService(repo).Create(context.Background(), valid); err != nil {
		t.Fatal(err)
	}
	want := CreateInput{Title: "Bike", Description: "Used bike", City: "Pune", Price: 100}
	if repo.created != want {
		t.Fatalf("normalized input = %+v, want %+v", repo.created, want)
	}
}

func TestListingRoutes(t *testing.T) {
	id := uuid.New().String()
	updateBody := `{"title":"Bike","description":"Used bike","city":"Pune","price":100}`
	tests := []struct {
		name   string
		method string
		path   string
		body   string
		err    error
		status int
		match  string
	}{
		{"update success", "PUT", "/listings/" + id, updateBody, nil, 200, `"id":"` + id + `"`},
		{"missing update", "PUT", "/listings/" + id, updateBody, fmt.Errorf("update: %w", ErrNotFound), 404, `"code":"not_found"`},
		{"update invalid ID", "PUT", "/listings/invalid", updateBody, nil, 400, `"code":"invalid_id"`},
		{"update zero ID", "PUT", "/listings/" + uuid.Nil.String(), updateBody, nil, 400, `"code":"invalid_id"`},
		{"update invalid JSON", "PUT", "/listings/" + id, `{`, nil, 400, `"code":"invalid_json"`},
		{"update unknown field", "PUT", "/listings/" + id, `{"unknown":true}`, nil, 400, `"code":"invalid_json"`},
		{"update missing fields", "PUT", "/listings/" + id, `{"title":"Bike"}`, nil, 422, `"code":"validation_failed"`},
		{"update database error", "PUT", "/listings/" + id, updateBody, errors.New("private database details"), 500, `"code":"internal_error"`},
		{"missing listing", "GET", "/listings/" + id, "", fmt.Errorf("lookup: %w", ErrNotFound), 404, `"code":"not_found"`},
		{"missing delete", "DELETE", "/listings/" + id, "", ErrNotFound, 404, `"code":"not_found"`},
		{"invalid ID", "GET", "/listings/invalid", "", nil, 400, `"code":"invalid_id"`},
		{"zero ID lookup", "GET", "/listings/" + uuid.Nil.String(), "", nil, 400, `"code":"invalid_id"`},
		{"zero ID delete", "DELETE", "/listings/" + uuid.Nil.String(), "", nil, 400, `"code":"invalid_id"`},
		{"database error", "GET", "/listings/" + id, "", errors.New("private database details"), 500, `"code":"internal_error"`},
		{"empty list", "GET", "/listings", "", nil, 200, `"data":[]`},
		{"delete success", "DELETE", "/listings/" + id, "", nil, 204, ""},
		{"invalid JSON", "POST", "/listings", `{"title":`, nil, 400, `"code":"invalid_json"`},
		{"unknown field", "POST", "/listings", `{"unknown":true}`, nil, 400, `"code":"invalid_json"`},
		{"multiple objects", "POST", "/listings", `{} {}`, nil, 400, `"code":"invalid_json"`},
		{"invalid input", "POST", "/listings", `{}`, nil, 422, `"code":"validation_failed"`},
		{"create success", "POST", "/listings", `{"title":"Bike","description":"Used bike","city":"Pune","price":100}`, nil, 201, `"title":"Bike"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &stubRepository{err: tt.err}
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			mux := http.NewServeMux()
			RegisterRoutes(mux, NewHandler(NewService(repo), logger))
			w := httptest.NewRecorder()
			middleware.RequestId(mux).ServeHTTP(w, httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body)))
			if w.Code != tt.status || !strings.Contains(w.Body.String(), tt.match) {
				t.Fatalf("got %d %s; want status %d containing %q", w.Code, w.Body.String(), tt.status, tt.match)
			}
			if strings.Contains(w.Body.String(), "private database details") {
				t.Fatal("database error exposed to client")
			}
			if w.Header().Get("X-Request-Id") == "" {
				t.Fatal("response is missing its request ID")
			}
			if tt.status == http.StatusNoContent && w.Body.Len() != 0 {
				t.Fatal("204 response must have no body")
			}
		})
	}
}

package handlers

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/RitochitGhosh/olx-api/internal/middleware"
)

type ListingHandler struct {
	db     *sql.DB
	logger *slog.Logger
}

// Constructor Pattern
func NewListingHandler(db *sql.DB, logger *slog.Logger) *ListingHandler {
	return &ListingHandler{
		db:     db,
		logger: logger,
	}
}

// `json:"name" is calld struct tag. It tells Go’s JSON encoder/decoder what JSON key name to use for that field.
type listing struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Price       int64     `json:"price"`
	City        string    `json:"city"`
	CreatedAt   time.Time `json:"created_at"`
}

// Method Receiver
func (lh ListingHandler) FetchListings(w http.ResponseWriter, r *http.Request) {
	// generate request id

	// request scoped context
	ctx := r.Context()

	rows, err := lh.db.QueryContext(ctx, `
		SELECT id, title, description, price, city, created_at
		FROM listings
		ORDER BY created_at DESC
		LIMIT 100
	`)
	if err != nil {
		lh.logger.Error("fetch listenings failed", "error", err)
		http.Error(w, "failed to fetch listings", http.StatusNotFound)
		return
	}
	defer rows.Close()

	listings := []listing{}
	for rows.Next() {
		var l listing
		if err := rows.Scan(&l.ID, &l.Title, &l.Description, &l.Price, &l.City, &l.CreatedAt); err != nil {
			lh.logger.Error("rows scan error", "error", err)
			http.Error(w, "failed to scan listings", http.StatusInternalServerError)
			return
		}

		listings = append(listings, l)
	}

	if err := rows.Err(); err != nil {
		lh.logger.Error("iteration error", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// w.Write([]byte("all ok"))
	_ = json.NewEncoder(w).Encode(listings)
}

func (lh ListingHandler) DeleteListing(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestId := middleware.RequestIdFromContext(ctx)
	// requestId := ctx.Value("requestCtxId").(string)
	id := r.PathValue("id")

	lh.logger.Debug("debug log", "listing_id", id)

	_, err := lh.db.ExecContext(ctx, `
		DELETE FROM listings
		WHERE id = $1
	`, id)

	if err != nil {
		// log.Printf("db.ExecContext: %v", err)
		lh.logger.Error("delete failed", "listing_id", id, "request_Id", requestId, "err", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

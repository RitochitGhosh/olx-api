package handlers

import (
	"database/sql"
	"log/slog"
	"net/http"
	"time"
	"uuid"

	"github.com/RitochitGhosh/olx-api/internal/httpx"
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
func (lh *ListingHandler) FetchListings(w http.ResponseWriter, r *http.Request) {
	// request scoped context
	ctx := r.Context()

	// generate request id
	requestID := middleware.RequestIdFromContext(ctx)

	rows, err := lh.db.QueryContext(ctx, `
		SELECT id, title, description, price, city, created_at
		FROM listings
		ORDER BY created_at DESC
		LIMIT 100
	`)
	if err != nil {
		lh.logger.Error(
			"fetch listings failed",
			"request_id", requestID,
			"error", err,
		)
		httpx.Error(w, httpx.CodeInternalError, requestID)
		return
	}
	defer rows.Close()

	listings := []listing{}
	for rows.Next() {
		var l listing
		if err := rows.Scan(&l.ID, &l.Title, &l.Description, &l.Price, &l.City, &l.CreatedAt); err != nil {
			lh.logger.Error(
				"scan listing failed",
				"request_id", requestID,
				"error", err,
			)
			httpx.Error(w, httpx.CodeInternalError, requestID)
			return
		}

		listings = append(listings, l)
	}

	if err := rows.Err(); err != nil {
		lh.logger.Error(
			"listings iteration failed",
			"request_id", requestID,
			"error", err,
		)
		httpx.Error(w, httpx.CodeInternalError, requestID)
		return
	}

	httpx.Success(w, http.StatusOK, listings, requestID)
}

func (lh *ListingHandler) DeleteListing(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	requestID := middleware.RequestIdFromContext(ctx)
	// requestID := ctx.Value("requestCtxId").(string)
	id := r.PathValue("id")

	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		httpx.Error(w, httpx.CodeInvalidID, requestID)
		return
	}

	lh.logger.Debug("debug log", "listing_id", id)

	result, err := lh.db.ExecContext(ctx, `
		DELETE FROM listings
		WHERE id = $1
	`, id)

	if err != nil {
		// log.Printf("db.ExecContext: %v", err)
		lh.logger.Error(
			"delete listing failed",
			"listing_id", id,
			"request_Id", requestID,
			"error", err,
		)
		// http.Error(w, "internal server error", http.StatusInternalServerError)
		httpx.Error(w, httpx.CodeInternalError, requestID)
		return
	}

	affected, err := result.RowsAffected()
	if err != nil {
		lh.logger.Error(
			"failed to read affected rows",
			"listing_id", id,
			"request_id", requestID,
			"error", err,
		)
		httpx.Error(w, httpx.CodeInternalError, requestID)
		return
	}

	if affected == 0 {
		httpx.Error(w, httpx.CodeNotFound, requestID, "Listing not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

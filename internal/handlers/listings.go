package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type ListingHandler struct {
	db *sql.DB
}

// Constructor Pattern
func NewListingHandler(db *sql.DB) *ListingHandler {
	return &ListingHandler{
		db: db,
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
	rows, err := lh.db.Query(`
			SELECT id, title, description, price, city, created_at
			FROM listings
			ORDER BY created_at DESC
			LIMIT 100
		`)
	if err != nil {
		log.Printf("query: %v", err)
		http.Error(w, "failed to fetch listings", http.StatusNotFound)
		return
	}
	defer rows.Close()

	listings := []listing{}
	for rows.Next() {
		var l listing
		if err := rows.Scan(&l.ID, &l.Title, &l.Description, &l.Price, &l.City, &l.CreatedAt); err != nil {
			log.Printf("rows.Scan: %v", err)
			http.Error(w, "failed to scan listings", http.StatusInternalServerError)
			return
		}

		listings = append(listings, l)
	}

	if err := rows.Err(); err != nil {
		log.Printf("rows.Err: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// w.Write([]byte("all ok"))
	_ = json.NewEncoder(w).Encode(listings)
}

func (lh ListingHandler) DeleteListing(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	fmt.Println("id: ", id)

	result, err := lh.db.Exec(`
			DELETE FROM listings
			WHERE id = $1
		`, id)
	if err != nil {
		log.Printf("db.Exec: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	affected, err := result.RowsAffected()
	if err != nil {
		log.Printf("rows affected: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if affected == 0 {
		log.Printf("rows affected: %v", err)
		http.Error(w, "listing not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

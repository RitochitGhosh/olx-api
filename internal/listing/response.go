package listing

import "time"

type ListingResponse struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Price       int64     `json:"price"`
	City        string    `json:"city"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func toListingResponse(l Listing) ListingResponse {
	return ListingResponse{
		ID:          l.ID.String(),
		Title:       l.Title,
		Description: l.Description,
		Price:       l.Price,
		City:        l.City,
		CreatedAt:   l.CreatedAt,
		UpdatedAt:   l.UpdatedAt,
	}
}

func toListingResponses(listings []Listing) []ListingResponse {
	result := make([]ListingResponse, 0, len(listings))

	for _, listing := range listings {
		result = append(result, toListingResponse(listing))
	}

	return result
}

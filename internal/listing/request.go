package listing

type CreateListingRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Price       int64  `json:"price"`
	City        string `json:"city"`
}

// PUT requires all editable fields, just like a create request.
type UpdateListingRequest CreateListingRequest

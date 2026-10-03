package listing

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/RitochitGhosh/olx-api/internal/httpx"
	"github.com/RitochitGhosh/olx-api/internal/middleware"
	"github.com/google/uuid"
)

// Handler translates HTTP requests and service results into JSON responses.
type Handler struct {
	service *Service
	logger  *slog.Logger
}

func NewHandler(service *Service, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	requestID := middleware.RequestIdFromContext(ctx)

	var req CreateListingRequest

	if err := httpx.ReadJSON(w, r, &req); err != nil {
		h.logger.Warn(
			"invalid create listing request",
			"request_id", requestID,
			"error", err,
		)

		httpx.Error(
			w,
			httpx.CodeInvalidJSON,
			requestID,
			err.Error(),
		)
		return
	}

	input := CreateInput{
		Title:       req.Title,
		Description: req.Description,
		Price:       req.Price,
		City:        req.City,
	}

	created, err := h.service.Create(ctx, input)

	if err != nil {
		var validationErr *ValidationError

		if errors.As(err, &validationErr) {
			httpx.Error(
				w,
				httpx.CodeValidationFailed,
				requestID,
				validationErr.Message,
			)
			return
		}

		h.logger.Error(
			"create listing failed",
			"request_id", requestID,
			"error", err,
		)

		httpx.Error(
			w,
			httpx.CodeInternalError,
			requestID,
		)
		return
	}

	h.logger.Info(
		"listing created",
		"listing_id", created.ID,
		"request_id", requestID,
	)

	httpx.Success(
		w,
		http.StatusCreated,
		toListingResponse(created),
		requestID,
	)
}

// Update handles PUT requests that replace all editable listing fields.
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := middleware.RequestIdFromContext(ctx)
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil || id == uuid.Nil {
		httpx.Error(w, httpx.CodeInvalidID, requestID)
		return
	}

	var req UpdateListingRequest
	if err := httpx.ReadJSON(w, r, &req); err != nil {
		httpx.Error(w, httpx.CodeInvalidJSON, requestID, err.Error())
		return
	}
	updated, err := h.service.Update(ctx, id, UpdateInput{
		Title: req.Title, Description: req.Description, Price: req.Price, City: req.City,
	})
	if err != nil {
		var validationErr *ValidationError
		switch {
		case errors.As(err, &validationErr):
			httpx.Error(w, httpx.CodeValidationFailed, requestID, validationErr.Message)
		case errors.Is(err, ErrNotFound):
			httpx.Error(w, httpx.CodeNotFound, requestID, "Listing not found")
		default:
			h.logger.Error("update listing failed", "listing_id", id, "request_id", requestID, "error", err)
			httpx.Error(w, httpx.CodeInternalError, requestID)
		}
		return
	}
	h.logger.Info("listing updated", "listing_id", id, "request_id", requestID)
	httpx.Success(w, http.StatusOK, toListingResponse(updated), requestID)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	requestID := middleware.RequestIdFromContext(ctx)

	listings, err := h.service.List(ctx)

	if err != nil {
		h.logger.Error(
			"fetch listings failed",
			"request_id", requestID,
			"error", err,
		)

		httpx.Error(
			w,
			httpx.CodeInternalError,
			requestID,
		)
		return
	}

	httpx.Success(
		w,
		http.StatusOK,
		toListingResponses(listings),
		requestID,
	)
}

func (h *Handler) FindByID(
	w http.ResponseWriter,
	r *http.Request,
) {
	ctx := r.Context()

	requestID := middleware.RequestIdFromContext(ctx)

	id, err := uuid.Parse(r.PathValue("id"))

	// A zero UUID is syntactically valid but is not a usable listing ID.
	if err != nil || id == uuid.Nil {
		httpx.Error(
			w,
			httpx.CodeInvalidID,
			requestID,
		)
		return
	}

	found, err := h.service.FindByID(ctx, id)

	if errors.Is(err, ErrNotFound) {
		httpx.Error(
			w,
			httpx.CodeNotFound,
			requestID,
			"Listing not found",
		)
		return
	}

	if err != nil {
		h.logger.Error(
			"fetch listing failed",
			"listing_id", id,
			"request_id", requestID,
			"error", err,
		)

		httpx.Error(
			w,
			httpx.CodeInternalError,
			requestID,
		)
		return
	}

	httpx.Success(
		w,
		http.StatusOK,
		toListingResponse(found),
		requestID,
	)
}

func (h *Handler) Delete(
	w http.ResponseWriter,
	r *http.Request,
) {
	ctx := r.Context()

	requestID := middleware.RequestIdFromContext(ctx)

	id, err := uuid.Parse(r.PathValue("id"))

	if err != nil || id == uuid.Nil {
		httpx.Error(
			w,
			httpx.CodeInvalidID,
			requestID,
		)
		return
	}

	err = h.service.Delete(ctx, id)

	if errors.Is(err, ErrNotFound) {
		httpx.Error(
			w,
			httpx.CodeNotFound,
			requestID,
			"Listing not found",
		)
		return
	}

	if err != nil {
		h.logger.Error(
			"delete listing failed",
			"listing_id", id,
			"request_id", requestID,
			"error", err,
		)

		httpx.Error(
			w,
			httpx.CodeInternalError,
			requestID,
		)
		return
	}

	h.logger.Info(
		"listing deleted",
		"listing_id", id,
		"request_id", requestID,
	)

	w.WriteHeader(http.StatusNoContent)
}

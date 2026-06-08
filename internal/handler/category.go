package handler

import (
	"context"
	"net/http"

	"github.com/Azzt17/finance-tracker/internal/middleware"
	"github.com/Azzt17/finance-tracker/internal/model"
)

type CategoryRepository interface {
	List(ctx context.Context, userID int64) ([]model.Category, error)
	Create(ctx context.Context, userID int64, input model.CategoryInput) (model.Category, error)
	Update(ctx context.Context, userID int64, id int64, input model.CategoryInput) (model.Category, error)
	Delete(ctx context.Context, userID int64, id int64) error
}

type CategoryHandler struct {
	repository CategoryRepository
}

func NewCategoryHandler(repository CategoryRepository) *CategoryHandler {
	return &CategoryHandler{repository: repository}
}

func (h *CategoryHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/categories", h.list)
	mux.HandleFunc("POST /api/v1/categories", h.create)
	mux.HandleFunc("PATCH /api/v1/categories/{id}", h.update)
	mux.HandleFunc("DELETE /api/v1/categories/{id}", h.delete)
}

func (h *CategoryHandler) list(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserFromContext(r.Context()).ID
	categories, err := h.repository.List(r.Context(), userID)
	if err != nil {
		writeRepositoryError(w, err)
		return
	}

	if categories == nil {
		categories = []model.Category{}
	}
	writeJSON(w, http.StatusOK, categories)
}

func (h *CategoryHandler) create(w http.ResponseWriter, r *http.Request) {
	var input model.CategoryInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, err.Error(), "VALIDATION_ERROR")
		return
	}

	if input.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required", "VALIDATION_ERROR")
		return
	}

	userID := middleware.UserFromContext(r.Context()).ID
	category, err := h.repository.Create(r.Context(), userID, input)
	if err != nil {
		writeRepositoryError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, category)
}

func (h *CategoryHandler) update(w http.ResponseWriter, r *http.Request) {
	var input model.CategoryInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, err.Error(), "VALIDATION_ERROR")
		return
	}

	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid category id", "VALIDATION_ERROR")
		return
	}

	userID := middleware.UserFromContext(r.Context()).ID
	category, err := h.repository.Update(r.Context(), userID, id, input)
	if err != nil {
		writeRepositoryError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, category)
}

func (h *CategoryHandler) delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid category id", "VALIDATION_ERROR")
		return
	}
	userID := middleware.UserFromContext(r.Context()).ID
	if err := h.repository.Delete(r.Context(), userID, id); err != nil {
		writeRepositoryError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "deleted successfully"})
}

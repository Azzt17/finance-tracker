package handler

import (
	"context"
	"net/http"

	"github.com/Azzt17/finance-tracker/internal/model"
)

type CategoryRepository interface {
	List(ctx context.Context) ([]model.Category, error)
	Create(ctx context.Context, input model.CategoryInput) (model.Category, error)
	Update(ctx context.Context, id int64, input model.CategoryInput) (model.Category, error)
	Delete(ctx context.Context, id int64) error
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
	categories, err := h.repository.List(r.Context())
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

	category, err := h.repository.Create(r.Context(), input)
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

	category, err := h.repository.Update(r.Context(), id, input)
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
	if err := h.repository.Delete(r.Context(), id); err != nil {
		writeRepositoryError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "deleted successfully"})
}

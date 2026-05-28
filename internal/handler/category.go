package handler

import (
	"net/http"
	"time"

	"github.com/Azzt17/finance-tracker/internal/model"
)

type CategoryHandler struct{}

func NewCategoryHandler() *CategoryHandler {
	return &CategoryHandler{}
}

func (h *CategoryHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/categories", h.list)
	mux.HandleFunc("POST /api/v1/categories", h.create)
	mux.HandleFunc("GET /api/v1/categories/{id}", h.get)
	mux.HandleFunc("PUT /api/v1/categories/{id}", h.update)
	mux.HandleFunc("DELETE /api/v1/categories/{id}", h.delete)
}

func (h *CategoryHandler) list(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, envelope{
		"data": []model.Category{},
	})
}

func (h *CategoryHandler) create(w http.ResponseWriter, r *http.Request) {
	var input model.CategoryInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	category := model.Category{
		ID:        "pending",
		Name:      input.Name,
		Type:      input.Type,
		CreatedAt: time.Now().UTC(),
	}

	writeJSON(w, http.StatusCreated, envelope{
		"data": category,
	})
}

func (h *CategoryHandler) get(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, envelope{
		"data": envelope{
			"id": r.PathValue("id"),
		},
	})
}

func (h *CategoryHandler) update(w http.ResponseWriter, r *http.Request) {
	var input model.CategoryInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, envelope{
		"data": envelope{
			"id": r.PathValue("id"),
		},
	})
}

func (h *CategoryHandler) delete(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusNoContent, nil)
}

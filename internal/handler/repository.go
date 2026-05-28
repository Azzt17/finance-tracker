package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/Azzt17/finance-tracker/internal/repository"
)

func parseID(r *http.Request) (int64, error) {
	return strconv.ParseInt(r.PathValue("id"), 10, 64)
}

func writeRepositoryError(w http.ResponseWriter, err error) {
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, "resource not found", "NOT_FOUND")
		return
	}

	writeError(w, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError), "INTERNAL_ERROR")
}

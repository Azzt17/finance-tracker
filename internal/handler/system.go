package handler

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/Azzt17/finance-tracker/internal/service"
)

type SystemHandler struct {
	Config Config
}

func (h *SystemHandler) DownloadBackup(w http.ResponseWriter, r *http.Request) {
	destPath, err := service.BackupDatabase(h.Config.DB, h.Config.DatabaseURL)
	if err != nil {
		slog.Error("Failed to backup database", "error", err)
		http.Error(w, "Failed to create backup", http.StatusInternalServerError)
		return
	}

	file, err := os.Open(destPath)
	if err != nil {
		slog.Error("Failed to open backup file", "error", err)
		http.Error(w, "Failed to read backup", http.StatusInternalServerError)
		return
	}
	defer func() { _ = file.Close() }()

	stat, err := file.Stat()
	if err != nil {
		slog.Error("Failed to stat backup file", "error", err)
		http.Error(w, "Failed to read backup stats", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename="+stat.Name())
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", stat.Size()))

	http.ServeContent(w, r, stat.Name(), stat.ModTime(), file)
}

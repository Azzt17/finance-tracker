package handler

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Azzt17/finance-tracker/internal/middleware"
	"github.com/Azzt17/finance-tracker/internal/model"
)

type ExportRepository interface {
	GetAggregation(ctx context.Context, userID int64, yearMonth string) (model.BudgetAggregation, error)
	ListTransactions(ctx context.Context, userID int64, yearMonth string) ([]model.Transaction, error)
}

type ExportHandler struct {
	repository ExportRepository
}

func NewExportHandler(repository ExportRepository) *ExportHandler {
	return &ExportHandler{repository: repository}
}

func (h *ExportHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/export/markdown", h.exportMarkdown)
}

func (h *ExportHandler) exportMarkdown(w http.ResponseWriter, r *http.Request) {
	yearMonth := r.URL.Query().Get("year_month")
	if yearMonth == "" {
		yearMonth = time.Now().Format("2006-01") // Default to current month
	}

	userID := middleware.UserFromContext(r.Context()).ID
	agg, err := h.repository.GetAggregation(r.Context(), userID, yearMonth)
	if err != nil {
		writeRepositoryError(w, err)
		return
	}

	transactions, err := h.repository.ListTransactions(r.Context(), userID, yearMonth)
	if err != nil {
		writeRepositoryError(w, err)
		return
	}

	var buf bytes.Buffer
	fmt.Fprintf(&buf, "# Laporan Keuangan: %s\n\n", yearMonth)
	buf.WriteString("## Ringkasan Anggaran\n")
	fmt.Fprintf(&buf, "- **Total Anggaran:** Rp %d\n", agg.TotalBudget)
	fmt.Fprintf(&buf, "- **Total Pengeluaran:** Rp %d\n", agg.TotalSpent)
	fmt.Fprintf(&buf, "- **Sisa Saldo:** Rp %d\n\n", agg.RemainingBalance)

	buf.WriteString("## Pengeluaran Berdasarkan Kategori\n")
	for _, cat := range agg.SpendingByCategory {
		fmt.Fprintf(&buf, "- **%s**: Rp %d\n", cat.CategoryName, cat.Total)
	}
	buf.WriteString("\n")

	buf.WriteString("## Daftar Transaksi\n")
	for _, t := range transactions {
		dateStr := t.TransactedAt.Format("02 Jan 2006")
		note := t.Note
		if note == "" {
			note = "Tidak ada catatan"
		}
		fmt.Fprintf(&buf, "- %s: Rp %d (%s)\n", dateStr, t.Amount, note)
	}

	w.Header().Set("Content-Type", "text/markdown")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"export-%s.md\"", yearMonth))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(buf.Bytes())
}

package database

import (
	"context"
	"database/sql"
	"log/slog"
)

func Seed(ctx context.Context, db *sql.DB) error {
	categories := []struct {
		Name       string
		IconEmoji  string
		IsQuickAdd int
		SortOrder  int
	}{
		{"Makan", "🍛", 1, 1},
		{"Transportasi", "🚌", 1, 2},
		{"Entertainment", "🎮", 1, 3},
	}

	for _, c := range categories {
		var exists bool
		err := db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM categories WHERE name = ?)", c.Name).Scan(&exists)
		if err != nil {
			return err
		}
		if !exists {
			_, err = db.ExecContext(ctx, "INSERT INTO categories (name, icon_emoji, is_quick_add, sort_order) VALUES (?, ?, ?, ?)", c.Name, c.IconEmoji, c.IsQuickAdd, c.SortOrder)
			if err != nil {
				return err
			}
			slog.Info("Seeded category", "name", c.Name)
		}
	}
	return nil
}

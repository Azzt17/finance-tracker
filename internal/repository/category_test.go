package repository_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/Azzt17/finance-tracker/internal/database"
	"github.com/Azzt17/finance-tracker/internal/model"
	"github.com/Azzt17/finance-tracker/internal/repository"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := database.Open(context.Background(), "file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	if err := database.Migrate(context.Background(), db); err != nil {
		t.Fatalf("failed to migrate test db: %v", err)
	}
	return db
}

func TestCategoryRepository_CreateAndList(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := repository.NewCategoryRepository(db)
	ctx := context.Background()

	input := model.CategoryInput{
		Name:       "Test Category",
		IconEmoji:  "🚀",
		IsQuickAdd: true,
		SortOrder:  1,
	}

	created, err := repo.Create(ctx, input)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if created.Name != input.Name {
		t.Errorf("expected name %s, got %s", input.Name, created.Name)
	}

	categories, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(categories) != 1 {
		t.Errorf("expected 1 category, got %d", len(categories))
	}
}

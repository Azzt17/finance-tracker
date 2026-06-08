package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Azzt17/finance-tracker/internal/handler"
	"github.com/Azzt17/finance-tracker/internal/middleware"
	"github.com/Azzt17/finance-tracker/internal/model"
)

type mockCategoryRepo struct {
	categories []model.Category
}

func (m *mockCategoryRepo) List(ctx context.Context, userID int64) ([]model.Category, error) {
	return m.categories, nil
}

func (m *mockCategoryRepo) Create(ctx context.Context, userID int64, input model.CategoryInput) (model.Category, error) {
	c := model.Category{
		ID:         1,
		Name:       input.Name,
		IconEmoji:  input.IconEmoji,
		IsQuickAdd: input.IsQuickAdd,
		SortOrder:  input.SortOrder,
		CreatedAt:  time.Now(),
	}
	m.categories = append(m.categories, c)
	return c, nil
}

func (m *mockCategoryRepo) Update(ctx context.Context, userID int64, id int64, input model.CategoryInput) (model.Category, error) {
	if len(m.categories) > 0 {
		m.categories[0].Name = input.Name
		return m.categories[0], nil
	}
	return model.Category{}, nil
}

func (m *mockCategoryRepo) Delete(ctx context.Context, userID int64, id int64) error {
	m.categories = []model.Category{}
	return nil
}

func TestCategoryHandler_ListCategories(t *testing.T) {
	repo := &mockCategoryRepo{
		categories: []model.Category{
			{ID: 1, Name: "Makan", IconEmoji: "🍔"},
		},
	}
	h := handler.NewCategoryHandler(repo)

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/categories", nil)
	req = req.WithContext(middleware.ContextWithUser(req.Context(), &model.User{ID: 1}))
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

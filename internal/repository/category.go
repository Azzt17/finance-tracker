package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Azzt17/finance-tracker/internal/model"
)

type CategoryRepository struct {
	db *sql.DB
}

func NewCategoryRepository(db *sql.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) List(ctx context.Context) (categories []model.Category, err error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, icon_emoji, is_quick_add, sort_order, created_at
		FROM categories
		ORDER BY sort_order ASC, name ASC, id ASC
	`)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	categories = []model.Category{}
	for rows.Next() {
		category, err := scanCategory(rows)
		if err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}

func (r *CategoryRepository) Create(ctx context.Context, input model.CategoryInput) (model.Category, error) {
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO categories (name, icon_emoji, is_quick_add, sort_order)
		VALUES (?, ?, ?, ?)
	`, input.Name, nullableString(input.IconEmoji), boolToInt(input.IsQuickAdd), input.SortOrder)
	if err != nil {
		return model.Category{}, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return model.Category{}, err
	}

	return r.Get(ctx, id)
}

func (r *CategoryRepository) Get(ctx context.Context, id int64) (model.Category, error) {
	category, err := scanCategory(r.db.QueryRowContext(ctx, `
		SELECT id, name, icon_emoji, is_quick_add, sort_order, created_at
		FROM categories
		WHERE id = ?
	`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return model.Category{}, ErrNotFound
	}
	if err != nil {
		return model.Category{}, err
	}

	return category, nil
}

func (r *CategoryRepository) Update(ctx context.Context, id int64, input model.CategoryInput) (model.Category, error) {
	result, err := r.db.ExecContext(ctx, `
		UPDATE categories
		SET name = ?, icon_emoji = ?, is_quick_add = ?, sort_order = ?
		WHERE id = ?
	`, input.Name, nullableString(input.IconEmoji), boolToInt(input.IsQuickAdd), input.SortOrder, id)
	if err != nil {
		return model.Category{}, err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return model.Category{}, err
	}
	if affected == 0 {
		return model.Category{}, ErrNotFound
	}

	return r.Get(ctx, id)
}

func (r *CategoryRepository) Delete(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM categories WHERE id = ?`, id)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrNotFound
	}

	return nil
}

type categoryScanner interface {
	Scan(dest ...any) error
}

func scanCategory(scanner categoryScanner) (model.Category, error) {
	var (
		category   model.Category
		iconEmoji  sql.NullString
		isQuickAdd int
		createdAt  string
	)
	if err := scanner.Scan(
		&category.ID,
		&category.Name,
		&iconEmoji,
		&isQuickAdd,
		&category.SortOrder,
		&createdAt,
	); err != nil {
		return model.Category{}, err
	}

	parsedCreatedAt, err := parseDBTime(createdAt)
	if err != nil {
		return model.Category{}, err
	}

	category.IconEmoji = stringFromNull(iconEmoji)
	category.IsQuickAdd = intToBool(isQuickAdd)
	category.CreatedAt = parsedCreatedAt

	return category, nil
}

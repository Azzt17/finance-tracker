package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Azzt17/finance-tracker/internal/model"
)

type SavingsGoalRepository struct {
	db *sql.DB
}

func NewSavingsGoalRepository(db *sql.DB) *SavingsGoalRepository {
	return &SavingsGoalRepository{db: db}
}

func (r *SavingsGoalRepository) List(ctx context.Context) (goals []model.SavingsGoal, err error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, target_amount, current_saved, year_month, is_achieved, created_at
		FROM savings_goals
		ORDER BY year_month DESC, id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	goals = []model.SavingsGoal{}
	for rows.Next() {
		goal, err := scanSavingsGoal(rows)
		if err != nil {
			return nil, err
		}
		goals = append(goals, goal)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return goals, nil
}

func (r *SavingsGoalRepository) Create(ctx context.Context, input model.SavingsGoalInput) (model.SavingsGoal, error) {
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO savings_goals (
			name,
			target_amount,
			current_saved,
			year_month,
			is_achieved
		)
		VALUES (?, ?, ?, ?, ?)
	`, input.Name, input.TargetAmount, input.CurrentSaved, input.YearMonth, boolToInt(input.IsAchieved))
	if err != nil {
		return model.SavingsGoal{}, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return model.SavingsGoal{}, err
	}

	return r.Get(ctx, id)
}

func (r *SavingsGoalRepository) Get(ctx context.Context, id int64) (model.SavingsGoal, error) {
	goal, err := scanSavingsGoal(r.db.QueryRowContext(ctx, `
		SELECT id, name, target_amount, current_saved, year_month, is_achieved, created_at
		FROM savings_goals
		WHERE id = ?
	`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return model.SavingsGoal{}, ErrNotFound
	}
	if err != nil {
		return model.SavingsGoal{}, err
	}

	return goal, nil
}

func (r *SavingsGoalRepository) Update(ctx context.Context, id int64, input model.SavingsGoalInput) (model.SavingsGoal, error) {
	result, err := r.db.ExecContext(ctx, `
		UPDATE savings_goals
		SET name = ?,
			target_amount = ?,
			current_saved = ?,
			year_month = ?,
			is_achieved = ?
		WHERE id = ?
	`, input.Name, input.TargetAmount, input.CurrentSaved, input.YearMonth, boolToInt(input.IsAchieved), id)
	if err != nil {
		return model.SavingsGoal{}, err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return model.SavingsGoal{}, err
	}
	if affected == 0 {
		return model.SavingsGoal{}, ErrNotFound
	}

	return r.Get(ctx, id)
}

func (r *SavingsGoalRepository) Delete(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM savings_goals WHERE id = ?`, id)
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

type savingsGoalScanner interface {
	Scan(dest ...any) error
}

func scanSavingsGoal(scanner savingsGoalScanner) (model.SavingsGoal, error) {
	var (
		goal       model.SavingsGoal
		isAchieved int
		createdAt  string
	)
	if err := scanner.Scan(
		&goal.ID,
		&goal.Name,
		&goal.TargetAmount,
		&goal.CurrentSaved,
		&goal.YearMonth,
		&isAchieved,
		&createdAt,
	); err != nil {
		return model.SavingsGoal{}, err
	}

	parsedCreatedAt, err := parseDBTime(createdAt)
	if err != nil {
		return model.SavingsGoal{}, err
	}

	goal.IsAchieved = intToBool(isAchieved)
	goal.CreatedAt = parsedCreatedAt

	return goal, nil
}

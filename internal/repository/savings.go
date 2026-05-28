package repository

import (
	"context"
	"database/sql"

	"github.com/Azzt17/finance-tracker/internal/model"
)

type SavingsRepository struct {
	db *sql.DB
}

func NewSavingsRepository(db *sql.DB) *SavingsRepository {
	return &SavingsRepository{db: db}
}

func (r *SavingsRepository) List(ctx context.Context, yearMonth string) ([]model.SavingsGoal, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, target_amount, current_saved, year_month, is_achieved, created_at
		FROM savings_goals
		WHERE year_month = ?
		ORDER BY id DESC
	`, yearMonth)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var goals []model.SavingsGoal
	for rows.Next() {
		var g model.SavingsGoal
		var isAchieved int
		var createdAt string
		if err := rows.Scan(&g.ID, &g.Name, &g.TargetAmount, &g.CurrentSaved, &g.YearMonth, &isAchieved, &createdAt); err != nil {
			return nil, err
		}
		g.IsAchieved = isAchieved != 0
		g.CreatedAt, _ = parseDBTime(createdAt)
		goals = append(goals, g)
	}
	if goals == nil {
		goals = []model.SavingsGoal{}
	}
	return goals, nil
}

func (r *SavingsRepository) Create(ctx context.Context, input model.SavingsGoalInput) (model.SavingsGoal, error) {
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO savings_goals (name, target_amount, current_saved, year_month, is_achieved)
		VALUES (?, ?, ?, ?, ?)
	`, input.Name, input.TargetAmount, input.CurrentSaved, input.YearMonth, boolToInt(input.IsAchieved))
	if err != nil {
		return model.SavingsGoal{}, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return model.SavingsGoal{}, err
	}

	return model.SavingsGoal{
		ID:           id,
		Name:         input.Name,
		TargetAmount: input.TargetAmount,
		CurrentSaved: input.CurrentSaved,
		YearMonth:    input.YearMonth,
		IsAchieved:   input.IsAchieved,
	}, nil
}

func (r *SavingsRepository) Update(ctx context.Context, id int64, input model.SavingsGoalInput) (model.SavingsGoal, error) {
	_, err := r.db.ExecContext(ctx, `
		UPDATE savings_goals
		SET name = ?, target_amount = ?, current_saved = ?, year_month = ?, is_achieved = ?
		WHERE id = ?
	`, input.Name, input.TargetAmount, input.CurrentSaved, input.YearMonth, boolToInt(input.IsAchieved), id)
	if err != nil {
		return model.SavingsGoal{}, err
	}

	return model.SavingsGoal{
		ID:           id,
		Name:         input.Name,
		TargetAmount: input.TargetAmount,
		CurrentSaved: input.CurrentSaved,
		YearMonth:    input.YearMonth,
		IsAchieved:   input.IsAchieved,
	}, nil
}

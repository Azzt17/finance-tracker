package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Azzt17/finance-tracker/internal/model"
)

type AnalyticsRepository struct {
	db *sql.DB
}

func NewAnalyticsRepository(db *sql.DB) *AnalyticsRepository {
	return &AnalyticsRepository{db: db}
}

func (r *AnalyticsRepository) GetSpendingByCategory(ctx context.Context, yearMonth string) (model.AnalyticsSpendingByCategory, error) {
	query := `
		SELECT 
			COALESCE(t.category_id, 0) as category_id,
			COALESCE(c.name, 'Uncategorized') as category_name,
			COALESCE(c.icon_emoji, '') as icon_emoji,
			SUM(t.amount) as total
		FROM transactions t
		LEFT JOIN categories c ON t.category_id = c.id
		WHERE strftime('%Y-%m', t.transacted_at) = ? AND t.amount > 0
		GROUP BY COALESCE(t.category_id, 0)
		ORDER BY total DESC
	`
	rows, err := r.db.QueryContext(ctx, query, yearMonth)
	if err != nil {
		return model.AnalyticsSpendingByCategory{}, err
	}
	defer func() { _ = rows.Close() }()

	var totalSpent int64
	var breakdown []model.AnalyticsCategorySpending

	for rows.Next() {
		var item model.AnalyticsCategorySpending
		if err := rows.Scan(&item.CategoryID, &item.CategoryName, &item.IconEmoji, &item.Total); err != nil {
			return model.AnalyticsSpendingByCategory{}, err
		}
		totalSpent += item.Total
		breakdown = append(breakdown, item)
	}
	if err := rows.Err(); err != nil {
		return model.AnalyticsSpendingByCategory{}, err
	}

	for i := range breakdown {
		if totalSpent > 0 {
			breakdown[i].Percentage = float64(breakdown[i].Total) / float64(totalSpent) * 100.0
		}
	}

	if breakdown == nil {
		breakdown = []model.AnalyticsCategorySpending{}
	}

	return model.AnalyticsSpendingByCategory{
		YearMonth: yearMonth,
		Total:     totalSpent,
		Breakdown: breakdown,
	}, nil
}

func (r *AnalyticsRepository) GetMonthlyTrend(ctx context.Context, months int) (model.AnalyticsMonthlyTrend, error) {
	if months <= 0 {
		months = 6
	}
	if months > 12 {
		months = 12
	}

	query := `
		WITH RECURSIVE months_cte AS (
			SELECT strftime('%Y-%m', 'now', 'localtime', '-' || (? - 1) || ' months') AS ym
			UNION ALL
			SELECT strftime('%Y-%m', ym || '-01', '+1 month')
			FROM months_cte
			WHERE ym < strftime('%Y-%m', 'now', 'localtime')
		)
		SELECT 
			m.ym as year_month,
			COALESCE(SUM(t.amount), 0) as total_spent,
			COALESCE(b.total_budget, 0) as total_budget
		FROM months_cte m
		LEFT JOIN transactions t ON strftime('%Y-%m', t.transacted_at) = m.ym AND t.amount > 0
		LEFT JOIN budget_allocation b ON b.year_month = m.ym
		GROUP BY m.ym
		ORDER BY m.ym ASC
	`
	rows, err := r.db.QueryContext(ctx, query, months)
	if err != nil {
		return model.AnalyticsMonthlyTrend{}, err
	}
	defer func() { _ = rows.Close() }()

	var data []model.AnalyticsMonthlyTrendData
	for rows.Next() {
		var item model.AnalyticsMonthlyTrendData
		if err := rows.Scan(&item.YearMonth, &item.TotalSpent, &item.TotalBudget); err != nil {
			return model.AnalyticsMonthlyTrend{}, err
		}
		data = append(data, item)
	}
	if err := rows.Err(); err != nil {
		return model.AnalyticsMonthlyTrend{}, err
	}

	if data == nil {
		data = []model.AnalyticsMonthlyTrendData{}
	}

	return model.AnalyticsMonthlyTrend{
		Months: months,
		Data:   data,
	}, nil
}

func (r *AnalyticsRepository) GetDailySpending(ctx context.Context, yearMonth string) (model.AnalyticsDailySpending, error) {
	_, err := time.Parse("2006-01", yearMonth)
	if err != nil {
		return model.AnalyticsDailySpending{}, fmt.Errorf("invalid year_month format: %v", err)
	}

	query := `
		WITH RECURSIVE dates_cte AS (
			SELECT ? || '-01' AS d
			UNION ALL
			SELECT strftime('%Y-%m-%d', d, '+1 day')
			FROM dates_cte
			WHERE strftime('%m', d, '+1 day') = strftime('%m', d)
		)
		SELECT 
			d.d as date,
			COALESCE(SUM(t.amount), 0) as total
		FROM dates_cte d
		LEFT JOIN transactions t ON strftime('%Y-%m-%d', t.transacted_at) = d.d AND t.amount > 0
		GROUP BY d.d
		ORDER BY d.d ASC
	`
	rows, err := r.db.QueryContext(ctx, query, yearMonth)
	if err != nil {
		return model.AnalyticsDailySpending{}, err
	}
	defer func() { _ = rows.Close() }()

	var data []model.AnalyticsDailySpendingData
	for rows.Next() {
		var item model.AnalyticsDailySpendingData
		if err := rows.Scan(&item.Date, &item.Total); err != nil {
			return model.AnalyticsDailySpending{}, err
		}
		data = append(data, item)
	}
	if err := rows.Err(); err != nil {
		return model.AnalyticsDailySpending{}, err
	}

	if data == nil {
		data = []model.AnalyticsDailySpendingData{}
	}

	return model.AnalyticsDailySpending{
		YearMonth: yearMonth,
		Data:      data,
	}, nil
}

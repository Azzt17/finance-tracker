package model

type AnalyticsCategorySpending struct {
	CategoryID   int64   `json:"category_id"`
	CategoryName string  `json:"category_name"`
	IconEmoji    string  `json:"icon_emoji"`
	Total        int64   `json:"total"`
	Percentage   float64 `json:"percentage"`
}

type AnalyticsSpendingByCategory struct {
	YearMonth string                      `json:"year_month"`
	Total     int64                       `json:"total"`
	Breakdown []AnalyticsCategorySpending `json:"breakdown"`
}

type AnalyticsMonthlyTrendData struct {
	YearMonth   string `json:"year_month"`
	TotalSpent  int64  `json:"total_spent"`
	TotalBudget int64  `json:"total_budget"`
}

type AnalyticsMonthlyTrend struct {
	Months int                         `json:"months"`
	Data   []AnalyticsMonthlyTrendData `json:"data"`
}

type AnalyticsDailySpendingData struct {
	Date  string `json:"date"`
	Total int64  `json:"total"`
}

type AnalyticsDailySpending struct {
	YearMonth string                       `json:"year_month"`
	Data      []AnalyticsDailySpendingData `json:"data"`
}

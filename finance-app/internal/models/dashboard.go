package models

import "time"

type DashboardPeriod struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// Scalar amounts are omitted when the corresponding totals span currencies.
type DashboardSummary struct {
	TotalBalance   *int64 `json:"totalBalance,omitempty"`
	Currency       string `json:"currency,omitempty"`
	PeriodIncome   *int64 `json:"periodIncome,omitempty"`
	PeriodExpenses *int64 `json:"periodExpenses,omitempty"`
	NetSavings     *int64 `json:"netSavings,omitempty"`
}

type AccountBalance struct {
	AccountID   int64  `json:"accountId"`
	AccountName string `json:"accountName"`
	Currency    string `json:"currency"`
	Balance     int64  `json:"balance"`
}

type CurrencyBalance struct {
	Currency string `json:"currency"`
	Amount   int64  `json:"amount"`
}

type CurrencySummary struct {
	Currency       string `json:"currency"`
	PeriodIncome   int64  `json:"periodIncome"`
	PeriodExpenses int64  `json:"periodExpenses"`
	NetSavings     int64  `json:"netSavings"`
}

type CategoryTotal struct {
	CategoryID   int64  `json:"categoryId"`
	CategoryName string `json:"categoryName"`
	Currency     string `json:"currency"`
	Amount       int64  `json:"amount"`
}

type DashboardMovement struct {
	ID          int64     `json:"id"`
	Kind        string    `json:"kind"`
	Amount      int64     `json:"amount"`
	Description string    `json:"description"`
	Date        time.Time `json:"date"`
	AccountID   int64     `json:"accountId"`
	ToAccountID int64     `json:"toAccountId,omitempty"`
}

type DashboardResponse struct {
	Period             DashboardPeriod     `json:"period"`
	Summary            DashboardSummary    `json:"summary"`
	BalancesByCurrency []CurrencyBalance   `json:"balancesByCurrency"`
	SummaryByCurrency  []CurrencySummary   `json:"summaryByCurrency"`
	Accounts           []AccountBalance    `json:"accounts"`
	ExpensesByCategory []CategoryTotal     `json:"expensesByCategory"`
	IncomeByCategory   []CategoryTotal     `json:"incomeByCategory"`
	RecentMovements    []DashboardMovement `json:"recentMovements"`
}

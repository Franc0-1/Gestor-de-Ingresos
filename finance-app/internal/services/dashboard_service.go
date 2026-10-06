package services

import (
	"cmp"
	"errors"
	"fmt"
	"math"
	"slices"
	"time"

	"finance-app/internal/models"
	"finance-app/internal/repository"
)

var ErrMoneyOverflow = errors.New("financial total exceeds int64 range")

type DashboardService struct {
	accounts     repository.AccountRepository
	categories   repository.CategoryRepository
	transactions repository.TransactionRepository
	transfers    repository.TransferRepository
}

func NewDashboardService(accounts repository.AccountRepository, categories repository.CategoryRepository, transactions repository.TransactionRepository, transfers repository.TransferRepository) *DashboardService {
	return &DashboardService{accounts, categories, transactions, transfers}
}

func dashboardPeriod(from, to time.Time) (time.Time, time.Time, error) {
	if from.IsZero() && to.IsZero() {
		now := time.Now().UTC()
		from = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	}
	if from.IsZero() {
		from = time.Date(to.Year(), to.Month(), 1, 0, 0, 0, 0, time.UTC)
	}
	if to.IsZero() {
		start := time.Date(from.Year(), from.Month(), 1, 0, 0, 0, 0, time.UTC)
		to = start.AddDate(0, 1, 0).Add(-time.Nanosecond)
	}
	if err := validateDashboardRange(from, to); err != nil {
		return from, to, err
	}
	return from, to, nil
}

func validateDashboardRange(from, to time.Time) error {
	if from.After(to) {
		return &models.ValidationError{Field: "from", Reason: "must not be after to"}
	}
	return nil
}

func addMoney(a, b int64) (int64, error) {
	if (b > 0 && a > math.MaxInt64-b) || (b < 0 && a < math.MinInt64-b) {
		return 0, ErrMoneyOverflow
	}
	return a + b, nil
}

func (s *DashboardService) GetDashboard(from, to time.Time) (models.DashboardResponse, error) {
	from, to, err := dashboardPeriod(from, to)
	if err != nil {
		return models.DashboardResponse{}, err
	}
	accounts, err := s.accounts.GetAll()
	if err != nil {
		return models.DashboardResponse{}, fmt.Errorf("load dashboard accounts: %w", err)
	}
	categories, err := s.categories.GetAll()
	if err != nil {
		return models.DashboardResponse{}, fmt.Errorf("load dashboard categories: %w", err)
	}
	txs, err := s.transactions.GetAll()
	if err != nil {
		return models.DashboardResponse{}, fmt.Errorf("load dashboard transactions: %w", err)
	}
	transfers, err := s.transfers.GetAll()
	if err != nil {
		return models.DashboardResponse{}, fmt.Errorf("load dashboard transfers: %w", err)
	}
	response := models.DashboardResponse{Period: models.DashboardPeriod{From: from.Format(time.DateOnly), To: to.Format(time.DateOnly)}}
	response.Accounts, response.BalancesByCurrency, err = calculateAccountBalances(accounts, txs, transfers)
	if err != nil {
		return models.DashboardResponse{}, err
	}
	response.SummaryByCurrency, response.IncomeByCategory, response.ExpensesByCategory, err = calculatePeriodTotals(accounts, categories, txs, from, to)
	if err != nil {
		return models.DashboardResponse{}, err
	}
	response.RecentMovements = recentMovements(txs, transfers, from, to)
	zero := int64(0)
	if len(response.BalancesByCurrency) == 0 {
		response.Summary.TotalBalance = &zero
	}
	if len(response.BalancesByCurrency) == 1 {
		response.Summary.TotalBalance = &response.BalancesByCurrency[0].Amount
	}
	if len(response.SummaryByCurrency) == 0 {
		response.Summary.PeriodIncome = &zero
		response.Summary.PeriodExpenses = &zero
		response.Summary.NetSavings = &zero
	} else if len(response.SummaryByCurrency) == 1 {
		total := response.SummaryByCurrency[0]
		response.Summary.Currency = total.Currency
		response.Summary.PeriodIncome = &total.PeriodIncome
		response.Summary.PeriodExpenses = &total.PeriodExpenses
		response.Summary.NetSavings = &total.NetSavings
	}
	return response, nil
}

func calculateAccountBalances(accounts []models.Account, txs []models.Transaction, transfers []models.Transfer) ([]models.AccountBalance, []models.CurrencyBalance, error) {
	balances := make(map[int64]models.AccountBalance)
	for _, a := range accounts {
		if a.Active {
			balances[a.ID] = models.AccountBalance{AccountID: a.ID, AccountName: a.Name, Currency: a.Currency, Balance: a.InitialBalance}
		}
	}
	apply := func(id, amount int64) error {
		balance, ok := balances[id]
		if !ok {
			return nil
		}
		var err error
		balance.Balance, err = addMoney(balance.Balance, amount)
		balances[id] = balance
		return err
	}
	for _, tx := range txs {
		amount := tx.Amount
		if tx.Type == "expense" {
			amount = -amount
		}
		if err := apply(tx.AccountID, amount); err != nil {
			return nil, nil, err
		}
	}
	for _, tr := range transfers {
		if err := apply(tr.FromAccountID, -tr.Amount); err != nil {
			return nil, nil, err
		}
		if err := apply(tr.ToAccountID, tr.Amount); err != nil {
			return nil, nil, err
		}
	}
	result := make([]models.AccountBalance, 0, len(balances))
	currencies := make(map[string]int64)
	for _, balance := range balances {
		result = append(result, balance)
		amount, err := addMoney(currencies[balance.Currency], balance.Balance)
		if err != nil {
			return nil, nil, err
		}
		currencies[balance.Currency] = amount
	}
	slices.SortFunc(result, func(a, b models.AccountBalance) int {
		if c := cmp.Compare(b.Balance, a.Balance); c != 0 {
			return c
		}
		return cmp.Compare(a.AccountID, b.AccountID)
	})
	byCurrency := make([]models.CurrencyBalance, 0, len(currencies))
	for currency, amount := range currencies {
		byCurrency = append(byCurrency, models.CurrencyBalance{Currency: currency, Amount: amount})
	}
	slices.SortFunc(byCurrency, func(a, b models.CurrencyBalance) int { return cmp.Compare(a.Currency, b.Currency) })
	return result, byCurrency, nil
}

type categoryCurrency struct {
	id       int64
	currency string
}

func calculatePeriodTotals(accounts []models.Account, categories []models.Category, txs []models.Transaction, from, to time.Time) ([]models.CurrencySummary, []models.CategoryTotal, []models.CategoryTotal, error) {
	accountCurrencies := make(map[int64]string)
	stats := make(map[string]models.CurrencySummary)
	for _, a := range accounts {
		accountCurrencies[a.ID] = a.Currency
		if a.Active {
			stats[a.Currency] = models.CurrencySummary{Currency: a.Currency}
		}
	}
	names := make(map[int64]string)
	for _, c := range categories {
		names[c.ID] = c.Name
	}
	income, expenses := make(map[categoryCurrency]int64), make(map[categoryCurrency]int64)
	for _, tx := range txs {
		if tx.Date.Before(from) || tx.Date.After(to) {
			continue
		}
		currency := accountCurrencies[tx.AccountID]
		stat := stats[currency]
		stat.Currency = currency
		group := income
		var err error
		if tx.Type == "income" {
			stat.PeriodIncome, err = addMoney(stat.PeriodIncome, tx.Amount)
		} else {
			stat.PeriodExpenses, err = addMoney(stat.PeriodExpenses, tx.Amount)
			group = expenses
		}
		if err != nil {
			return nil, nil, nil, err
		}
		key := categoryCurrency{tx.CategoryID, currency}
		group[key], err = addMoney(group[key], tx.Amount)
		if err != nil {
			return nil, nil, nil, err
		}
		stats[currency] = stat
	}
	result := make([]models.CurrencySummary, 0, len(stats))
	for _, stat := range stats {
		stat.NetSavings = stat.PeriodIncome - stat.PeriodExpenses
		result = append(result, stat)
	}
	slices.SortFunc(result, func(a, b models.CurrencySummary) int { return cmp.Compare(a.Currency, b.Currency) })
	return result, categoryTotals(income, names), categoryTotals(expenses, names), nil
}

func categoryTotals(groups map[categoryCurrency]int64, names map[int64]string) []models.CategoryTotal {
	result := make([]models.CategoryTotal, 0, len(groups))
	for key, amount := range groups {
		result = append(result, models.CategoryTotal{CategoryID: key.id, CategoryName: names[key.id], Currency: key.currency, Amount: amount})
	}
	slices.SortFunc(result, func(a, b models.CategoryTotal) int {
		if c := cmp.Compare(b.Amount, a.Amount); c != 0 {
			return c
		}
		if c := cmp.Compare(a.CategoryID, b.CategoryID); c != 0 {
			return c
		}
		return cmp.Compare(a.Currency, b.Currency)
	})
	return result
}

func recentMovements(txs []models.Transaction, transfers []models.Transfer, from, to time.Time) []models.DashboardMovement {
	result := make([]models.DashboardMovement, 0)
	for _, tx := range txs {
		if !tx.Date.Before(from) && !tx.Date.After(to) {
			result = append(result, models.DashboardMovement{ID: tx.ID, Kind: tx.Type, Amount: tx.Amount, Description: tx.Description, Date: tx.Date, AccountID: tx.AccountID})
		}
	}
	for _, tr := range transfers {
		if !tr.Date.Before(from) && !tr.Date.After(to) {
			result = append(result, models.DashboardMovement{ID: tr.ID, Kind: "transfer", Amount: tr.Amount, Description: tr.Description, Date: tr.Date, AccountID: tr.FromAccountID, ToAccountID: tr.ToAccountID})
		}
	}
	slices.SortFunc(result, func(a, b models.DashboardMovement) int {
		if c := b.Date.Compare(a.Date); c != 0 {
			return c
		}
		if c := cmp.Compare(a.Kind, b.Kind); c != 0 {
			return c
		}
		return cmp.Compare(b.ID, a.ID)
	})
	return result[:min(10, len(result))]
}

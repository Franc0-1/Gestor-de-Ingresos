package services

import (
	"errors"
	"math"
	"testing"
	"time"

	"finance-app/internal/models"
)

func dashboardFixture() (*DashboardService, *accountFake, *categoryFake, *transactionFake, *transferFake) {
	a := newAccountFake()
	c := newCategoryFake()
	tx := &transactionFake{records: make(map[int64]models.Transaction), failures: make(map[string]error)}
	tr := &transferFake{records: make(map[int64]models.Transfer), failures: make(map[string]error)}
	return NewDashboardService(a, c, tx, tr), a, c, tx, tr
}

func dashboardDay(day int) time.Time { return time.Date(2026, 10, day, 0, 0, 0, 0, time.UTC) }

func TestDashboardAccountBalances(t *testing.T) {
	for _, tc := range []struct {
		name                                                 string
		income, expense, sent, received, want1, want2, total int64
	}{
		{"initial only", 0, 0, 0, 0, 100, 200, 300},
		{"income", 20, 0, 0, 0, 120, 200, 320},
		{"expense", 0, 10, 0, 0, 90, 200, 290},
		{"sent", 0, 0, 15, 0, 85, 215, 300},
		{"received", 0, 0, 0, 30, 130, 170, 300},
		{"combined", 20, 10, 15, 30, 125, 185, 310},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s, a, _, tx, tr := dashboardFixture()
			a.accounts[1] = models.Account{ID: 1, Name: "Cash", Currency: "ARS", InitialBalance: 100, Active: true}
			a.accounts[2] = models.Account{ID: 2, Name: "Bank", Currency: "ARS", InitialBalance: 200, Active: true}
			if tc.income > 0 {
				tx.records[1] = models.Transaction{ID: 1, Type: "income", Amount: tc.income, AccountID: 1, Date: dashboardDay(1)}
			}
			if tc.expense > 0 {
				tx.records[2] = models.Transaction{ID: 2, Type: "expense", Amount: tc.expense, AccountID: 1, Date: dashboardDay(1)}
			}
			if tc.sent > 0 {
				tr.records[1] = models.Transfer{ID: 1, FromAccountID: 1, ToAccountID: 2, Amount: tc.sent, Date: dashboardDay(1)}
			}
			if tc.received > 0 {
				tr.records[2] = models.Transfer{ID: 2, FromAccountID: 2, ToAccountID: 1, Amount: tc.received, Date: dashboardDay(1)}
			}
			result, err := s.GetDashboard(dashboardDay(1), dashboardDay(32).Add(-time.Nanosecond))
			if err != nil {
				t.Fatal(err)
			}
			balances := make(map[int64]int64)
			for _, b := range result.Accounts {
				balances[b.AccountID] = b.Balance
			}
			if balances[1] != tc.want1 || balances[2] != tc.want2 || result.Summary.TotalBalance == nil || *result.Summary.TotalBalance != tc.total {
				t.Fatalf("incorrect balances: %+v", result)
			}
			if *result.Summary.PeriodIncome != tc.income || *result.Summary.PeriodExpenses != tc.expense {
				t.Fatal("transfers counted as income/expense")
			}
			if a.accounts[1].InitialBalance != 100 || a.accounts[2].InitialBalance != 200 || tx.writes != 0 || tr.writes != 0 {
				t.Fatal("derived values were persisted")
			}
		})
	}
}

func TestDashboardPeriodAndCategories(t *testing.T) {
	s, a, c, tx, tr := dashboardFixture()
	a.accounts[1] = models.Account{ID: 1, Name: "Cash", Currency: "ARS", InitialBalance: 1000, Active: true}
	a.accounts[2] = models.Account{ID: 2, Name: "Wallet", Currency: "ARS", InitialBalance: 200, Active: true}
	a.accounts[3] = models.Account{ID: 3, Name: "Inactive", Currency: "ARS", InitialBalance: 500, Active: false}
	c.records[1] = models.Category{ID: 1, Name: "Salary", Type: "income"}
	c.records[2] = models.Category{ID: 2, Name: "Food", Type: "expense"}
	c.records[3] = models.Category{ID: 3, Name: "Transport", Type: "expense"}
	end := dashboardDay(32).Add(-time.Nanosecond)
	tx.records[1] = models.Transaction{ID: 1, Type: "income", Amount: 100, AccountID: 1, CategoryID: 1, Date: dashboardDay(1).Add(-time.Nanosecond)}
	tx.records[2] = models.Transaction{ID: 2, Type: "income", Amount: 300, AccountID: 1, CategoryID: 1, Date: dashboardDay(1)}
	tx.records[3] = models.Transaction{ID: 3, Type: "income", Amount: 50, AccountID: 2, CategoryID: 1, Date: dashboardDay(2)}
	tx.records[4] = models.Transaction{ID: 4, Type: "expense", Amount: 120, AccountID: 1, CategoryID: 2, Date: end}
	tx.records[5] = models.Transaction{ID: 5, Type: "expense", Amount: 30, AccountID: 2, CategoryID: 3, Date: dashboardDay(3)}
	tx.records[6] = models.Transaction{ID: 6, Type: "expense", Amount: 10, AccountID: 3, CategoryID: 2, Date: dashboardDay(4)}
	tr.records[1] = models.Transfer{ID: 1, FromAccountID: 1, ToAccountID: 2, Amount: 75, Date: dashboardDay(2)}
	tr.records[2] = models.Transfer{ID: 2, FromAccountID: 2, ToAccountID: 1, Amount: 25, Date: dashboardDay(1).Add(-time.Nanosecond)}
	result, err := s.GetDashboard(dashboardDay(1), end)
	if err != nil {
		t.Fatal(err)
	}
	if *result.Summary.PeriodIncome != 350 || *result.Summary.PeriodExpenses != 160 || *result.Summary.NetSavings != 190 || *result.Summary.TotalBalance != 1500 {
		t.Fatalf("incorrect summary: %+v", result.Summary)
	}
	if len(result.Accounts) != 2 || result.Accounts[0].AccountID != 1 || result.Accounts[0].Balance != 1230 || result.Accounts[1].Balance != 270 {
		t.Fatalf("accounts: %+v", result.Accounts)
	}
	if len(result.ExpensesByCategory) != 2 || result.ExpensesByCategory[0].CategoryName != "Food" || result.ExpensesByCategory[0].Amount != 130 || result.ExpensesByCategory[1].Amount != 30 {
		t.Fatalf("expenses: %+v", result.ExpensesByCategory)
	}
	if len(result.IncomeByCategory) != 1 || result.IncomeByCategory[0].CategoryName != "Salary" || result.IncomeByCategory[0].Amount != 350 || len(result.RecentMovements) != 6 {
		t.Fatalf("period totals: %+v", result)
	}
	if result.RecentMovements[0].ID != 4 || !result.RecentMovements[0].Date.Equal(end) {
		t.Fatal("last instant of period excluded")
	}
	for _, movement := range result.RecentMovements {
		if movement.Date.Before(dashboardDay(1)) || movement.Date.After(end) {
			t.Fatal("movement outside period included")
		}
	}
	// Recalculate after a repository change, without a cache or derived writes.
	v := tx.records[2]
	v.Amount += 10
	tx.records[2] = v
	updated, err := s.GetDashboard(dashboardDay(1), end)
	if err != nil || *updated.Summary.TotalBalance != 1510 || *updated.Summary.PeriodIncome != 360 {
		t.Fatalf("dashboard did not recompute: %+v %v", updated, err)
	}
}

func TestDashboardMultipleCurrencies(t *testing.T) {
	s, a, _, tx, _ := dashboardFixture()
	a.accounts[1] = models.Account{ID: 1, Name: "ARS", Currency: "ARS", InitialBalance: 100, Active: true}
	a.accounts[2] = models.Account{ID: 2, Name: "USD", Currency: "USD", InitialBalance: 200, Active: true}
	tx.records[1] = models.Transaction{ID: 1, Type: "income", Amount: 20, AccountID: 1, CategoryID: 1, Date: dashboardDay(1)}
	tx.records[2] = models.Transaction{ID: 2, Type: "income", Amount: 50, AccountID: 2, CategoryID: 1, Date: dashboardDay(1)}
	tx.records[3] = models.Transaction{ID: 3, Type: "expense", Amount: 30, AccountID: 2, CategoryID: 2, Date: dashboardDay(1)}
	result, err := s.GetDashboard(dashboardDay(1), dashboardDay(31))
	if err != nil {
		t.Fatal(err)
	}
	if result.Summary.TotalBalance != nil || result.Summary.PeriodIncome != nil || result.Summary.NetSavings != nil {
		t.Fatal("different currencies were summed")
	}
	if len(result.BalancesByCurrency) != 2 || result.BalancesByCurrency[0].Currency != "ARS" || result.BalancesByCurrency[0].Amount != 120 || result.BalancesByCurrency[1].Amount != 220 {
		t.Fatalf("currency balances: %+v", result.BalancesByCurrency)
	}
	if len(result.SummaryByCurrency) != 2 || result.SummaryByCurrency[0].NetSavings != 20 || result.SummaryByCurrency[1].NetSavings != 20 {
		t.Fatalf("currency summary: %+v", result.SummaryByCurrency)
	}
	if len(result.IncomeByCategory) != 2 || result.IncomeByCategory[0].Currency != "USD" || result.IncomeByCategory[0].Amount != 50 || result.IncomeByCategory[1].Amount != 20 {
		t.Fatalf("categories mixed currencies: %+v", result.IncomeByCategory)
	}
}

func TestDashboardRecentMovements(t *testing.T) {
	s, a, _, tx, tr := dashboardFixture()
	a.accounts[1] = models.Account{ID: 1, Currency: "ARS", Active: true}
	a.accounts[2] = models.Account{ID: 2, Currency: "ARS", Active: true}
	for i := 1; i <= 15; i++ {
		if i%3 == 0 {
			tr.records[int64(i)] = models.Transfer{ID: int64(i), FromAccountID: 1, ToAccountID: 2, Amount: int64(i), Date: dashboardDay(i)}
		} else {
			tx.records[int64(i)] = models.Transaction{ID: int64(i), Type: "income", AccountID: 1, Amount: int64(i), Date: dashboardDay(i)}
		}
	}
	result, err := s.GetDashboard(dashboardDay(1), dashboardDay(31))
	if err != nil {
		t.Fatal(err)
	}
	if len(result.RecentMovements) != 10 {
		t.Fatalf("recent count: %d", len(result.RecentMovements))
	}
	for i, movement := range result.RecentMovements {
		if movement.ID != int64(15-i) {
			t.Fatalf("recent order: %+v", result.RecentMovements)
		}
		if movement.ID%3 == 0 && (movement.Kind != "transfer" || movement.AccountID != 1 || movement.ToAccountID != 2) {
			t.Fatalf("transfer representation: %+v", movement)
		}
	}
}

func TestDashboardDefaultsAndErrors(t *testing.T) {
	s, a, c, tx, tr := dashboardFixture()
	result, err := s.GetDashboard(time.Time{}, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	if result.Period.From != start.Format(time.DateOnly) || result.Period.To != start.AddDate(0, 1, 0).Add(-time.Nanosecond).Format(time.DateOnly) {
		t.Fatalf("default period: %+v", result.Period)
	}
	if result.Accounts == nil || result.BalancesByCurrency == nil || result.SummaryByCurrency == nil || result.IncomeByCategory == nil || result.ExpensesByCategory == nil || result.RecentMovements == nil || *result.Summary.TotalBalance != 0 {
		t.Fatal("empty dashboard must contain arrays and zero totals")
	}
	_, err = s.GetDashboard(dashboardDay(7), dashboardDay(6))
	var validation *models.ValidationError
	if !errors.As(err, &validation) {
		t.Fatalf("date range validation: %v", err)
	}
	failure := errors.New("repository unavailable")
	for _, failures := range []map[string]error{a.errors, c.failures, tx.failures, tr.failures} {
		failures["all"] = failure
		_, err := s.GetDashboard(dashboardDay(1), dashboardDay(31))
		if !errors.Is(err, failure) {
			t.Fatalf("repository error not propagated: %v", err)
		}
		delete(failures, "all")
	}
	a.accounts[1] = models.Account{ID: 1, Currency: "ARS", InitialBalance: math.MaxInt64, Active: true}
	tx.records[1] = models.Transaction{ID: 1, AccountID: 1, Type: "income", Amount: 1, Date: dashboardDay(1)}
	if _, err := s.GetDashboard(dashboardDay(1), dashboardDay(31)); !errors.Is(err, ErrMoneyOverflow) {
		t.Fatalf("money overflow not detected: %v", err)
	}
}

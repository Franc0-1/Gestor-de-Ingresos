package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"finance-app/internal/models"
	"finance-app/internal/repository"
	"finance-app/internal/services"
)

func newTestDashboardHandler() (*http.ServeMux, *handlerRepository, *transactionHandlerRepository) {
	a := &handlerRepository{accounts: map[int64]models.Account{1: {ID: 1, Name: "Cash", Currency: "ARS", InitialBalance: 100, Active: true}}}
	c := &categoryHandlerRepository{records: map[int64]models.Category{1: {ID: 1, Name: "Salary", Type: "income"}}}
	tx := &transactionHandlerRepository{records: make(map[int64]models.Transaction)}
	tr := &transferHandlerRepository{records: make(map[int64]models.Transfer)}
	mux := http.NewServeMux()
	NewDashboardHandler(services.NewDashboardService(a, c, tx, tr)).RegisterRoutes(mux)
	return mux, a, tx
}

func TestDashboardHandlerDefaultAndExplicitPeriod(t *testing.T) {
	mux, _, tx := newTestDashboardHandler()
	w := request(t, mux, "GET", "/api/dashboard", "", http.StatusOK)
	var dashboard models.DashboardResponse
	if err := json.Unmarshal(w.Body.Bytes(), &dashboard); err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	if dashboard.Period.From != start.Format(time.DateOnly) || dashboard.Period.To != start.AddDate(0, 1, 0).Add(-time.Nanosecond).Format(time.DateOnly) || *dashboard.Summary.TotalBalance != 100 {
		t.Fatalf("default response: %+v", dashboard)
	}
	end := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC).Add(-time.Nanosecond)
	tx.records[1] = models.Transaction{ID: 1, Type: "income", AccountID: 1, CategoryID: 1, Amount: 23, Date: end}
	w = request(t, mux, "GET", "/api/dashboard?from=2026-10-01&to=2026-10-31", "", http.StatusOK)
	if err := json.Unmarshal(w.Body.Bytes(), &dashboard); err != nil {
		t.Fatal(err)
	}
	if dashboard.Period.From != "2026-10-01" || dashboard.Period.To != "2026-10-31" || *dashboard.Summary.TotalBalance != 123 || *dashboard.Summary.PeriodIncome != 23 || len(dashboard.RecentMovements) != 1 || dashboard.Accounts[0].AccountName != "Cash" {
		t.Fatalf("explicit response: %+v", dashboard)
	}
}

func TestDashboardHandlerBadDates(t *testing.T) {
	for _, query := range []string{"from=bad", "to=2026-02-30", "from=2026-10-01T00:00:00Z", "from=", "to=", "from=2026-10-07&to=2026-10-06"} {
		t.Run(query, func(t *testing.T) {
			mux, _, _ := newTestDashboardHandler()
			request(t, mux, "GET", "/api/dashboard?"+query, "", http.StatusBadRequest)
		})
	}
}

func TestDashboardHandlerCurrenciesAndInternalErrors(t *testing.T) {
	mux, a, _ := newTestDashboardHandler()
	a.accounts[2] = models.Account{ID: 2, Name: "USD", Currency: "USD", InitialBalance: 200, Active: true}
	w := request(t, mux, "GET", "/api/dashboard", "", http.StatusOK)
	var body struct {
		Summary  map[string]json.RawMessage `json:"summary"`
		Balances []models.CurrencyBalance   `json:"balancesByCurrency"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if _, ok := body.Summary["totalBalance"]; ok {
		t.Fatal("misleading mixed-currency total exposed")
	}
	if len(body.Balances) != 2 {
		t.Fatalf("missing currency balances: %s", w.Body.String())
	}
	for _, failure := range []error{errors.New("sensitive workbook path"), repository.ErrNotFound} {
		a.err = failure
		w = request(t, mux, "GET", "/api/dashboard", "", http.StatusInternalServerError)
		if strings.TrimSpace(w.Body.String()) != `{"error":"internal server error"}` {
			t.Fatalf("internal error leaked: %s", w.Body.String())
		}
	}
}

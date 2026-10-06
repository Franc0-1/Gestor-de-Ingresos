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

func fakePagination(total, page, size int) models.Pagination {
	pages := total / size
	if total%size != 0 {
		pages++
	}
	return models.Pagination{Page: page, PageSize: size, TotalItems: total, TotalPages: pages}
}

func (f *transactionHandlerRepository) Query(q models.TransactionQuery) (models.TransactionPage, error) {
	items, err := f.GetAll()
	if items == nil {
		items = []models.Transaction{}
	}
	return models.TransactionPage{Items: items, Pagination: fakePagination(len(items), q.Page, q.PageSize)}, err
}
func (f *transferHandlerRepository) Query(q models.TransferQuery) (models.TransferPage, error) {
	items, err := f.GetAll()
	if items == nil {
		items = []models.Transfer{}
	}
	return models.TransferPage{Items: items, Pagination: fakePagination(len(items), q.Page, q.PageSize)}, err
}

type transactionQueryRepository struct {
	repository.TransactionRepository
	query  models.TransactionQuery
	calls  int
	result models.TransactionPage
	err    error
}

func (f *transactionQueryRepository) Query(q models.TransactionQuery) (models.TransactionPage, error) {
	f.query = q
	f.calls++
	return f.result, f.err
}

type transferQueryRepository struct {
	repository.TransferRepository
	query  models.TransferQuery
	calls  int
	result models.TransferPage
	err    error
}

func (f *transferQueryRepository) Query(q models.TransferQuery) (models.TransferPage, error) {
	f.query = q
	f.calls++
	return f.result, f.err
}

func queryHandlerFixture() (*http.ServeMux, *transactionQueryRepository, *transferQueryRepository) {
	txRepo := &transactionQueryRepository{result: models.TransactionPage{Items: []models.Transaction{}, Pagination: models.Pagination{Page: 2, PageSize: 3, TotalItems: 7, TotalPages: 3}}}
	trRepo := &transferQueryRepository{result: models.TransferPage{Items: []models.Transfer{}, Pagination: models.Pagination{Page: 2, PageSize: 3, TotalItems: 7, TotalPages: 3}}}
	mux := http.NewServeMux()
	NewTransactionHandler(services.NewTransactionService(txRepo, nil, nil)).RegisterRoutes(mux)
	NewTransferHandler(services.NewTransferService(trRepo, nil)).RegisterRoutes(mux)
	return mux, txRepo, trRepo
}

func TestQueryHandlersParsingAndResponse(t *testing.T) {
	mux, txRepo, trRepo := queryHandlerFixture()
	common := "from=2026-10-01&to=2026-10-31&search=SuPeR&page=2&pageSize=3&sortBy=amount&sortOrder=ASC"
	w := request(t, mux, "GET", "/api/transactions?type=expense&accountId=9223372036854775807&categoryId=2&"+common, "", http.StatusOK)
	var page models.TransactionPage
	if err := json.Unmarshal(w.Body.Bytes(), &page); err != nil {
		t.Fatal(err)
	}
	if page.Items == nil || page.Pagination != txRepo.result.Pagination {
		t.Fatalf("incorrect paginated response: %+v", page)
	}
	q := txRepo.query
	from := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC).Add(-time.Nanosecond)
	if q.Type != "expense" || q.AccountID != 9223372036854775807 || q.CategoryID != 2 || !q.DateFrom.Equal(from) || !q.DateTo.Equal(to) || q.Search != "SuPeR" || q.SortBy != "amount" || q.SortOrder != "asc" || q.Page != 2 || q.PageSize != 3 {
		t.Fatalf("transaction parsing: %+v", q)
	}
	w = request(t, mux, "GET", "/api/transfers?fromAccountId=1&toAccountId=2&"+common, "", http.StatusOK)
	var transfers models.TransferPage
	if err := json.Unmarshal(w.Body.Bytes(), &transfers); err != nil {
		t.Fatal(err)
	}
	if transfers.Items == nil || transfers.Pagination != trRepo.result.Pagination {
		t.Fatalf("incorrect transfer page: %+v", transfers)
	}
	tq := trRepo.query
	if tq.FromAccountID != 1 || tq.ToAccountID != 2 || !tq.DateFrom.Equal(from) || !tq.DateTo.Equal(to) || tq.Search != "SuPeR" || tq.SortBy != "amount" || tq.SortOrder != "asc" || tq.Page != 2 || tq.PageSize != 3 {
		t.Fatalf("transfer parsing: %+v", tq)
	}
}

func TestQueryHandlersDefaults(t *testing.T) {
	for _, path := range []string{"/api/transactions", "/api/transfers"} {
		for _, suffix := range []string{"", "?page=0&pageSize=-1", "?page=-5&pageSize=101", "?from=2026-10-06&to=2026-10-06"} {
			mux, txRepo, trRepo := queryHandlerFixture()
			request(t, mux, "GET", path+suffix, "", http.StatusOK)
			page, size, by, order := txRepo.query.Page, txRepo.query.PageSize, txRepo.query.SortBy, txRepo.query.SortOrder
			if path == "/api/transfers" {
				page, size, by, order = trRepo.query.Page, trRepo.query.PageSize, trRepo.query.SortBy, trRepo.query.SortOrder
			}
			wantSize := 20
			if strings.Contains(suffix, "101") {
				wantSize = 100
			}
			if page != 1 || size != wantSize || by != "date" || order != "desc" {
				t.Fatalf("defaults for %s%s: %d %d %s %s", path, suffix, page, size, by, order)
			}
		}
	}
}

func TestQueryHandlersBadRequests(t *testing.T) {
	common := []string{
		"page=abc", "page=1.5", "page=", "page=9999999999999999999999",
		"pageSize=abc", "pageSize=1.5", "pageSize=", "pageSize=9999999999999999999999",
		"from=bad", "to=2026-02-30", "from=2026-10-01T00:00:00Z", "from=", "to=",
		"from=2026-10-07&to=2026-10-06", "sortBy=description", "sortOrder=random",
	}
	for _, path := range []string{"/api/transactions", "/api/transfers"} {
		invalid := append([]string{}, common...)
		if path == "/api/transactions" {
			invalid = append(invalid, "type=transfer", "accountId=abc", "accountId=0", "accountId=-1", "accountId=9223372036854775808", "categoryId=abc", "categoryId=0", "categoryId=-1")
		} else {
			invalid = append(invalid, "sortBy=type", "fromAccountId=abc", "fromAccountId=0", "fromAccountId=-1", "toAccountId=abc", "toAccountId=0", "toAccountId=-1", "toAccountId=9223372036854775808")
		}
		for _, raw := range invalid {
			t.Run(path+raw, func(t *testing.T) {
				mux, txRepo, trRepo := queryHandlerFixture()
				request(t, mux, "GET", path+"?"+raw, "", http.StatusBadRequest)
				if txRepo.calls != 0 || trRepo.calls != 0 {
					t.Fatal("invalid query reached repository")
				}
			})
		}
	}
}

func TestQueryHandlersInternalErrors(t *testing.T) {
	for _, path := range []string{"/api/transactions", "/api/transfers"} {
		mux, txRepo, trRepo := queryHandlerFixture()
		failure := errors.New("sensitive workbook path")
		txRepo.err = failure
		trRepo.err = failure
		w := request(t, mux, "GET", path, "", http.StatusInternalServerError)
		if strings.TrimSpace(w.Body.String()) != `{"error":"internal server error"}` {
			t.Fatalf("internal details leaked: %s", w.Body.String())
		}
	}
}

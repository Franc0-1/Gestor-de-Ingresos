package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"finance-app/internal/models"
	"finance-app/internal/repository"
	"finance-app/internal/services"
)

type transactionHandlerRepository struct {
	records map[int64]models.Transaction
	nextID  int64
	err     error
}

func (f *transactionHandlerRepository) Create(tx models.Transaction) (models.Transaction, error) {
	if f.err != nil {
		return models.Transaction{}, f.err
	}
	f.nextID++
	tx.ID = f.nextID
	f.records[tx.ID] = tx
	return tx, nil
}
func (f *transactionHandlerRepository) GetAll() ([]models.Transaction, error) {
	if f.err != nil {
		return nil, f.err
	}
	var txs []models.Transaction
	for _, tx := range f.records {
		txs = append(txs, tx)
	}
	return txs, nil
}
func (f *transactionHandlerRepository) GetByID(id int64) (models.Transaction, error) {
	if f.err != nil {
		return models.Transaction{}, f.err
	}
	tx, ok := f.records[id]
	if !ok {
		return models.Transaction{}, fmt.Errorf("ID %d: %w", id, repository.ErrNotFound)
	}
	return tx, nil
}
func (f *transactionHandlerRepository) Update(tx models.Transaction) error {
	if f.err != nil {
		return f.err
	}
	if _, ok := f.records[tx.ID]; !ok {
		return repository.ErrNotFound
	}
	// Simulate the repository assigning its own timestamp when persisting.
	tx.UpdatedAt = tx.UpdatedAt.Add(time.Millisecond)
	f.records[tx.ID] = tx
	return nil
}
func (f *transactionHandlerRepository) Delete(id int64) error {
	if f.err != nil {
		return f.err
	}
	if _, ok := f.records[id]; !ok {
		return repository.ErrNotFound
	}
	delete(f.records, id)
	return nil
}

func newTestTransactionHandler() (*http.ServeMux, *transactionHandlerRepository, *handlerRepository, *categoryHandlerRepository) {
	txs := &transactionHandlerRepository{records: make(map[int64]models.Transaction)}
	accounts := &handlerRepository{accounts: map[int64]models.Account{1: {ID: 1, Name: "Cash", Type: "cash", Currency: "ARS"}}}
	categories := &categoryHandlerRepository{records: map[int64]models.Category{1: {ID: 1, Name: "Salary", Type: "income"}, 2: {ID: 2, Name: "Food", Type: "expense"}}}
	mux := http.NewServeMux()
	NewTransactionHandler(services.NewTransactionService(txs, accounts, categories)).RegisterRoutes(mux)
	return mux, txs, accounts, categories
}

const validTransactionBody = `{"type":"expense","amount":1850000,"description":"Supermercado","categoryId":2,"accountId":1,"date":"2026-10-06T00:00:00Z"}`

func responseTransaction(t *testing.T, w *httptest.ResponseRecorder) models.Transaction {
	t.Helper()
	var tx models.Transaction
	if err := json.Unmarshal(w.Body.Bytes(), &tx); err != nil {
		t.Fatal(err)
	}
	return tx
}

func TestTransactionHandlerCRUD(t *testing.T) {
	mux, repo, _, _ := newTestTransactionHandler()
	w := request(t, mux, "GET", "/api/transactions", "", http.StatusOK)
	if strings.TrimSpace(w.Body.String()) != `{"items":[],"pagination":{"page":1,"pageSize":20,"totalItems":0,"totalPages":0}}` {
		t.Fatalf("expected empty page: %s", w.Body.String())
	}
	body := `{"id":99,"type":" expense ","amount":9223372036854775807,"description":" Supermercado ","categoryId":2,"accountId":1,"date":"2026-10-06T00:00:00.123456789-03:00","createdAt":"client value","updatedAt":"client value"}`
	w = request(t, mux, "POST", "/api/transactions", body, http.StatusCreated)
	created := responseTransaction(t, w)
	expectedDate := time.Date(2026, 10, 6, 3, 0, 0, 123456789, time.UTC)
	if created.ID != 1 || created.Type != "expense" || created.Amount != 9223372036854775807 || created.Description != "Supermercado" || created.CategoryID != 2 || created.AccountID != 1 || !created.Date.Equal(expectedDate) || created.CreatedAt.IsZero() || !created.CreatedAt.Equal(created.UpdatedAt) {
		t.Fatalf("incorrect creation: %+v", created)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(w.Body.Bytes(), &fields); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"id", "type", "amount", "description", "categoryId", "accountId", "date", "createdAt", "updatedAt"} {
		if _, ok := fields[key]; !ok {
			t.Errorf("missing JSON field %s", key)
		}
	}
	if len(fields) != 9 {
		t.Fatalf("unexpected JSON fields: %v", fields)
	}
	w = request(t, mux, "GET", "/api/transactions/1", "", http.StatusOK)
	if got := responseTransaction(t, w); got != created {
		t.Fatalf("GET by ID: %+v", got)
	}
	w = request(t, mux, "GET", "/api/transactions", "", http.StatusOK)
	var page models.TransactionPage
	if err := json.Unmarshal(w.Body.Bytes(), &page); err != nil {
		t.Fatal(err)
	}
	if len(page.Items) != 1 || page.Items[0] != created {
		t.Fatalf("GET all: %+v", page)
	}
	body = `{"id":999,"type":"income","amount":500,"categoryId":1,"accountId":1,"date":"2026-10-07T00:00:00Z","createdAt":"2000-01-01T00:00:00Z","updatedAt":"2000-01-01T00:00:00Z"}`
	w = request(t, mux, "PUT", "/api/transactions/1", body, http.StatusOK)
	updated := responseTransaction(t, w)
	if updated.ID != 1 || updated.Type != "income" || updated.Amount != 500 || updated.CategoryID != 1 || updated.Description != "" || !updated.CreatedAt.Equal(created.CreatedAt) || !updated.UpdatedAt.After(created.UpdatedAt) || updated != repo.records[1] {
		t.Fatalf("incorrect persisted update: %+v", updated)
	}
	request(t, mux, "DELETE", "/api/transactions/1", "", http.StatusNoContent)
	request(t, mux, "GET", "/api/transactions/1", "", http.StatusNotFound)
	request(t, mux, "PUT", "/api/transactions/1", validTransactionBody, http.StatusNotFound)
	request(t, mux, "DELETE", "/api/transactions/1", "", http.StatusNotFound)
}

func TestTransactionHandlerInvalidJSON(t *testing.T) {
	for _, body := range []string{
		"", "{", "null", "[]", "true", `"text"`,
		validTransactionBody + ` {}`, validTransactionBody + ` null`, validTransactionBody + ` trailing`,
		strings.Replace(validTransactionBody, `1850000`, `1.5`, 1),
		strings.Replace(validTransactionBody, `1850000`, `9223372036854775808`, 1),
		strings.Replace(validTransactionBody, `1850000`, `"100"`, 1),
		strings.Replace(validTransactionBody, `"categoryId":2`, `"categoryId":"2"`, 1),
		strings.Replace(validTransactionBody, `"accountId":1`, `"accountId":1.5`, 1),
		strings.Replace(validTransactionBody, `2026-10-06T00:00:00Z`, `invalid date`, 1),
		strings.Replace(validTransactionBody, `2026-10-06T00:00:00Z`, `2026-10-06`, 1),
	} {
		for _, method := range []string{"POST", "PUT"} {
			t.Run(method+body, func(t *testing.T) {
				mux, _, _, _ := newTestTransactionHandler()
				path := "/api/transactions"
				if method == "PUT" {
					path += "/1"
				}
				request(t, mux, method, path, body, http.StatusBadRequest)
			})
		}
	}
}

func TestTransactionHandlerInvalidData(t *testing.T) {
	for _, body := range []string{
		`{}`,
		strings.Replace(validTransactionBody, `"expense"`, `" "`, 1),
		strings.Replace(validTransactionBody, `"expense"`, `"transfer"`, 1),
		strings.Replace(validTransactionBody, `1850000`, `0`, 1),
		strings.Replace(validTransactionBody, `1850000`, `-1`, 1),
		strings.Replace(validTransactionBody, `"accountId":1`, `"accountId":0`, 1),
		strings.Replace(validTransactionBody, `"categoryId":2`, `"categoryId":0`, 1),
		strings.Replace(validTransactionBody, `,"date":"2026-10-06T00:00:00Z"`, ``, 1),
		strings.Replace(validTransactionBody, `"date":"2026-10-06T00:00:00Z"`, `"date":null`, 1),
		strings.Replace(validTransactionBody, `"accountId":1`, `"accountId":99`, 1),
		strings.Replace(validTransactionBody, `"categoryId":2`, `"categoryId":99`, 1),
		strings.Replace(validTransactionBody, `"categoryId":2`, `"categoryId":1`, 1),
		strings.Replace(validTransactionBody, `"expense"`, `"income"`, 1),
	} {
		for _, method := range []string{"POST", "PUT"} {
			t.Run(method+body, func(t *testing.T) {
				mux, repo, _, _ := newTestTransactionHandler()
				created := responseTransaction(t, request(t, mux, "POST", "/api/transactions", validTransactionBody, http.StatusCreated))
				path := "/api/transactions"
				if method == "PUT" {
					path += "/1"
				}
				request(t, mux, method, path, body, http.StatusBadRequest)
				if len(repo.records) != 1 || repo.records[1] != created {
					t.Fatal("invalid input modified storage")
				}
			})
		}
	}
}

func TestTransactionHandlerIDsAndNotFound(t *testing.T) {
	for _, method := range []string{"GET", "PUT", "DELETE"} {
		for _, id := range []string{"abc", "0", "-1", "1.5", "9223372036854775808", "42"} {
			t.Run(method+id, func(t *testing.T) {
				mux, _, _, _ := newTestTransactionHandler()
				status := http.StatusBadRequest
				if id == "42" {
					status = http.StatusNotFound
				}
				request(t, mux, method, "/api/transactions/"+id, validTransactionBody, status)
			})
		}
	}
}

func TestTransactionHandlerInternalErrors(t *testing.T) {
	for _, endpoint := range []struct{ method, path string }{
		{"GET", "/api/transactions"}, {"GET", "/api/transactions/1"},
		{"POST", "/api/transactions"}, {"PUT", "/api/transactions/1"}, {"DELETE", "/api/transactions/1"},
	} {
		t.Run(endpoint.method+endpoint.path, func(t *testing.T) {
			mux, repo, _, _ := newTestTransactionHandler()
			repo.err = errors.New("sensitive workbook path")
			w := request(t, mux, endpoint.method, endpoint.path, validTransactionBody, http.StatusInternalServerError)
			if strings.TrimSpace(w.Body.String()) != `{"error":"internal server error"}` {
				t.Fatalf("internal details leaked: %s", w.Body.String())
			}
		})
	}
	for _, entity := range []string{"account", "category"} {
		for _, method := range []string{"POST", "PUT"} {
			t.Run(entity+method, func(t *testing.T) {
				mux, _, accounts, categories := newTestTransactionHandler()
				request(t, mux, "POST", "/api/transactions", validTransactionBody, http.StatusCreated)
				failure := errors.New("sensitive lookup details")
				if entity == "account" {
					accounts.err = failure
				} else {
					categories.err = failure
				}
				path := "/api/transactions"
				if method == "PUT" {
					path += "/1"
				}
				w := request(t, mux, method, path, validTransactionBody, http.StatusInternalServerError)
				if strings.Contains(w.Body.String(), "sensitive") {
					t.Fatal("internal lookup details leaked")
				}
			})
		}
	}
}

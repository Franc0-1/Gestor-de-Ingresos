package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
	"time"

	"finance-app/internal/models"
	"finance-app/internal/repository"
	"finance-app/internal/services"
)

type handlerRepository struct {
	accounts map[int64]models.Account
	nextID   int64
	err      error
}

func (f *handlerRepository) Create(a models.Account) (models.Account, error) {
	if f.err != nil {
		return models.Account{}, f.err
	}
	f.nextID++
	a.ID = f.nextID
	f.accounts[a.ID] = a
	return a, nil
}
func (f *handlerRepository) GetAll() ([]models.Account, error) {
	if f.err != nil {
		return nil, f.err
	}
	var accounts []models.Account
	for _, a := range f.accounts {
		accounts = append(accounts, a)
	}
	slices.SortFunc(accounts, func(a, b models.Account) int {
		if a.ID < b.ID {
			return -1
		}
		if a.ID > b.ID {
			return 1
		}
		return 0
	})
	return accounts, nil
}
func (f *handlerRepository) GetByID(id int64) (models.Account, error) {
	if f.err != nil {
		return models.Account{}, f.err
	}
	a, ok := f.accounts[id]
	if !ok {
		return models.Account{}, fmt.Errorf("missing ID %d: %w", id, repository.ErrNotFound)
	}
	return a, nil
}
func (f *handlerRepository) Update(a models.Account) error {
	if f.err != nil {
		return f.err
	}
	if _, ok := f.accounts[a.ID]; !ok {
		return repository.ErrNotFound
	}
	// Simulate the repository assigning a timestamp at persistence time.
	a.UpdatedAt = a.UpdatedAt.Add(time.Millisecond)
	f.accounts[a.ID] = a
	return nil
}
func (f *handlerRepository) Delete(id int64) error {
	if f.err != nil {
		return f.err
	}
	if _, ok := f.accounts[id]; !ok {
		return repository.ErrNotFound
	}
	delete(f.accounts, id)
	return nil
}

func newTestHandler() (*http.ServeMux, *handlerRepository) {
	repo := &handlerRepository{accounts: make(map[int64]models.Account)}
	mux := http.NewServeMux()
	NewAccountHandler(services.NewAccountService(repo)).RegisterRoutes(mux)
	return mux, repo
}

const validBody = `{"name":"Mercado Pago","type":"wallet","currency":"ARS","initialBalance":1500000,"active":true}`

func request(t *testing.T, mux *http.ServeMux, method, path, body string, status int) *httptest.ResponseRecorder {
	t.Helper()
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)
	if w.Code != status {
		t.Fatalf("%s %s: status %d, want %d; body: %s", method, path, w.Code, status, w.Body.String())
	}
	if status == http.StatusNoContent {
		if w.Body.Len() != 0 {
			t.Fatal("204 response has a body")
		}
	} else if w.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("incorrect Content-Type: %s", w.Header().Get("Content-Type"))
	}
	if status >= 400 {
		var fields map[string]any
		if err := json.Unmarshal(w.Body.Bytes(), &fields); err != nil {
			t.Fatal(err)
		}
		message, ok := fields["error"].(string)
		if len(fields) != 1 || !ok || message == "" {
			t.Fatalf("inconsistent error response: %v", fields)
		}
	}
	return w
}

func responseAccount(t *testing.T, w *httptest.ResponseRecorder) models.Account {
	t.Helper()
	var a models.Account
	if err := json.Unmarshal(w.Body.Bytes(), &a); err != nil {
		t.Fatal(err)
	}
	return a
}

func TestAccountHandlerCRUD(t *testing.T) {
	mux, repo := newTestHandler()
	w := request(t, mux, "GET", "/api/accounts", "", http.StatusOK)
	if strings.TrimSpace(w.Body.String()) != "[]" {
		t.Fatalf("empty collection must be []: %s", w.Body.String())
	}
	body := `{"id":99,"name":" Mercado Pago ","type":" WALLET ","currency":" ars ","initialBalance":9223372036854775807,"active":true,"createdAt":"client value","updatedAt":"client value"}`
	w = request(t, mux, "POST", "/api/accounts", body, http.StatusCreated)
	created := responseAccount(t, w)
	if created.ID != 1 || created.Name != "Mercado Pago" || created.Type != "wallet" || created.Currency != "ARS" || created.InitialBalance != 9223372036854775807 || !created.Active || created.CreatedAt.IsZero() || created.UpdatedAt.IsZero() {
		t.Fatalf("incorrect created account: %+v", created)
	}
	var fields map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &fields); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"id", "name", "type", "currency", "initialBalance", "active", "createdAt", "updatedAt"} {
		if _, ok := fields[key]; !ok {
			t.Errorf("missing JSON field %s", key)
		}
	}
	if len(fields) != 8 {
		t.Fatalf("unexpected fields: %v", fields)
	}
	w = request(t, mux, "GET", "/api/accounts/1", "", http.StatusOK)
	got := responseAccount(t, w)
	if got != created {
		t.Fatalf("GET by ID: %+v", got)
	}
	w = request(t, mux, "GET", "/api/accounts", "", http.StatusOK)
	var all []models.Account
	if err := json.Unmarshal(w.Body.Bytes(), &all); err != nil {
		t.Fatal(err)
	}
	if len(all) != 1 || all[0] != created {
		t.Fatalf("GET all: %+v", all)
	}
	body = `{"id":999,"name":"Banco","type":"bank","currency":"USD","initialBalance":-1050025,"active":false,"createdAt":"2000-01-01T00:00:00Z","updatedAt":"2000-01-01T00:00:00Z"}`
	w = request(t, mux, "PUT", "/api/accounts/1", body, http.StatusOK)
	updated := responseAccount(t, w)
	if updated.ID != 1 || updated.Name != "Banco" || updated.Type != "bank" || updated.Currency != "USD" || updated.InitialBalance != -1050025 || updated.Active || !updated.CreatedAt.Equal(created.CreatedAt) || !updated.UpdatedAt.After(created.UpdatedAt) {
		t.Fatalf("incorrect updated account: %+v", updated)
	}
	if repo.accounts[1].ID != 1 {
		t.Fatal("body ID overrode URL ID")
	}
	request(t, mux, "DELETE", "/api/accounts/1", "", http.StatusNoContent)
	request(t, mux, "GET", "/api/accounts/1", "", http.StatusNotFound)
	request(t, mux, "PUT", "/api/accounts/1", validBody, http.StatusNotFound)
	request(t, mux, "DELETE", "/api/accounts/1", "", http.StatusNotFound)
}

func TestAccountHandlerInvalidJSON(t *testing.T) {
	for _, body := range []string{
		"", " ", "{", "null", "[]", "true", `"text"`,
		validBody + ` {}`, validBody + ` null`, validBody + ` trailing`,
		`{"name":123}`, `{"active":"true"}`,
		`{"name":"Cash","type":"cash","currency":"ARS","initialBalance":1.5}`,
		`{"name":"Cash","type":"cash","currency":"ARS","initialBalance":9223372036854775808}`,
		`{"name":"Cash","type":"cash","currency":"ARS","initialBalance":"100"}`,
	} {
		for _, method := range []string{"POST", "PUT"} {
			t.Run(method+body, func(t *testing.T) {
				mux, repo := newTestHandler()
				path := "/api/accounts"
				if method == "PUT" {
					path += "/1"
				}
				request(t, mux, method, path, body, http.StatusBadRequest)
				if len(repo.accounts) != 0 {
					t.Fatal("invalid JSON created a record")
				}
			})
		}
	}
}

func TestAccountHandlerInvalidData(t *testing.T) {
	for _, body := range []string{
		`{}`, `{"name":" ","type":"cash","currency":"ARS"}`,
		`{"name":"Cash","type":"invalid","currency":"ARS"}`,
		`{"name":"Cash","type":"cash","currency":" "}`,
	} {
		for _, method := range []string{"POST", "PUT"} {
			t.Run(method+body, func(t *testing.T) {
				mux, _ := newTestHandler()
				path := "/api/accounts"
				if method == "PUT" {
					path += "/1"
				}
				request(t, mux, method, path, body, http.StatusBadRequest)
			})
		}
	}
}

func TestAccountHandlerInvalidID(t *testing.T) {
	for _, id := range []string{"abc", "0", "-1", "1.5", "9223372036854775808"} {
		for _, method := range []string{"GET", "PUT", "DELETE"} {
			t.Run(method+id, func(t *testing.T) {
				mux, _ := newTestHandler()
				request(t, mux, method, "/api/accounts/"+id, validBody, http.StatusBadRequest)
			})
		}
	}
}

func TestAccountHandlerNotFound(t *testing.T) {
	for _, method := range []string{"GET", "PUT", "DELETE"} {
		t.Run(method, func(t *testing.T) {
			mux, _ := newTestHandler()
			request(t, mux, method, "/api/accounts/42", validBody, http.StatusNotFound)
		})
	}
}

func TestAccountHandlerInternalErrors(t *testing.T) {
	for _, endpoint := range []struct{ method, path string }{
		{"GET", "/api/accounts"}, {"GET", "/api/accounts/1"},
		{"POST", "/api/accounts"}, {"PUT", "/api/accounts/1"}, {"DELETE", "/api/accounts/1"},
	} {
		t.Run(endpoint.method+endpoint.path, func(t *testing.T) {
			mux, repo := newTestHandler()
			repo.err = errors.New("sensitive workbook path and internal details")
			w := request(t, mux, endpoint.method, endpoint.path, validBody, http.StatusInternalServerError)
			if strings.Contains(w.Body.String(), "sensitive") || strings.TrimSpace(w.Body.String()) != `{"error":"internal server error"}` {
				t.Fatalf("internal error leaked: %s", w.Body.String())
			}
		})
	}
}

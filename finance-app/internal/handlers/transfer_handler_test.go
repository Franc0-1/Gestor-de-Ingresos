package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"finance-app/internal/models"
	"finance-app/internal/repository"
	"finance-app/internal/services"
)

type transferHandlerRepository struct {
	records map[int64]models.Transfer
	nextID  int64
	err     error
}

func (f *transferHandlerRepository) Create(tr models.Transfer) (models.Transfer, error) {
	if f.err != nil {
		return models.Transfer{}, f.err
	}
	f.nextID++
	tr.ID = f.nextID
	f.records[tr.ID] = tr
	return tr, nil
}
func (f *transferHandlerRepository) GetAll() ([]models.Transfer, error) {
	if f.err != nil {
		return nil, f.err
	}
	var result []models.Transfer
	for _, tr := range f.records {
		result = append(result, tr)
	}
	return result, nil
}
func (f *transferHandlerRepository) GetByID(id int64) (models.Transfer, error) {
	if f.err != nil {
		return models.Transfer{}, f.err
	}
	tr, ok := f.records[id]
	if !ok {
		return models.Transfer{}, fmt.Errorf("ID %d: %w", id, repository.ErrNotFound)
	}
	return tr, nil
}
func (f *transferHandlerRepository) Update(tr models.Transfer) error {
	if f.err != nil {
		return f.err
	}
	if _, ok := f.records[tr.ID]; !ok {
		return repository.ErrNotFound
	}
	tr.UpdatedAt = tr.UpdatedAt.Add(time.Millisecond)
	f.records[tr.ID] = tr
	return nil
}
func (f *transferHandlerRepository) Delete(id int64) error {
	if f.err != nil {
		return f.err
	}
	if _, ok := f.records[id]; !ok {
		return repository.ErrNotFound
	}
	delete(f.records, id)
	return nil
}

func newTestTransferHandler() (*http.ServeMux, *transferHandlerRepository, *handlerRepository) {
	repo := &transferHandlerRepository{records: make(map[int64]models.Transfer)}
	accounts := &handlerRepository{accounts: map[int64]models.Account{
		1: {ID: 1, Name: "Cash", Type: "cash", Currency: "ARS", Active: true},
		2: {ID: 2, Name: "Wallet", Type: "wallet", Currency: "ARS", Active: true},
	}}
	mux := http.NewServeMux()
	NewTransferHandler(services.NewTransferService(repo, accounts)).RegisterRoutes(mux)
	return mux, repo, accounts
}

const validTransferBody = `{"fromAccountId":1,"toAccountId":2,"amount":500000,"description":"Mover dinero","date":"2026-10-06T00:00:00Z"}`

func TestTransferHandlerCRUD(t *testing.T) {
	mux, repo, _ := newTestTransferHandler()
	w := request(t, mux, "GET", "/api/transfers", "", http.StatusOK)
	if strings.TrimSpace(w.Body.String()) != `{"items":[],"pagination":{"page":1,"pageSize":20,"totalItems":0,"totalPages":0}}` {
		t.Fatalf("expected empty page: %s", w.Body.String())
	}
	body := `{"id":99,"fromAccountId":1,"toAccountId":2,"amount":9223372036854775807,"description":" Mover dinero ","date":"2026-10-06T00:00:00.123456789-03:00","createdAt":"client value","updatedAt":"client value"}`
	w = request(t, mux, "POST", "/api/transfers", body, http.StatusCreated)
	var created models.Transfer
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	expectedDate := time.Date(2026, 10, 6, 3, 0, 0, 123456789, time.UTC)
	if created.ID != 1 || created.FromAccountID != 1 || created.ToAccountID != 2 || created.Amount != 9223372036854775807 || created.Description != "Mover dinero" || !created.Date.Equal(expectedDate) || created.CreatedAt.IsZero() || !created.CreatedAt.Equal(created.UpdatedAt) {
		t.Fatalf("incorrect creation: %+v", created)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(w.Body.Bytes(), &fields); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"id", "fromAccountId", "toAccountId", "amount", "description", "date", "createdAt", "updatedAt"} {
		if _, ok := fields[key]; !ok {
			t.Errorf("missing JSON field %s", key)
		}
	}
	if len(fields) != 8 {
		t.Fatalf("unexpected JSON fields: %v", fields)
	}
	w = request(t, mux, "GET", "/api/transfers/1", "", http.StatusOK)
	var got models.Transfer
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil || got != created {
		t.Fatalf("GET by ID: %+v %v", got, err)
	}
	w = request(t, mux, "GET", "/api/transfers", "", http.StatusOK)
	var page models.TransferPage
	if err := json.Unmarshal(w.Body.Bytes(), &page); err != nil || len(page.Items) != 1 || page.Items[0] != created {
		t.Fatalf("GET all: %+v %v", page, err)
	}
	body = `{"id":999,"fromAccountId":2,"toAccountId":1,"amount":100,"date":"2026-10-07T00:00:00Z","createdAt":"2000-01-01T00:00:00Z","updatedAt":"2000-01-01T00:00:00Z"}`
	w = request(t, mux, "PUT", "/api/transfers/1", body, http.StatusOK)
	var updated models.Transfer
	if err := json.Unmarshal(w.Body.Bytes(), &updated); err != nil {
		t.Fatal(err)
	}
	if updated.ID != 1 || updated.FromAccountID != 2 || updated.ToAccountID != 1 || updated.Amount != 100 || updated.Description != "" || !updated.CreatedAt.Equal(created.CreatedAt) || !updated.UpdatedAt.After(created.UpdatedAt) || updated != repo.records[1] {
		t.Fatalf("incorrect persisted update: %+v", updated)
	}
	request(t, mux, "DELETE", "/api/transfers/1", "", http.StatusNoContent)
	request(t, mux, "GET", "/api/transfers/1", "", http.StatusNotFound)
	request(t, mux, "PUT", "/api/transfers/1", validTransferBody, http.StatusNotFound)
	request(t, mux, "DELETE", "/api/transfers/1", "", http.StatusNotFound)
}

func TestTransferHandlerInvalidJSON(t *testing.T) {
	for _, body := range []string{
		"", "{", "null", "[]", "true", `"text"`,
		validTransferBody + ` {}`, validTransferBody + ` null`, validTransferBody + ` trailing`,
		strings.Replace(validTransferBody, `500000`, `1.5`, 1),
		strings.Replace(validTransferBody, `500000`, `9223372036854775808`, 1),
		strings.Replace(validTransferBody, `500000`, `"100"`, 1),
		strings.Replace(validTransferBody, `"fromAccountId":1`, `"fromAccountId":"1"`, 1),
		strings.Replace(validTransferBody, `"toAccountId":2`, `"toAccountId":2.5`, 1),
		strings.Replace(validTransferBody, `2026-10-06T00:00:00Z`, `invalid`, 1),
	} {
		for _, method := range []string{"POST", "PUT"} {
			t.Run(method+body, func(t *testing.T) {
				mux, _, _ := newTestTransferHandler()
				path := "/api/transfers"
				if method == "PUT" {
					path += "/1"
				}
				request(t, mux, method, path, body, http.StatusBadRequest)
			})
		}
	}
}

func TestTransferHandlerInvalidData(t *testing.T) {
	for _, body := range []string{
		`{}`,
		strings.Replace(validTransferBody, `"fromAccountId":1`, `"fromAccountId":0`, 1),
		strings.Replace(validTransferBody, `"toAccountId":2`, `"toAccountId":-1`, 1),
		strings.Replace(validTransferBody, `"toAccountId":2`, `"toAccountId":1`, 1),
		strings.Replace(validTransferBody, `500000`, `0`, 1),
		strings.Replace(validTransferBody, `500000`, `-1`, 1),
		strings.Replace(validTransferBody, `,"date":"2026-10-06T00:00:00Z"`, ``, 1),
		strings.Replace(validTransferBody, `"date":"2026-10-06T00:00:00Z"`, `"date":null`, 1),
		strings.Replace(validTransferBody, `"fromAccountId":1`, `"fromAccountId":99`, 1),
		strings.Replace(validTransferBody, `"toAccountId":2`, `"toAccountId":99`, 1),
	} {
		for _, method := range []string{"POST", "PUT"} {
			t.Run(method+body, func(t *testing.T) {
				mux, repo, _ := newTestTransferHandler()
				request(t, mux, "POST", "/api/transfers", validTransferBody, http.StatusCreated)
				original := repo.records[1]
				path := "/api/transfers"
				if method == "PUT" {
					path += "/1"
				}
				request(t, mux, method, path, body, http.StatusBadRequest)
				if len(repo.records) != 1 || repo.records[1] != original {
					t.Fatal("invalid input modified storage")
				}
			})
		}
	}
	for _, id := range []int64{1, 2} {
		for _, method := range []string{"POST", "PUT"} {
			mux, _, accounts := newTestTransferHandler()
			request(t, mux, "POST", "/api/transfers", validTransferBody, http.StatusCreated)
			a := accounts.accounts[id]
			a.Active = false
			accounts.accounts[id] = a
			path := "/api/transfers"
			if method == "PUT" {
				path += "/1"
			}
			request(t, mux, method, path, validTransferBody, http.StatusBadRequest)
		}
	}
}

func TestTransferHandlerIDsAndNotFound(t *testing.T) {
	for _, method := range []string{"GET", "PUT", "DELETE"} {
		for _, id := range []string{"abc", "0", "-1", "1.5", "9223372036854775808", "42"} {
			t.Run(method+id, func(t *testing.T) {
				mux, _, _ := newTestTransferHandler()
				status := http.StatusBadRequest
				if id == "42" {
					status = http.StatusNotFound
				}
				request(t, mux, method, "/api/transfers/"+id, validTransferBody, status)
			})
		}
	}
}

func TestTransferHandlerInternalErrors(t *testing.T) {
	for _, endpoint := range []struct{ method, path string }{
		{"GET", "/api/transfers"}, {"GET", "/api/transfers/1"},
		{"POST", "/api/transfers"}, {"PUT", "/api/transfers/1"}, {"DELETE", "/api/transfers/1"},
	} {
		mux, repo, _ := newTestTransferHandler()
		repo.err = errors.New("sensitive workbook path")
		w := request(t, mux, endpoint.method, endpoint.path, validTransferBody, http.StatusInternalServerError)
		if strings.TrimSpace(w.Body.String()) != `{"error":"internal server error"}` {
			t.Fatalf("internal details leaked: %s", w.Body.String())
		}
	}
	for _, method := range []string{"POST", "PUT"} {
		mux, _, accounts := newTestTransferHandler()
		request(t, mux, "POST", "/api/transfers", validTransferBody, http.StatusCreated)
		accounts.err = errors.New("sensitive lookup details")
		path := "/api/transfers"
		if method == "PUT" {
			path += "/1"
		}
		w := request(t, mux, method, path, validTransferBody, http.StatusInternalServerError)
		if strings.Contains(w.Body.String(), "sensitive") {
			t.Fatal("lookup details leaked")
		}
	}
}

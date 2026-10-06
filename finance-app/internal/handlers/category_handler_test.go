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

type categoryHandlerRepository struct {
	records map[int64]models.Category
	nextID  int64
	err     error
}

func (f *categoryHandlerRepository) Create(c models.Category) (models.Category, error) {
	if f.err != nil {
		return models.Category{}, f.err
	}
	f.nextID++
	c.ID = f.nextID
	f.records[c.ID] = c
	return c, nil
}
func (f *categoryHandlerRepository) GetAll() ([]models.Category, error) {
	if f.err != nil {
		return nil, f.err
	}
	var result []models.Category
	for _, c := range f.records {
		result = append(result, c)
	}
	return result, nil
}
func (f *categoryHandlerRepository) GetByID(id int64) (models.Category, error) {
	if f.err != nil {
		return models.Category{}, f.err
	}
	c, ok := f.records[id]
	if !ok {
		return models.Category{}, fmt.Errorf("ID %d: %w", id, repository.ErrNotFound)
	}
	return c, nil
}
func (f *categoryHandlerRepository) Update(c models.Category) error {
	if f.err != nil {
		return f.err
	}
	if _, ok := f.records[c.ID]; !ok {
		return repository.ErrNotFound
	}
	// Simulate the repository assigning a timestamp at persistence time.
	c.UpdatedAt = c.UpdatedAt.Add(time.Millisecond)
	f.records[c.ID] = c
	return nil
}
func (f *categoryHandlerRepository) Delete(id int64) error {
	if f.err != nil {
		return f.err
	}
	if _, ok := f.records[id]; !ok {
		return repository.ErrNotFound
	}
	delete(f.records, id)
	return nil
}

func newTestCategoryHandler() (*http.ServeMux, *categoryHandlerRepository) {
	repo := &categoryHandlerRepository{records: make(map[int64]models.Category)}
	mux := http.NewServeMux()
	NewCategoryHandler(services.NewCategoryService(repo)).RegisterRoutes(mux)
	return mux, repo
}

const validCategoryBody = `{"name":"Comida","type":"expense"}`

func responseCategory(t *testing.T, w *httptest.ResponseRecorder) models.Category {
	t.Helper()
	var c models.Category
	if err := json.Unmarshal(w.Body.Bytes(), &c); err != nil {
		t.Fatal(err)
	}
	return c
}

func TestCategoryHandlerCRUD(t *testing.T) {
	mux, repo := newTestCategoryHandler()
	w := request(t, mux, "GET", "/api/categories", "", http.StatusOK)
	if strings.TrimSpace(w.Body.String()) != "[]" {
		t.Fatalf("expected empty JSON array: %s", w.Body.String())
	}
	body := `{"id":99,"name":" Comida ","type":" expense ","createdAt":"client timestamp","updatedAt":"client timestamp"}`
	w = request(t, mux, "POST", "/api/categories", body, http.StatusCreated)
	created := responseCategory(t, w)
	if created.ID != 1 || created.Name != "Comida" || created.Type != "expense" || created.CreatedAt.IsZero() || !created.CreatedAt.Equal(created.UpdatedAt) {
		t.Fatalf("incorrect creation: %+v", created)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(w.Body.Bytes(), &fields); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"id", "name", "type", "createdAt", "updatedAt"} {
		if _, ok := fields[key]; !ok {
			t.Errorf("missing JSON field %s", key)
		}
	}
	if len(fields) != 5 {
		t.Fatalf("unexpected fields: %v", fields)
	}
	w = request(t, mux, "GET", "/api/categories/1", "", http.StatusOK)
	if got := responseCategory(t, w); got != created {
		t.Fatalf("GET by ID mismatch: %+v", got)
	}
	w = request(t, mux, "GET", "/api/categories", "", http.StatusOK)
	var all []models.Category
	if err := json.Unmarshal(w.Body.Bytes(), &all); err != nil {
		t.Fatal(err)
	}
	if len(all) != 1 || all[0] != created {
		t.Fatalf("GET all mismatch: %+v", all)
	}
	body = `{"id":999,"name":" Sueldo ","type":" income ","createdAt":"2000-01-01T00:00:00Z","updatedAt":"2000-01-01T00:00:00Z"}`
	w = request(t, mux, "PUT", "/api/categories/1", body, http.StatusOK)
	updated := responseCategory(t, w)
	if updated.ID != 1 || updated.Name != "Sueldo" || updated.Type != "income" || !updated.CreatedAt.Equal(created.CreatedAt) || !updated.UpdatedAt.After(created.UpdatedAt) {
		t.Fatalf("incorrect update: %+v", updated)
	}
	if repo.records[1].ID != 1 {
		t.Fatal("body ID used instead of URL ID")
	}
	request(t, mux, "DELETE", "/api/categories/1", "", http.StatusNoContent)
	request(t, mux, "GET", "/api/categories/1", "", http.StatusNotFound)
	request(t, mux, "PUT", "/api/categories/1", validCategoryBody, http.StatusNotFound)
	request(t, mux, "DELETE", "/api/categories/1", "", http.StatusNotFound)
}

func TestCategoryHandlerBadRequests(t *testing.T) {
	for _, body := range []string{
		"", " ", "{", "null", "[]", "true", `"text"`,
		validCategoryBody + ` {}`, validCategoryBody + ` null`, validCategoryBody + ` trailing`,
		`{"name":123,"type":"expense"}`, `{"name":"Comida","type":123}`,
		`{}`, `{"name":" ","type":"expense"}`, `{"name":"Comida"}`,
		`{"name":"Comida","type":" "}`, `{"name":"Comida","type":"other"}`,
	} {
		for _, method := range []string{"POST", "PUT"} {
			t.Run(method+body, func(t *testing.T) {
				mux, repo := newTestCategoryHandler()
				path := "/api/categories"
				if method == "PUT" {
					path += "/1"
				}
				request(t, mux, method, path, body, http.StatusBadRequest)
				if len(repo.records) != 0 {
					t.Fatal("invalid input created record")
				}
			})
		}
	}
}

func TestCategoryHandlerInvalidID(t *testing.T) {
	for _, id := range []string{"abc", "0", "-1", "1.5", "9223372036854775808"} {
		for _, method := range []string{"GET", "PUT", "DELETE"} {
			t.Run(method+id, func(t *testing.T) {
				mux, _ := newTestCategoryHandler()
				request(t, mux, method, "/api/categories/"+id, validCategoryBody, http.StatusBadRequest)
			})
		}
	}
}

func TestCategoryHandlerNotFound(t *testing.T) {
	for _, method := range []string{"GET", "PUT", "DELETE"} {
		t.Run(method, func(t *testing.T) {
			mux, _ := newTestCategoryHandler()
			request(t, mux, method, "/api/categories/42", validCategoryBody, http.StatusNotFound)
		})
	}
}

func TestCategoryHandlerInternalErrors(t *testing.T) {
	for _, endpoint := range []struct{ method, path string }{
		{"GET", "/api/categories"}, {"GET", "/api/categories/1"},
		{"POST", "/api/categories"}, {"PUT", "/api/categories/1"}, {"DELETE", "/api/categories/1"},
	} {
		t.Run(endpoint.method+endpoint.path, func(t *testing.T) {
			mux, repo := newTestCategoryHandler()
			repo.err = errors.New("sensitive workbook path")
			w := request(t, mux, endpoint.method, endpoint.path, validCategoryBody, http.StatusInternalServerError)
			if strings.TrimSpace(w.Body.String()) != `{"error":"internal server error"}` {
				t.Fatalf("internal details leaked: %s", w.Body.String())
			}
		})
	}
}

func TestAccountAndCategoryRoutesCoexist(t *testing.T) {
	mux, _ := newTestHandler()
	repo := &categoryHandlerRepository{records: make(map[int64]models.Category)}
	NewCategoryHandler(services.NewCategoryService(repo)).RegisterRoutes(mux)
	request(t, mux, "POST", "/api/accounts", validBody, http.StatusCreated)
	request(t, mux, "POST", "/api/categories", validCategoryBody, http.StatusCreated)
	request(t, mux, "GET", "/api/accounts/1", "", http.StatusOK)
	request(t, mux, "GET", "/api/categories/1", "", http.StatusOK)
}

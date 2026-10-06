package services

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"finance-app/internal/models"
	"finance-app/internal/repository"
)

type categoryFake struct {
	records  map[int64]models.Category
	nextID   int64
	calls    int
	failures map[string]error
}

func newCategoryFake() *categoryFake {
	return &categoryFake{records: make(map[int64]models.Category), failures: make(map[string]error)}
}
func (f *categoryFake) Create(c models.Category) (models.Category, error) {
	f.calls++
	if err := f.failures["create"]; err != nil {
		return models.Category{}, err
	}
	f.nextID++
	c.ID = f.nextID
	f.records[c.ID] = c
	return c, nil
}
func (f *categoryFake) GetAll() ([]models.Category, error) {
	f.calls++
	if err := f.failures["all"]; err != nil {
		return nil, err
	}
	var result []models.Category
	for _, c := range f.records {
		result = append(result, c)
	}
	return result, nil
}
func (f *categoryFake) GetByID(id int64) (models.Category, error) {
	f.calls++
	if err := f.failures["get"]; err != nil {
		return models.Category{}, err
	}
	c, ok := f.records[id]
	if !ok {
		return models.Category{}, fmt.Errorf("ID %d: %w", id, repository.ErrNotFound)
	}
	return c, nil
}
func (f *categoryFake) Update(c models.Category) error {
	f.calls++
	if err := f.failures["update"]; err != nil {
		return err
	}
	if _, ok := f.records[c.ID]; !ok {
		return repository.ErrNotFound
	}
	c.UpdatedAt = c.UpdatedAt.Add(time.Millisecond)
	f.records[c.ID] = c
	return nil
}
func (f *categoryFake) Delete(id int64) error {
	f.calls++
	if err := f.failures["delete"]; err != nil {
		return err
	}
	if _, ok := f.records[id]; !ok {
		return repository.ErrNotFound
	}
	delete(f.records, id)
	return nil
}

func TestCategoryServiceCRUD(t *testing.T) {
	fake := newCategoryFake()
	s := NewCategoryService(fake)
	all, err := s.GetAll()
	if err != nil || all == nil || len(all) != 0 {
		t.Fatalf("empty collection: %v %v", all, err)
	}
	before := time.Now().UTC()
	created, err := s.Create(models.Category{ID: 99, Name: " Comida \t", Type: " expense ", CreatedAt: before.Add(-time.Hour), UpdatedAt: before.Add(-time.Hour)})
	if err != nil {
		t.Fatal(err)
	}
	if created.ID != 1 || created.Name != "Comida" || created.Type != "expense" || created.CreatedAt.Before(before) || !created.CreatedAt.Equal(created.UpdatedAt) {
		t.Fatalf("incorrect creation: %+v", created)
	}
	got, err := s.GetByID(1)
	if err != nil || got != created {
		t.Fatalf("GetByID: %+v %v", got, err)
	}
	all, err = s.GetAll()
	if err != nil || len(all) != 1 || all[0] != created {
		t.Fatalf("GetAll: %+v %v", all, err)
	}
	updated, err := s.Update(1, models.Category{ID: 999, Name: " Sueldo ", Type: " income ", CreatedAt: before.Add(-time.Hour)})
	if err != nil {
		t.Fatal(err)
	}
	if updated.ID != 1 || updated.Name != "Sueldo" || updated.Type != "income" || !updated.CreatedAt.Equal(created.CreatedAt) || !updated.UpdatedAt.After(created.UpdatedAt) || updated != fake.records[1] {
		t.Fatalf("incorrect persisted update: %+v", updated)
	}
	if err := s.Delete(1); err != nil {
		t.Fatal(err)
	}
	_, getErr := s.GetByID(1)
	_, updateErr := s.Update(1, models.Category{Name: "Comida", Type: "expense"})
	for _, err := range []error{getErr, updateErr, s.Delete(1)} {
		if !errors.Is(err, repository.ErrNotFound) {
			t.Fatalf("ErrNotFound not propagated: %v", err)
		}
	}
}

func TestCategoryServiceValidation(t *testing.T) {
	cases := []struct {
		name     string
		category models.Category
		field    string
	}{
		{"empty name", models.Category{Type: "expense"}, "Name"},
		{"blank name", models.Category{Name: " \t", Type: "expense"}, "Name"},
		{"empty type", models.Category{Name: "Comida"}, "Type"},
		{"blank type", models.Category{Name: "Comida", Type: " "}, "Type"},
		{"invalid type", models.Category{Name: "Comida", Type: "other"}, "Type"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fake := newCategoryFake()
			s := NewCategoryService(fake)
			_, createErr := s.Create(tc.category)
			_, updateErr := s.Update(1, tc.category)
			for _, err := range []error{createErr, updateErr} {
				var validation *models.ValidationError
				if !errors.As(err, &validation) || validation.Field != tc.field {
					t.Fatalf("expected %s error: %v", tc.field, err)
				}
			}
			if fake.calls != 0 {
				t.Fatal("invalid data reached repository")
			}
		})
	}
	for _, id := range []int64{0, -1} {
		fake := newCategoryFake()
		s := NewCategoryService(fake)
		_, getErr := s.GetByID(id)
		_, updateErr := s.Update(id, models.Category{Name: "Comida", Type: "expense"})
		for _, err := range []error{getErr, updateErr, s.Delete(id)} {
			var validation *models.ValidationError
			if !errors.As(err, &validation) || validation.Field != "ID" {
				t.Fatalf("expected ID error: %v", err)
			}
		}
		if fake.calls != 0 {
			t.Fatal("invalid ID reached repository")
		}
	}
}

func TestCategoryServiceRepositoryErrors(t *testing.T) {
	for _, operation := range []string{"create", "all", "get", "update", "delete"} {
		t.Run(operation, func(t *testing.T) {
			fake := newCategoryFake()
			fake.records[1] = models.Category{ID: 1, Name: "Comida", Type: "expense"}
			failure := errors.New("repository unavailable")
			fake.failures[operation] = failure
			s := NewCategoryService(fake)
			var err error
			c := models.Category{Name: "Comida", Type: "expense"}
			switch operation {
			case "create":
				_, err = s.Create(c)
			case "all":
				_, err = s.GetAll()
			case "get":
				_, err = s.GetByID(1)
			case "update":
				_, err = s.Update(1, c)
			case "delete":
				err = s.Delete(1)
			}
			if !errors.Is(err, failure) {
				t.Fatalf("error not propagated: %v", err)
			}
		})
	}
}

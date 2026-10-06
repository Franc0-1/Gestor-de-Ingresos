package services

import (
	"errors"
	"fmt"
	"math"
	"testing"
	"time"

	"finance-app/internal/models"
	"finance-app/internal/repository"
)

type transactionFake struct {
	records  map[int64]models.Transaction
	nextID   int64
	writes   int
	failures map[string]error
}

func (f *transactionFake) Create(tx models.Transaction) (models.Transaction, error) {
	if err := f.failures["create"]; err != nil {
		return models.Transaction{}, err
	}
	f.nextID++
	tx.ID = f.nextID
	f.records[tx.ID] = tx
	f.writes++
	return tx, nil
}
func (f *transactionFake) GetAll() ([]models.Transaction, error) {
	if err := f.failures["all"]; err != nil {
		return nil, err
	}
	var txs []models.Transaction
	for _, tx := range f.records {
		txs = append(txs, tx)
	}
	return txs, nil
}
func (f *transactionFake) GetByID(id int64) (models.Transaction, error) {
	if err := f.failures["get"]; err != nil {
		return models.Transaction{}, err
	}
	tx, ok := f.records[id]
	if !ok {
		return models.Transaction{}, fmt.Errorf("ID %d: %w", id, repository.ErrNotFound)
	}
	return tx, nil
}
func (f *transactionFake) Update(tx models.Transaction) error {
	if err := f.failures["update"]; err != nil {
		return err
	}
	if _, ok := f.records[tx.ID]; !ok {
		return repository.ErrNotFound
	}
	tx.UpdatedAt = tx.UpdatedAt.Add(time.Millisecond)
	f.records[tx.ID] = tx
	f.writes++
	return nil
}
func (f *transactionFake) Delete(id int64) error {
	if err := f.failures["delete"]; err != nil {
		return err
	}
	if _, ok := f.records[id]; !ok {
		return repository.ErrNotFound
	}
	delete(f.records, id)
	f.writes++
	return nil
}

func transactionFixture() (*TransactionService, *transactionFake, *accountFake, *categoryFake) {
	txs := &transactionFake{records: make(map[int64]models.Transaction), failures: make(map[string]error)}
	accounts := newAccountFake()
	accounts.accounts[1] = models.Account{ID: 1, Name: "Cash", Type: "cash", Currency: "ARS", InitialBalance: 100, Active: false}
	categories := newCategoryFake()
	categories.records[1] = models.Category{ID: 1, Name: "Salary", Type: "income"}
	categories.records[2] = models.Category{ID: 2, Name: "Food", Type: "expense"}
	return NewTransactionService(txs, accounts, categories), txs, accounts, categories
}

func validTransaction(typ string) models.Transaction {
	categoryID := int64(1)
	if typ == "expense" {
		categoryID = 2
	}
	return models.Transaction{Type: typ, Amount: 1850000, Description: "Supermercado", AccountID: 1, CategoryID: categoryID, Date: time.Date(2026, 10, 6, 0, 0, 0, 123456789, time.FixedZone("ART", -3*60*60))}
}

func TestTransactionServiceCreate(t *testing.T) {
	for _, typ := range []string{"income", "expense"} {
		t.Run(typ, func(t *testing.T) {
			s, repo, accounts, _ := transactionFixture()
			tx := validTransaction(typ)
			tx.Type = " " + typ + " "
			tx.Description = " Supermercado \t"
			tx.Amount = math.MaxInt64
			tx.ID = 99
			tx.CreatedAt = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
			tx.UpdatedAt = tx.CreatedAt
			before := time.Now().UTC()
			created, err := s.Create(tx)
			if err != nil {
				t.Fatal(err)
			}
			if created.ID != 1 || created.Type != typ || created.Description != "Supermercado" || created.Amount != math.MaxInt64 || !created.Date.Equal(tx.Date) || created.Date.Location() != time.UTC || created.CreatedAt.Before(before) || !created.CreatedAt.Equal(created.UpdatedAt) || repo.records[1] != created {
				t.Fatalf("incorrect creation: %+v", created)
			}
			if accounts.accounts[1].InitialBalance != 100 {
				t.Fatal("account balance changed")
			}
		})
	}
}

func TestTransactionServiceValidation(t *testing.T) {
	cases := []struct {
		name, field string
		mutate      func(*models.Transaction)
	}{
		{"empty type", "Type", func(tx *models.Transaction) { tx.Type = " " }},
		{"invalid type", "Type", func(tx *models.Transaction) { tx.Type = "transfer" }},
		{"zero amount", "Amount", func(tx *models.Transaction) { tx.Amount = 0 }},
		{"negative amount", "Amount", func(tx *models.Transaction) { tx.Amount = -1 }},
		{"zero account", "AccountID", func(tx *models.Transaction) { tx.AccountID = 0 }},
		{"negative account", "AccountID", func(tx *models.Transaction) { tx.AccountID = -1 }},
		{"zero category", "CategoryID", func(tx *models.Transaction) { tx.CategoryID = 0 }},
		{"negative category", "CategoryID", func(tx *models.Transaction) { tx.CategoryID = -1 }},
		{"missing date", "Date", func(tx *models.Transaction) { tx.Date = time.Time{} }},
		{"missing account", "AccountID", func(tx *models.Transaction) { tx.AccountID = 99 }},
		{"missing category", "CategoryID", func(tx *models.Transaction) { tx.CategoryID = 99 }},
		{"expense with income category", "CategoryID", func(tx *models.Transaction) { tx.CategoryID = 1 }},
		{"income with expense category", "CategoryID", func(tx *models.Transaction) { tx.Type = "income" }},
	}
	for _, tc := range cases {
		for _, operation := range []string{"create", "update"} {
			t.Run(operation+"/"+tc.name, func(t *testing.T) {
				s, repo, _, _ := transactionFixture()
				original := validTransaction("expense")
				original.ID = 1
				repo.records[1] = original
				tx := validTransaction("expense")
				tc.mutate(&tx)
				var err error
				if operation == "create" {
					_, err = s.Create(tx)
				} else {
					_, err = s.Update(1, tx)
				}
				var validation *models.ValidationError
				if !errors.As(err, &validation) || validation.Field != tc.field || errors.Is(err, repository.ErrNotFound) {
					t.Fatalf("expected %s validation: %v", tc.field, err)
				}
				if repo.writes != 0 || repo.records[1] != original {
					t.Fatal("invalid transaction was persisted")
				}
			})
		}
	}
}

func TestTransactionServiceCRUD(t *testing.T) {
	s, repo, accounts, _ := transactionFixture()
	all, err := s.GetAll()
	if err != nil || all == nil || len(all) != 0 {
		t.Fatalf("empty array: %v %v", all, err)
	}
	tx := validTransaction("expense")
	tx.Description = ""
	created, err := s.Create(tx)
	if err != nil {
		t.Fatal(err)
	}
	got, err := s.GetByID(1)
	if err != nil || got != created {
		t.Fatalf("GetByID: %+v %v", got, err)
	}
	all, err = s.GetAll()
	if err != nil || len(all) != 1 || all[0] != created {
		t.Fatalf("GetAll: %+v %v", all, err)
	}
	tx = validTransaction("income")
	tx.ID = 999
	tx.CreatedAt = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	updated, err := s.Update(1, tx)
	if err != nil {
		t.Fatal(err)
	}
	if updated.ID != 1 || updated.Type != "income" || updated.CategoryID != 1 || !updated.CreatedAt.Equal(created.CreatedAt) || !updated.UpdatedAt.After(created.UpdatedAt) || updated != repo.records[1] {
		t.Fatalf("incorrect update: %+v", updated)
	}
	if err := s.Delete(1); err != nil {
		t.Fatal(err)
	}
	_, getErr := s.GetByID(1)
	_, updateErr := s.Update(1, validTransaction("expense"))
	for _, err := range []error{getErr, updateErr, s.Delete(1)} {
		if !errors.Is(err, repository.ErrNotFound) {
			t.Fatalf("ErrNotFound not propagated: %v", err)
		}
	}
	if accounts.accounts[1].InitialBalance != 100 {
		t.Fatal("account balance changed")
	}
}

func TestTransactionServiceInvalidID(t *testing.T) {
	for _, id := range []int64{0, -1} {
		s, repo, accounts, categories := transactionFixture()
		_, getErr := s.GetByID(id)
		_, updateErr := s.Update(id, validTransaction("expense"))
		for _, err := range []error{getErr, updateErr, s.Delete(id)} {
			var validation *models.ValidationError
			if !errors.As(err, &validation) || validation.Field != "ID" {
				t.Fatalf("expected ID error: %v", err)
			}
		}
		if repo.writes != 0 || accounts.calls != 0 || categories.calls != 0 {
			t.Fatal("invalid ID reached repositories")
		}
	}
}

func TestTransactionServiceRepositoryErrors(t *testing.T) {
	for _, operation := range []string{"create", "all", "get", "update", "delete"} {
		t.Run(operation, func(t *testing.T) {
			s, repo, _, _ := transactionFixture()
			repo.records[1] = validTransaction("expense")
			failure := errors.New("repository unavailable")
			repo.failures[operation] = failure
			var err error
			switch operation {
			case "create":
				_, err = s.Create(validTransaction("expense"))
			case "all":
				_, err = s.GetAll()
			case "get":
				_, err = s.GetByID(1)
			case "update":
				_, err = s.Update(1, validTransaction("expense"))
			case "delete":
				err = s.Delete(1)
			}
			if !errors.Is(err, failure) {
				t.Fatalf("error not propagated: %v", err)
			}
		})
	}
	for _, entity := range []string{"account", "category"} {
		for _, operation := range []string{"create", "update"} {
			t.Run(entity+operation, func(t *testing.T) {
				s, repo, accounts, categories := transactionFixture()
				repo.records[1] = validTransaction("expense")
				failure := errors.New("lookup unavailable")
				if entity == "account" {
					accounts.errors["get"] = failure
				} else {
					categories.failures["get"] = failure
				}
				var err error
				if operation == "create" {
					_, err = s.Create(validTransaction("expense"))
				} else {
					_, err = s.Update(1, validTransaction("expense"))
				}
				var validation *models.ValidationError
				if !errors.Is(err, failure) || errors.As(err, &validation) || repo.writes != 0 {
					t.Fatalf("lookup failure misclassified: %v", err)
				}
			})
		}
	}
}

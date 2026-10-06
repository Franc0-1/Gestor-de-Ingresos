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

type transferFake struct {
	records  map[int64]models.Transfer
	nextID   int64
	writes   int
	failures map[string]error
}

func (f *transferFake) Create(tr models.Transfer) (models.Transfer, error) {
	if err := f.failures["create"]; err != nil {
		return models.Transfer{}, err
	}
	f.nextID++
	tr.ID = f.nextID
	f.records[tr.ID] = tr
	f.writes++
	return tr, nil
}
func (f *transferFake) GetAll() ([]models.Transfer, error) {
	if err := f.failures["all"]; err != nil {
		return nil, err
	}
	var result []models.Transfer
	for _, tr := range f.records {
		result = append(result, tr)
	}
	return result, nil
}
func (f *transferFake) GetByID(id int64) (models.Transfer, error) {
	if err := f.failures["get"]; err != nil {
		return models.Transfer{}, err
	}
	tr, ok := f.records[id]
	if !ok {
		return models.Transfer{}, fmt.Errorf("ID %d: %w", id, repository.ErrNotFound)
	}
	return tr, nil
}
func (f *transferFake) Update(tr models.Transfer) error {
	if err := f.failures["update"]; err != nil {
		return err
	}
	if _, ok := f.records[tr.ID]; !ok {
		return repository.ErrNotFound
	}
	tr.UpdatedAt = tr.UpdatedAt.Add(time.Millisecond)
	f.records[tr.ID] = tr
	f.writes++
	return nil
}
func (f *transferFake) Delete(id int64) error {
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

func transferFixture() (*TransferService, *transferFake, *accountFake) {
	repo := &transferFake{records: make(map[int64]models.Transfer), failures: make(map[string]error)}
	accounts := newAccountFake()
	accounts.accounts[1] = models.Account{ID: 1, Name: "Cash", Type: "cash", Currency: "ARS", InitialBalance: 100, Active: true}
	accounts.accounts[2] = models.Account{ID: 2, Name: "Wallet", Type: "wallet", Currency: "ARS", InitialBalance: 200, Active: true}
	return NewTransferService(repo, accounts), repo, accounts
}

func validTransfer() models.Transfer {
	return models.Transfer{FromAccountID: 1, ToAccountID: 2, Amount: 500000, Description: "Mover dinero", Date: time.Date(2026, 10, 6, 0, 0, 0, 123456789, time.FixedZone("ART", -3*60*60))}
}

func TestTransferServiceCRUD(t *testing.T) {
	s, repo, accounts := transferFixture()
	source, destination := accounts.accounts[1], accounts.accounts[2]
	all, err := s.GetAll()
	if err != nil || all == nil || len(all) != 0 {
		t.Fatalf("empty array: %v %v", all, err)
	}
	tr := validTransfer()
	tr.ID = 99
	tr.Amount = math.MaxInt64
	tr.Description = " Mover dinero \t"
	tr.CreatedAt = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	tr.UpdatedAt = tr.CreatedAt
	before := time.Now().UTC()
	created, err := s.Create(tr)
	if err != nil {
		t.Fatal(err)
	}
	if created.ID != 1 || created.Amount != math.MaxInt64 || created.Description != "Mover dinero" || !created.Date.Equal(tr.Date) || created.Date.Location() != time.UTC || created.CreatedAt.Before(before) || !created.CreatedAt.Equal(created.UpdatedAt) {
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
	tr = validTransfer()
	tr.ID = 999
	tr.FromAccountID, tr.ToAccountID = 2, 1
	tr.Description = ""
	tr.CreatedAt = before.Add(-time.Hour)
	updated, err := s.Update(1, tr)
	if err != nil {
		t.Fatal(err)
	}
	if updated.ID != 1 || updated.FromAccountID != 2 || updated.ToAccountID != 1 || updated.Description != "" || !updated.CreatedAt.Equal(created.CreatedAt) || !updated.UpdatedAt.After(created.UpdatedAt) || updated != repo.records[1] {
		t.Fatalf("incorrect update: %+v", updated)
	}
	if err := s.Delete(1); err != nil {
		t.Fatal(err)
	}
	_, getErr := s.GetByID(1)
	_, updateErr := s.Update(1, validTransfer())
	for _, err := range []error{getErr, updateErr, s.Delete(1)} {
		if !errors.Is(err, repository.ErrNotFound) {
			t.Fatalf("ErrNotFound not propagated: %v", err)
		}
	}
	if accounts.accounts[1] != source || accounts.accounts[2] != destination {
		t.Fatal("transfer modified accounts")
	}
}

func TestTransferServiceValidation(t *testing.T) {
	cases := []struct {
		name, field string
		mutate      func(*models.Transfer, *accountFake)
	}{
		{"zero source", "FromAccountID", func(tr *models.Transfer, _ *accountFake) { tr.FromAccountID = 0 }},
		{"negative source", "FromAccountID", func(tr *models.Transfer, _ *accountFake) { tr.FromAccountID = -1 }},
		{"zero destination", "ToAccountID", func(tr *models.Transfer, _ *accountFake) { tr.ToAccountID = 0 }},
		{"negative destination", "ToAccountID", func(tr *models.Transfer, _ *accountFake) { tr.ToAccountID = -1 }},
		{"same account", "ToAccountID", func(tr *models.Transfer, _ *accountFake) { tr.ToAccountID = tr.FromAccountID }},
		{"zero amount", "Amount", func(tr *models.Transfer, _ *accountFake) { tr.Amount = 0 }},
		{"negative amount", "Amount", func(tr *models.Transfer, _ *accountFake) { tr.Amount = -1 }},
		{"missing date", "Date", func(tr *models.Transfer, _ *accountFake) { tr.Date = time.Time{} }},
		{"missing source", "FromAccountID", func(tr *models.Transfer, _ *accountFake) { tr.FromAccountID = 99 }},
		{"missing destination", "ToAccountID", func(tr *models.Transfer, _ *accountFake) { tr.ToAccountID = 99 }},
		{"inactive source", "FromAccountID", func(_ *models.Transfer, a *accountFake) { v := a.accounts[1]; v.Active = false; a.accounts[1] = v }},
		{"inactive destination", "ToAccountID", func(_ *models.Transfer, a *accountFake) { v := a.accounts[2]; v.Active = false; a.accounts[2] = v }},
	}
	for _, tc := range cases {
		for _, operation := range []string{"create", "update"} {
			t.Run(operation+tc.name, func(t *testing.T) {
				s, repo, accounts := transferFixture()
				original := validTransfer()
				original.ID = 1
				repo.records[1] = original
				tr := validTransfer()
				tc.mutate(&tr, accounts)
				var err error
				if operation == "create" {
					_, err = s.Create(tr)
				} else {
					_, err = s.Update(1, tr)
				}
				var validation *models.ValidationError
				if !errors.As(err, &validation) || validation.Field != tc.field || errors.Is(err, repository.ErrNotFound) {
					t.Fatalf("expected %s validation: %v", tc.field, err)
				}
				if repo.writes != 0 || repo.records[1] != original {
					t.Fatal("invalid transfer persisted")
				}
			})
		}
	}
	for _, id := range []int64{0, -1} {
		s, repo, accounts := transferFixture()
		_, getErr := s.GetByID(id)
		_, updateErr := s.Update(id, validTransfer())
		for _, err := range []error{getErr, updateErr, s.Delete(id)} {
			var validation *models.ValidationError
			if !errors.As(err, &validation) || validation.Field != "ID" {
				t.Fatalf("expected ID validation: %v", err)
			}
		}
		if repo.writes != 0 || accounts.calls != 0 {
			t.Fatal("invalid ID reached repositories")
		}
	}
}

func TestTransferServiceRepositoryErrors(t *testing.T) {
	for _, operation := range []string{"create", "all", "get", "update", "delete", "account create", "account update"} {
		t.Run(operation, func(t *testing.T) {
			s, repo, accounts := transferFixture()
			repo.records[1] = validTransfer()
			failure := errors.New("repository unavailable")
			if operation == "account create" || operation == "account update" {
				accounts.errors["get"] = failure
			} else {
				repo.failures[operation] = failure
			}
			var err error
			switch operation {
			case "create", "account create":
				_, err = s.Create(validTransfer())
			case "all":
				_, err = s.GetAll()
			case "get":
				_, err = s.GetByID(1)
			case "update", "account update":
				_, err = s.Update(1, validTransfer())
			case "delete":
				err = s.Delete(1)
			}
			var validation *models.ValidationError
			if !errors.Is(err, failure) || errors.As(err, &validation) {
				t.Fatalf("internal failure misclassified: %v", err)
			}
		})
	}
}

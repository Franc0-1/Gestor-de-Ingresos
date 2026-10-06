package services

import (
	"errors"
	"testing"
	"time"

	"finance-app/internal/models"
	"finance-app/internal/repository"
)

// Existing CRUD fakes retain their behavior when their contracts gain Query.
func (f *transactionFake) Query(q models.TransactionQuery) (models.TransactionPage, error) {
	items, err := f.GetAll()
	return models.TransactionPage{Items: items}, err
}
func (f *transferFake) Query(q models.TransferQuery) (models.TransferPage, error) {
	items, err := f.GetAll()
	return models.TransferPage{Items: items}, err
}

type transactionQueryFake struct {
	repository.TransactionRepository
	query  models.TransactionQuery
	calls  int
	result models.TransactionPage
	err    error
}

func (f *transactionQueryFake) Query(q models.TransactionQuery) (models.TransactionPage, error) {
	f.query = q
	f.calls++
	return f.result, f.err
}

type transferQueryFake struct {
	repository.TransferRepository
	query  models.TransferQuery
	calls  int
	result models.TransferPage
	err    error
}

func (f *transferQueryFake) Query(q models.TransferQuery) (models.TransferPage, error) {
	f.query = q
	f.calls++
	return f.result, f.err
}

func TestQueryServicesDefaultsAndDelegation(t *testing.T) {
	for _, tc := range []struct{ page, size, wantPage, wantSize int }{
		{0, 0, 1, 20}, {-2, -1, 1, 20}, {2, 500, 2, 100}, {3, 10, 3, 10},
	} {
		txRepo := &transactionQueryFake{result: models.TransactionPage{Items: []models.Transaction{{ID: 7}}, Pagination: models.Pagination{TotalItems: 1}}}
		trRepo := &transferQueryFake{result: models.TransferPage{Items: []models.Transfer{{ID: 8}}, Pagination: models.Pagination{TotalItems: 1}}}
		txService := NewTransactionService(txRepo, nil, nil)
		trService := NewTransferService(trRepo, nil)
		txPage, err := txService.Query(models.TransactionQuery{Type: " expense ", AccountID: 1, CategoryID: 2, Search: " super ", Page: tc.page, PageSize: tc.size})
		if err != nil || len(txPage.Items) != 1 || txPage.Items[0].ID != 7 || txPage.Pagination.TotalItems != 1 {
			t.Fatalf("transaction delegation: %+v %v", txPage, err)
		}
		trPage, err := trService.Query(models.TransferQuery{FromAccountID: 1, ToAccountID: 2, Search: " wallet ", Page: tc.page, PageSize: tc.size})
		if err != nil || len(trPage.Items) != 1 || trPage.Items[0].ID != 8 {
			t.Fatalf("transfer delegation: %+v %v", trPage, err)
		}
		if txRepo.query.Page != tc.wantPage || txRepo.query.PageSize != tc.wantSize || txRepo.query.SortBy != "date" || txRepo.query.SortOrder != "desc" || txRepo.query.Search != "super" || txRepo.query.Type != "expense" || txRepo.query.AccountID != 1 || txRepo.query.CategoryID != 2 {
			t.Fatalf("transaction normalization: %+v", txRepo.query)
		}
		if trRepo.query.Page != tc.wantPage || trRepo.query.PageSize != tc.wantSize || trRepo.query.SortBy != "date" || trRepo.query.SortOrder != "desc" || trRepo.query.Search != "wallet" || trRepo.query.FromAccountID != 1 || trRepo.query.ToAccountID != 2 {
			t.Fatalf("transfer normalization: %+v", trRepo.query)
		}
	}
	txRepo := &transactionQueryFake{}
	trRepo := &transferQueryFake{}
	if _, err := NewTransactionService(txRepo, nil, nil).Query(models.TransactionQuery{SortBy: "amount", SortOrder: " ASC "}); err != nil || txRepo.query.SortOrder != "asc" {
		t.Fatalf("sort normalization: %v", err)
	}
	if _, err := NewTransferService(trRepo, nil).Query(models.TransferQuery{SortBy: "id", SortOrder: " DESC "}); err != nil || trRepo.query.SortOrder != "desc" {
		t.Fatalf("sort normalization: %v", err)
	}
}

func TestQueryServicesValidation(t *testing.T) {
	from := time.Date(2026, 10, 7, 0, 0, 0, 0, time.UTC)
	to := from.Add(-time.Hour)
	for _, q := range []models.TransactionQuery{
		{Type: "other"}, {AccountID: -1}, {CategoryID: -1}, {DateFrom: from, DateTo: to}, {SortBy: "description"}, {SortOrder: "other"},
	} {
		repo := &transactionQueryFake{}
		_, err := NewTransactionService(repo, nil, nil).Query(q)
		var validation *models.ValidationError
		if !errors.As(err, &validation) || repo.calls != 0 {
			t.Fatalf("invalid transaction query reached repository: %+v %v", q, err)
		}
	}
	for _, q := range []models.TransferQuery{
		{FromAccountID: -1}, {ToAccountID: -1}, {DateFrom: from, DateTo: to}, {SortBy: "type"}, {SortBy: "description"}, {SortOrder: "other"},
	} {
		repo := &transferQueryFake{}
		_, err := NewTransferService(repo, nil).Query(q)
		var validation *models.ValidationError
		if !errors.As(err, &validation) || repo.calls != 0 {
			t.Fatalf("invalid transfer query reached repository: %+v %v", q, err)
		}
	}
}

func TestQueryServicesInternalErrors(t *testing.T) {
	failure := errors.New("repository unavailable")
	_, txErr := NewTransactionService(&transactionQueryFake{err: failure}, nil, nil).Query(models.TransactionQuery{})
	_, trErr := NewTransferService(&transferQueryFake{err: failure}, nil).Query(models.TransferQuery{})
	if !errors.Is(txErr, failure) || !errors.Is(trErr, failure) {
		t.Fatalf("internal error not propagated: %v / %v", txErr, trErr)
	}
}

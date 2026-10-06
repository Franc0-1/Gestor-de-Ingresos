package repository

import (
	"cmp"
	"slices"
	"strings"
	"time"

	"finance-app/internal/models"
)

type transactionTable struct {
	*excelTable[models.Transaction]
}
type transferTable struct{ *excelTable[models.Transfer] }

type sortKey struct {
	id, amount                 int64
	date, createdAt, updatedAt time.Time
	typ                        string
}

func compareKeys(a, b sortKey, by string) int {
	var result int
	switch by {
	case "id":
		result = cmp.Compare(a.id, b.id)
	case "amount":
		result = cmp.Compare(a.amount, b.amount)
	case "date":
		result = a.date.Compare(b.date)
	case "createdAt":
		result = a.createdAt.Compare(b.createdAt)
	case "updatedAt":
		result = a.updatedAt.Compare(b.updatedAt)
	case "type":
		result = strings.Compare(a.typ, b.typ)
	}
	// IDs make equal sort values deterministic across pages.
	if result == 0 {
		result = cmp.Compare(a.id, b.id)
	}
	return result
}

func inDateRange(date, from, to time.Time) bool {
	return (from.IsZero() || !date.Before(from)) && (to.IsZero() || !date.After(to))
}

func queryItems[T any](all []T, match func(T) bool, key func(T) sortKey, by, order string, page, size int) ([]T, models.Pagination) {
	items := make([]T, 0)
	for _, item := range all {
		if match(item) {
			items = append(items, item)
		}
	}
	slices.SortFunc(items, func(a, b T) int {
		result := compareKeys(key(a), key(b), by)
		if order == "desc" {
			return -result
		}
		return result
	})
	total := len(items)
	pages := total / size
	if total%size != 0 {
		pages++
	}
	metadata := models.Pagination{Page: page, PageSize: size, TotalItems: total, TotalPages: pages}
	// Check the page first so even MaxInt cannot overflow offset arithmetic.
	if page > pages {
		return []T{}, metadata
	}
	start := (page - 1) * size
	end := start + min(size, total-start)
	return items[start:end], metadata
}

func (r *transactionTable) Query(query models.TransactionQuery) (models.TransactionPage, error) {
	q, err := query.Normalize()
	if err != nil {
		return models.TransactionPage{}, err
	}
	all, err := r.GetAll()
	if err != nil {
		return models.TransactionPage{}, err
	}
	search := strings.ToLower(q.Search)
	items, pagination := queryItems(all, func(tx models.Transaction) bool {
		return (q.Type == "" || tx.Type == q.Type) &&
			(q.AccountID == 0 || tx.AccountID == q.AccountID) &&
			(q.CategoryID == 0 || tx.CategoryID == q.CategoryID) &&
			inDateRange(tx.Date, q.DateFrom, q.DateTo) &&
			strings.Contains(strings.ToLower(tx.Description), search)
	}, func(tx models.Transaction) sortKey {
		return sortKey{tx.ID, tx.Amount, tx.Date, tx.CreatedAt, tx.UpdatedAt, tx.Type}
	}, q.SortBy, q.SortOrder, q.Page, q.PageSize)
	return models.TransactionPage{Items: items, Pagination: pagination}, nil
}

func (r *transferTable) Query(query models.TransferQuery) (models.TransferPage, error) {
	q, err := query.Normalize()
	if err != nil {
		return models.TransferPage{}, err
	}
	all, err := r.GetAll()
	if err != nil {
		return models.TransferPage{}, err
	}
	search := strings.ToLower(q.Search)
	items, pagination := queryItems(all, func(tr models.Transfer) bool {
		return (q.FromAccountID == 0 || tr.FromAccountID == q.FromAccountID) &&
			(q.ToAccountID == 0 || tr.ToAccountID == q.ToAccountID) &&
			inDateRange(tr.Date, q.DateFrom, q.DateTo) &&
			strings.Contains(strings.ToLower(tr.Description), search)
	}, func(tr models.Transfer) sortKey {
		return sortKey{tr.ID, tr.Amount, tr.Date, tr.CreatedAt, tr.UpdatedAt, ""}
	}, q.SortBy, q.SortOrder, q.Page, q.PageSize)
	return models.TransferPage{Items: items, Pagination: pagination}, nil
}

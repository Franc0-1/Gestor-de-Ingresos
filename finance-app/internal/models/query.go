package models

import (
	"strings"
	"time"
)

type TransactionQuery struct {
	Type       string
	AccountID  int64
	CategoryID int64
	DateFrom   time.Time
	DateTo     time.Time
	Search     string
	SortBy     string
	SortOrder  string
	Page       int
	PageSize   int
}

type TransferQuery struct {
	FromAccountID int64
	ToAccountID   int64
	DateFrom      time.Time
	DateTo        time.Time
	Search        string
	SortBy        string
	SortOrder     string
	Page          int
	PageSize      int
}

type Pagination struct {
	Page       int `json:"page"`
	PageSize   int `json:"pageSize"`
	TotalItems int `json:"totalItems"`
	TotalPages int `json:"totalPages"`
}

type TransactionPage struct {
	Items      []Transaction `json:"items"`
	Pagination Pagination    `json:"pagination"`
}

type TransferPage struct {
	Items      []Transfer `json:"items"`
	Pagination Pagination `json:"pagination"`
}

func normalizePagination(page, size int) (int, int) {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}
	if size > 100 {
		size = 100
	}
	return page, size
}

func normalizeSort(by, order string, transaction bool) (string, string, error) {
	by = strings.TrimSpace(by)
	order = strings.ToLower(strings.TrimSpace(order))
	if by == "" {
		by = "date"
	}
	if order == "" {
		order = "desc"
	}
	switch by {
	case "id", "amount", "date", "createdAt", "updatedAt":
	case "type":
		if !transaction {
			return "", "", &ValidationError{Field: "SortBy", Reason: "invalid sort field"}
		}
	default:
		return "", "", &ValidationError{Field: "SortBy", Reason: "invalid sort field"}
	}
	if order != "asc" && order != "desc" {
		return "", "", &ValidationError{Field: "SortOrder", Reason: "must be asc or desc"}
	}
	return by, order, nil
}

func validateDateRange(from, to time.Time) error {
	if !from.IsZero() && !to.IsZero() && from.After(to) {
		return &ValidationError{Field: "DateFrom", Reason: "must not be after DateTo"}
	}
	return nil
}

// Normalize applies defaults and validates the query without HTTP concerns.
func (q TransactionQuery) Normalize() (TransactionQuery, error) {
	q.Type = strings.TrimSpace(q.Type)
	if q.Type != "" && q.Type != "income" && q.Type != "expense" {
		return q, &ValidationError{Field: "Type", Reason: "must be income or expense"}
	}
	if q.AccountID < 0 {
		return q, &ValidationError{Field: "AccountID", Reason: "must be positive when specified"}
	}
	if q.CategoryID < 0 {
		return q, &ValidationError{Field: "CategoryID", Reason: "must be positive when specified"}
	}
	if err := validateDateRange(q.DateFrom, q.DateTo); err != nil {
		return q, err
	}
	var err error
	q.SortBy, q.SortOrder, err = normalizeSort(q.SortBy, q.SortOrder, true)
	q.Page, q.PageSize = normalizePagination(q.Page, q.PageSize)
	q.Search = strings.TrimSpace(q.Search)
	return q, err
}

func (q TransferQuery) Normalize() (TransferQuery, error) {
	if q.FromAccountID < 0 {
		return q, &ValidationError{Field: "FromAccountID", Reason: "must be positive when specified"}
	}
	if q.ToAccountID < 0 {
		return q, &ValidationError{Field: "ToAccountID", Reason: "must be positive when specified"}
	}
	if err := validateDateRange(q.DateFrom, q.DateTo); err != nil {
		return q, err
	}
	var err error
	q.SortBy, q.SortOrder, err = normalizeSort(q.SortBy, q.SortOrder, false)
	q.Page, q.PageSize = normalizePagination(q.Page, q.PageSize)
	q.Search = strings.TrimSpace(q.Search)
	return q, err
}

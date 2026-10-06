package repository

import (
	"math"
	"slices"
	"testing"
	"time"

	"finance-app/internal/models"
)

func queryDay(day int) time.Time { return time.Date(2026, 10, day, 0, 0, 0, 0, time.UTC) }

func TestTransactionQueries(t *testing.T) {
	r, _ := newTestRepository(t)
	end := queryDay(32).Add(-time.Nanosecond)
	data := []models.Transaction{
		{Type: "expense", AccountID: 1, CategoryID: 2, Date: queryDay(1), Description: "SUPERMERCADO", Amount: math.MaxInt64},
		{Type: "income", AccountID: 1, CategoryID: 1, Date: queryDay(2), Description: "Sueldo", Amount: 200},
		{Type: "expense", AccountID: 2, CategoryID: 2, Date: end, Description: "Super barrio", Amount: 300},
		{Type: "expense", AccountID: 1, CategoryID: 3, Date: queryDay(32), Description: "Super extra", Amount: 400},
		{Type: "expense", AccountID: 1, CategoryID: 2, Date: queryDay(31), Description: "super final", Amount: 50},
	}
	for i, tx := range data {
		tx.CreatedAt = queryDay(i + 1)
		tx.UpdatedAt = queryDay(i + 2)
		if _, err := r.Transactions.Create(tx); err != nil {
			t.Fatal(err)
		}
	}
	cases := []struct {
		name         string
		q            models.TransactionQuery
		ids          []int64
		total, pages int
	}{
		{"default", models.TransactionQuery{}, []int64{4, 3, 5, 2, 1}, 5, 1},
		{"type", models.TransactionQuery{Type: "expense"}, []int64{4, 3, 5, 1}, 4, 1},
		{"account", models.TransactionQuery{AccountID: 1}, []int64{4, 5, 2, 1}, 4, 1},
		{"category", models.TransactionQuery{CategoryID: 2}, []int64{3, 5, 1}, 3, 1},
		{"dates inclusive", models.TransactionQuery{DateFrom: queryDay(1), DateTo: end}, []int64{3, 5, 2, 1}, 4, 1},
		{"from only", models.TransactionQuery{DateFrom: queryDay(31)}, []int64{4, 3, 5}, 3, 1},
		{"to only", models.TransactionQuery{DateTo: queryDay(2)}, []int64{2, 1}, 2, 1},
		{"search", models.TransactionQuery{Search: "SuPeR"}, []int64{4, 3, 5, 1}, 4, 1},
		{"combined before pagination", models.TransactionQuery{Type: "expense", AccountID: 1, CategoryID: 2, DateFrom: queryDay(1), DateTo: end, Search: "super", PageSize: 1}, []int64{5}, 2, 2},
		{"id asc", models.TransactionQuery{SortBy: "id", SortOrder: "asc"}, []int64{1, 2, 3, 4, 5}, 5, 1},
		{"id desc", models.TransactionQuery{SortBy: "id", SortOrder: "desc"}, []int64{5, 4, 3, 2, 1}, 5, 1},
		{"amount asc", models.TransactionQuery{SortBy: "amount", SortOrder: "asc"}, []int64{5, 2, 3, 4, 1}, 5, 1},
		{"amount desc", models.TransactionQuery{SortBy: "amount", SortOrder: "desc"}, []int64{1, 4, 3, 2, 5}, 5, 1},
		{"date asc", models.TransactionQuery{SortBy: "date", SortOrder: "asc"}, []int64{1, 2, 5, 3, 4}, 5, 1},
		{"created asc", models.TransactionQuery{SortBy: "createdAt", SortOrder: "asc"}, []int64{1, 2, 3, 4, 5}, 5, 1},
		{"updated desc", models.TransactionQuery{SortBy: "updatedAt", SortOrder: "desc"}, []int64{5, 4, 3, 2, 1}, 5, 1},
		{"type asc", models.TransactionQuery{SortBy: "type", SortOrder: "asc"}, []int64{1, 3, 4, 5, 2}, 5, 1},
		{"type desc", models.TransactionQuery{SortBy: "type", SortOrder: "desc"}, []int64{2, 5, 4, 3, 1}, 5, 1},
		{"page 2", models.TransactionQuery{Page: 2, PageSize: 2}, []int64{5, 2}, 5, 3},
		{"last page", models.TransactionQuery{Page: 3, PageSize: 2}, []int64{1}, 5, 3},
		{"huge page", models.TransactionQuery{Page: math.MaxInt, PageSize: 100}, []int64{}, 5, 1},
		{"max size", models.TransactionQuery{PageSize: 101}, []int64{4, 3, 5, 2, 1}, 5, 1},
		{"no results", models.TransactionQuery{Search: "absent"}, []int64{}, 0, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			page, err := r.Transactions.Query(tc.q)
			if err != nil {
				t.Fatal(err)
			}
			ids := make([]int64, 0, len(page.Items))
			for _, tx := range page.Items {
				ids = append(ids, tx.ID)
			}
			q, err := tc.q.Normalize()
			if err != nil {
				t.Fatal(err)
			}
			want := models.Pagination{Page: q.Page, PageSize: q.PageSize, TotalItems: tc.total, TotalPages: tc.pages}
			if page.Items == nil || !slices.Equal(ids, tc.ids) || page.Pagination != want {
				t.Fatalf("got IDs %v metadata %+v, want %v %+v", ids, page.Pagination, tc.ids, want)
			}
		})
	}
}

func TestTransferQueries(t *testing.T) {
	r, _ := newTestRepository(t)
	end := queryDay(32).Add(-time.Nanosecond)
	data := []models.Transfer{
		{FromAccountID: 1, ToAccountID: 2, Amount: 100, Date: queryDay(1), Description: "WALLET"},
		{FromAccountID: 2, ToAccountID: 1, Amount: math.MaxInt64, Date: queryDay(2), Description: "Bank"},
		{FromAccountID: 1, ToAccountID: 3, Amount: 300, Date: end, Description: "Wallet ahorro"},
		{FromAccountID: 3, ToAccountID: 2, Amount: 400, Date: queryDay(32), Description: "wallet extra"},
	}
	for i, tr := range data {
		tr.CreatedAt = queryDay(i + 1)
		tr.UpdatedAt = queryDay(i + 2)
		if _, err := r.Transfers.Create(tr); err != nil {
			t.Fatal(err)
		}
	}
	cases := []struct {
		name         string
		q            models.TransferQuery
		ids          []int64
		total, pages int
	}{
		{"default", models.TransferQuery{}, []int64{4, 3, 2, 1}, 4, 1},
		{"source", models.TransferQuery{FromAccountID: 1}, []int64{3, 1}, 2, 1},
		{"destination", models.TransferQuery{ToAccountID: 2}, []int64{4, 1}, 2, 1},
		{"dates", models.TransferQuery{DateFrom: queryDay(1), DateTo: end}, []int64{3, 2, 1}, 3, 1},
		{"search", models.TransferQuery{Search: "WaLlEt"}, []int64{4, 3, 1}, 3, 1},
		{"combined", models.TransferQuery{FromAccountID: 1, ToAccountID: 2, DateFrom: queryDay(1), DateTo: end, Search: "wallet", PageSize: 1}, []int64{1}, 1, 1},
		{"amount asc", models.TransferQuery{SortBy: "amount", SortOrder: "asc"}, []int64{1, 3, 4, 2}, 4, 1},
		{"amount desc", models.TransferQuery{SortBy: "amount", SortOrder: "desc"}, []int64{2, 4, 3, 1}, 4, 1},
		{"id asc", models.TransferQuery{SortBy: "id", SortOrder: "asc"}, []int64{1, 2, 3, 4}, 4, 1},
		{"created desc", models.TransferQuery{SortBy: "createdAt"}, []int64{4, 3, 2, 1}, 4, 1},
		{"updated asc", models.TransferQuery{SortBy: "updatedAt", SortOrder: "asc"}, []int64{1, 2, 3, 4}, 4, 1},
		{"date asc", models.TransferQuery{SortOrder: "asc"}, []int64{1, 2, 3, 4}, 4, 1},
		{"page 2", models.TransferQuery{Page: 2, PageSize: 3}, []int64{1}, 4, 2},
		{"out of range", models.TransferQuery{Page: math.MaxInt}, []int64{}, 4, 1},
		{"no results", models.TransferQuery{Search: "absent"}, []int64{}, 0, 0},
		{"size clamp", models.TransferQuery{PageSize: 999}, []int64{4, 3, 2, 1}, 4, 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			page, err := r.Transfers.Query(tc.q)
			if err != nil {
				t.Fatal(err)
			}
			ids := make([]int64, 0, len(page.Items))
			for _, tr := range page.Items {
				ids = append(ids, tr.ID)
			}
			q, err := tc.q.Normalize()
			if err != nil {
				t.Fatal(err)
			}
			want := models.Pagination{Page: q.Page, PageSize: q.PageSize, TotalItems: tc.total, TotalPages: tc.pages}
			if page.Items == nil || !slices.Equal(ids, tc.ids) || page.Pagination != want {
				t.Fatalf("got IDs %v metadata %+v, want %v %+v", ids, page.Pagination, tc.ids, want)
			}
		})
	}
}

func TestQuerySortTies(t *testing.T) {
	r, _ := newTestRepository(t)
	for i := 0; i < 3; i++ {
		if _, err := r.Transfers.Create(models.Transfer{FromAccountID: 1, ToAccountID: 2, Amount: 100, Date: queryDay(1)}); err != nil {
			t.Fatal(err)
		}
	}
	for _, tc := range []struct {
		order string
		want  int64
	}{{"asc", 2}, {"desc", 2}} {
		page, err := r.Transfers.Query(models.TransferQuery{SortBy: "amount", SortOrder: tc.order, Page: 2, PageSize: 1})
		if err != nil || len(page.Items) != 1 || page.Items[0].ID != tc.want || page.Pagination.TotalPages != 3 {
			t.Fatalf("unstable tie pagination: %+v %v", page, err)
		}
	}
}

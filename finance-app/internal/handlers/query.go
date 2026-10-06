package handlers

import (
	"net/http"
	"net/url"
	"strconv"
	"time"

	"finance-app/internal/models"
)

type queryParser struct {
	values url.Values
	err    error
}

func newQueryParser(r *http.Request) *queryParser {
	values, err := url.ParseQuery(r.URL.RawQuery)
	p := &queryParser{values: values}
	if err != nil {
		p.invalid("query")
	}
	return p
}

func (p *queryParser) invalid(field string) {
	if p.err == nil {
		p.err = &models.ValidationError{Field: field, Reason: "invalid query value"}
	}
}

func (p *queryParser) id(field string) int64 {
	if !p.values.Has(field) {
		return 0
	}
	value, err := strconv.ParseInt(p.values.Get(field), 10, 64)
	if err != nil || value <= 0 {
		p.invalid(field)
	}
	return value
}

func (p *queryParser) integer(field string) int {
	if !p.values.Has(field) {
		return 0
	}
	value, err := strconv.Atoi(p.values.Get(field))
	if err != nil {
		p.invalid(field)
	}
	return value
}

// Date-only filters describe complete UTC days, including the final nanosecond.
func (p *queryParser) date(field string, end bool) time.Time {
	if !p.values.Has(field) {
		return time.Time{}
	}
	value, err := time.Parse(time.DateOnly, p.values.Get(field))
	if err != nil {
		p.invalid(field)
		return time.Time{}
	}
	if end {
		value = value.AddDate(0, 0, 1).Add(-time.Nanosecond)
	}
	return value
}

func parseTransactionQuery(r *http.Request) (models.TransactionQuery, error) {
	p := newQueryParser(r)
	q := models.TransactionQuery{
		Type: p.values.Get("type"), AccountID: p.id("accountId"), CategoryID: p.id("categoryId"),
		DateFrom: p.date("from", false), DateTo: p.date("to", true), Search: p.values.Get("search"),
		SortBy: p.values.Get("sortBy"), SortOrder: p.values.Get("sortOrder"),
		Page: p.integer("page"), PageSize: p.integer("pageSize"),
	}
	return q, p.err
}

func parseTransferQuery(r *http.Request) (models.TransferQuery, error) {
	p := newQueryParser(r)
	q := models.TransferQuery{
		FromAccountID: p.id("fromAccountId"), ToAccountID: p.id("toAccountId"),
		DateFrom: p.date("from", false), DateTo: p.date("to", true), Search: p.values.Get("search"),
		SortBy: p.values.Get("sortBy"), SortOrder: p.values.Get("sortOrder"),
		Page: p.integer("page"), PageSize: p.integer("pageSize"),
	}
	return q, p.err
}

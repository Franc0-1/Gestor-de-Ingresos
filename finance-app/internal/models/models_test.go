package models

import (
	"errors"
	"testing"
	"time"
)

func TestValidation(t *testing.T) {
	a := Account{Name: "Cuenta", Type: "cash", Currency: "ARS"}
	c := Category{Name: "Categoría", Type: "income"}
	x := Transaction{Type: "expense", Amount: 1, CategoryID: 1, AccountID: 1, Date: time.Now()}
	for _, typ := range []string{"cash", "bank", "wallet", "savings", "other"} {
		a.Type = typ
		if err := a.Validate(); err != nil {
			t.Error(err)
		}
	}
	for _, typ := range []string{"income", "expense"} {
		c.Type = typ
		x.Type = typ
		if err := c.Validate(); err != nil {
			t.Error(err)
		}
		if err := x.Validate(); err != nil {
			t.Error(err)
		}
	}
	a.Type = "cash"
	cases := []struct {
		name  string
		value interface{ Validate() error }
		field string
	}{
		{"account name", Account{Type: a.Type, Currency: a.Currency}, "Name"},
		{"account whitespace", Account{Name: " \t", Type: a.Type, Currency: a.Currency}, "Name"},
		{"account type", Account{Name: a.Name, Type: "invalid", Currency: a.Currency}, "Type"},
		{"account currency", Account{Name: a.Name, Type: a.Type, Currency: " "}, "Currency"},
		{"category name", Category{Type: c.Type}, "Name"},
		{"category type", Category{Name: c.Name, Type: "other"}, "Type"},
		{"transaction type", Transaction{Type: "transfer", Amount: 1, CategoryID: 1, AccountID: 1, Date: x.Date}, "Type"},
		{"transaction amount zero", Transaction{Type: x.Type, CategoryID: 1, AccountID: 1, Date: x.Date}, "Amount"},
		{"transaction amount negative", Transaction{Type: x.Type, Amount: -1, CategoryID: 1, AccountID: 1, Date: x.Date}, "Amount"},
		{"transaction category", Transaction{Type: x.Type, Amount: 1, AccountID: 1, Date: x.Date}, "CategoryID"},
		{"transaction account", Transaction{Type: x.Type, Amount: 1, CategoryID: 1, Date: x.Date}, "AccountID"},
		{"transaction date", Transaction{Type: x.Type, Amount: 1, CategoryID: 1, AccountID: 1}, "Date"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var err *ValidationError
			if !errors.As(tc.value.Validate(), &err) || err.Field != tc.field {
				t.Fatalf("expected %s validation", tc.field)
			}
		})
	}
}

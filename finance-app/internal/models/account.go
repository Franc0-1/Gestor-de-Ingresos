package models

import (
	"strings"
	"time"
)

// ValidationError identifies an invalid model field.
type ValidationError struct {
	Field  string
	Reason string
}

func (e *ValidationError) Error() string { return e.Field + ": " + e.Reason }

type Account struct {
	ID             int64     `json:"id"`
	Name           string    `json:"name"`
	Type           string    `json:"type"`
	Currency       string    `json:"currency"`
	InitialBalance int64     `json:"initialBalance"`
	Active         bool      `json:"active"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

func (a Account) Validate() error {
	if strings.TrimSpace(a.Name) == "" {
		return &ValidationError{"Name", "required"}
	}
	switch a.Type {
	case "cash", "bank", "wallet", "savings", "other":
	default:
		return &ValidationError{"Type", "invalid account type"}
	}
	if strings.TrimSpace(a.Currency) == "" {
		return &ValidationError{"Currency", "required"}
	}
	return nil
}

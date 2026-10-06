package models

import "time"

type Transaction struct {
	ID          int64     `json:"id"`
	Type        string    `json:"type"`
	Amount      int64     `json:"amount"`
	Description string    `json:"description"`
	CategoryID  int64     `json:"categoryId"`
	AccountID   int64     `json:"accountId"`
	Date        time.Time `json:"date"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func (t Transaction) Validate() error {
	if t.Type != "income" && t.Type != "expense" {
		return &ValidationError{"Type", "must be income or expense"}
	}
	if t.Amount <= 0 {
		return &ValidationError{"Amount", "must be positive"}
	}
	if t.CategoryID <= 0 {
		return &ValidationError{"CategoryID", "must be positive"}
	}
	if t.AccountID <= 0 {
		return &ValidationError{"AccountID", "must be positive"}
	}
	if t.Date.IsZero() {
		return &ValidationError{"Date", "required"}
	}
	return nil
}

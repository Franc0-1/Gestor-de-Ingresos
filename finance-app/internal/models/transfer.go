package models

import "time"

type Transfer struct {
	ID            int64     `json:"id"`
	FromAccountID int64     `json:"fromAccountId"`
	ToAccountID   int64     `json:"toAccountId"`
	Amount        int64     `json:"amount"`
	Description   string    `json:"description"`
	Date          time.Time `json:"date"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

func (t Transfer) Validate() error {
	if t.FromAccountID <= 0 {
		return &ValidationError{Field: "FromAccountID", Reason: "must be positive"}
	}
	if t.ToAccountID <= 0 {
		return &ValidationError{Field: "ToAccountID", Reason: "must be positive"}
	}
	if t.FromAccountID == t.ToAccountID {
		return &ValidationError{Field: "ToAccountID", Reason: "must differ from source account"}
	}
	if t.Amount <= 0 {
		return &ValidationError{Field: "Amount", Reason: "must be positive"}
	}
	if t.Date.IsZero() {
		return &ValidationError{Field: "Date", Reason: "required"}
	}
	return nil
}

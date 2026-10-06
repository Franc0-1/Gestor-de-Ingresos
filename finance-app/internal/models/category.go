package models

import (
	"strings"
	"time"
)

type Category struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (c Category) Validate() error {
	if strings.TrimSpace(c.Name) == "" {
		return &ValidationError{"Name", "required"}
	}
	if c.Type != "income" && c.Type != "expense" {
		return &ValidationError{"Type", "must be income or expense"}
	}
	return nil
}

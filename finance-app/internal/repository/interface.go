package repository

import (
	"errors"
	"finance-app/internal/models"
)

var (
	ErrNotFound        = errors.New("record not found")
	ErrInvalidWorkbook = errors.New("invalid workbook")
	ErrIDExhausted     = errors.New("ID sequence exhausted")
)

type AccountRepository interface {
	Create(models.Account) (models.Account, error)
	GetAll() ([]models.Account, error)
	GetByID(int64) (models.Account, error)
	Update(models.Account) error
	Delete(int64) error
}

type CategoryRepository interface {
	Create(models.Category) (models.Category, error)
	GetAll() ([]models.Category, error)
	GetByID(int64) (models.Category, error)
	Update(models.Category) error
	Delete(int64) error
}

type TransactionRepository interface {
	Query(models.TransactionQuery) (models.TransactionPage, error)
	Create(models.Transaction) (models.Transaction, error)
	GetAll() ([]models.Transaction, error)
	GetByID(int64) (models.Transaction, error)
	Update(models.Transaction) error
	Delete(int64) error
}

type TransferRepository interface {
	Query(models.TransferQuery) (models.TransferPage, error)
	Create(models.Transfer) (models.Transfer, error)
	GetAll() ([]models.Transfer, error)
	GetByID(int64) (models.Transfer, error)
	Update(models.Transfer) error
	Delete(int64) error
}

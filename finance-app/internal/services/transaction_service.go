package services

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"finance-app/internal/models"
	"finance-app/internal/repository"
)

type TransactionService struct {
	repo       repository.TransactionRepository
	accounts   repository.AccountRepository
	categories repository.CategoryRepository
}

func (s *TransactionService) Query(query models.TransactionQuery) (models.TransactionPage, error) {
	query, err := query.Normalize()
	if err != nil {
		return models.TransactionPage{}, err
	}
	return s.repo.Query(query)
}

func NewTransactionService(repo repository.TransactionRepository, accounts repository.AccountRepository, categories repository.CategoryRepository) *TransactionService {
	return &TransactionService{repo: repo, accounts: accounts, categories: categories}
}

func normalizeTransaction(tx models.Transaction) models.Transaction {
	tx.Type = strings.TrimSpace(tx.Type)
	tx.Description = strings.TrimSpace(tx.Description)
	tx.Date = tx.Date.UTC()
	return tx
}

func (s *TransactionService) validateReferences(tx models.Transaction) error {
	if _, err := s.accounts.GetByID(tx.AccountID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return &models.ValidationError{Field: "AccountID", Reason: "account does not exist"}
		}
		return fmt.Errorf("lookup transaction account: %w", err)
	}
	category, err := s.categories.GetByID(tx.CategoryID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return &models.ValidationError{Field: "CategoryID", Reason: "category does not exist"}
		}
		return fmt.Errorf("lookup transaction category: %w", err)
	}
	if category.Type != tx.Type {
		return &models.ValidationError{Field: "CategoryID", Reason: "category type must match transaction type"}
	}
	return nil
}

func (s *TransactionService) Create(tx models.Transaction) (models.Transaction, error) {
	tx = normalizeTransaction(tx)
	if err := tx.Validate(); err != nil {
		return models.Transaction{}, err
	}
	if err := s.validateReferences(tx); err != nil {
		return models.Transaction{}, err
	}
	tx.ID = 0
	now := time.Now().UTC()
	tx.CreatedAt = now
	tx.UpdatedAt = now
	return s.repo.Create(tx)
}

func (s *TransactionService) GetAll() ([]models.Transaction, error) {
	txs, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}
	if txs == nil {
		txs = []models.Transaction{}
	}
	return txs, nil
}

func (s *TransactionService) GetByID(id int64) (models.Transaction, error) {
	if err := validateID(id); err != nil {
		return models.Transaction{}, err
	}
	return s.repo.GetByID(id)
}

func (s *TransactionService) Update(id int64, tx models.Transaction) (models.Transaction, error) {
	if err := validateID(id); err != nil {
		return models.Transaction{}, err
	}
	tx = normalizeTransaction(tx)
	if err := tx.Validate(); err != nil {
		return models.Transaction{}, err
	}
	existing, err := s.repo.GetByID(id)
	if err != nil {
		return models.Transaction{}, err
	}
	if err := s.validateReferences(tx); err != nil {
		return models.Transaction{}, err
	}
	tx.ID = id
	tx.CreatedAt = existing.CreatedAt
	tx.UpdatedAt = time.Now().UTC()
	if err := s.repo.Update(tx); err != nil {
		return models.Transaction{}, err
	}
	// Return the timestamp actually persisted by the Excel repository.
	return s.repo.GetByID(id)
}

func (s *TransactionService) Delete(id int64) error {
	if err := validateID(id); err != nil {
		return err
	}
	return s.repo.Delete(id)
}

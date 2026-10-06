package services

import (
	"strings"
	"time"

	"finance-app/internal/models"
	"finance-app/internal/repository"
)

type AccountService struct {
	repo repository.AccountRepository
}

func NewAccountService(repo repository.AccountRepository) *AccountService {
	return &AccountService{repo: repo}
}

func normalizeAccount(account models.Account) models.Account {
	account.Name = strings.TrimSpace(account.Name)
	account.Type = strings.ToLower(strings.TrimSpace(account.Type))
	account.Currency = strings.ToUpper(strings.TrimSpace(account.Currency))
	return account
}

func validateID(id int64) error {
	if id <= 0 {
		return &models.ValidationError{Field: "ID", Reason: "must be positive"}
	}
	return nil
}

func (s *AccountService) Create(account models.Account) (models.Account, error) {
	account = normalizeAccount(account)
	if err := account.Validate(); err != nil {
		return models.Account{}, err
	}
	account.ID = 0
	now := time.Now().UTC()
	account.CreatedAt = now
	account.UpdatedAt = now
	return s.repo.Create(account)
}

func (s *AccountService) GetAll() ([]models.Account, error) {
	accounts, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}
	if accounts == nil {
		accounts = []models.Account{}
	}
	return accounts, nil
}

func (s *AccountService) GetByID(id int64) (models.Account, error) {
	if err := validateID(id); err != nil {
		return models.Account{}, err
	}
	return s.repo.GetByID(id)
}

func (s *AccountService) Update(id int64, account models.Account) (models.Account, error) {
	if err := validateID(id); err != nil {
		return models.Account{}, err
	}
	account = normalizeAccount(account)
	if err := account.Validate(); err != nil {
		return models.Account{}, err
	}
	existing, err := s.repo.GetByID(id)
	if err != nil {
		return models.Account{}, err
	}
	account.ID = id
	account.CreatedAt = existing.CreatedAt
	account.UpdatedAt = time.Now().UTC()
	if err := s.repo.Update(account); err != nil {
		return models.Account{}, err
	}
	// The Excel repository assigns its own UpdatedAt; return the persisted value.
	return s.repo.GetByID(id)
}

func (s *AccountService) Delete(id int64) error {
	if err := validateID(id); err != nil {
		return err
	}
	return s.repo.Delete(id)
}

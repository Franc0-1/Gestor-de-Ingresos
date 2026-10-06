package services

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"finance-app/internal/models"
	"finance-app/internal/repository"
)

type TransferService struct {
	repo     repository.TransferRepository
	accounts repository.AccountRepository
}

func (s *TransferService) Query(query models.TransferQuery) (models.TransferPage, error) {
	query, err := query.Normalize()
	if err != nil {
		return models.TransferPage{}, err
	}
	return s.repo.Query(query)
}

func NewTransferService(repo repository.TransferRepository, accounts repository.AccountRepository) *TransferService {
	return &TransferService{repo: repo, accounts: accounts}
}

func normalizeTransfer(transfer models.Transfer) models.Transfer {
	transfer.Description = strings.TrimSpace(transfer.Description)
	transfer.Date = transfer.Date.UTC()
	return transfer
}

func (s *TransferService) validateAccounts(transfer models.Transfer) error {
	for _, ref := range []struct {
		field string
		id    int64
	}{
		{"FromAccountID", transfer.FromAccountID}, {"ToAccountID", transfer.ToAccountID},
	} {
		account, err := s.accounts.GetByID(ref.id)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return &models.ValidationError{Field: ref.field, Reason: "account does not exist"}
			}
			return fmt.Errorf("lookup transfer %s: %w", ref.field, err)
		}
		if !account.Active {
			return &models.ValidationError{Field: ref.field, Reason: "account must be active"}
		}
	}
	return nil
}

func (s *TransferService) Create(transfer models.Transfer) (models.Transfer, error) {
	transfer = normalizeTransfer(transfer)
	if err := transfer.Validate(); err != nil {
		return models.Transfer{}, err
	}
	if err := s.validateAccounts(transfer); err != nil {
		return models.Transfer{}, err
	}
	transfer.ID = 0
	now := time.Now().UTC()
	transfer.CreatedAt = now
	transfer.UpdatedAt = now
	return s.repo.Create(transfer)
}

func (s *TransferService) GetAll() ([]models.Transfer, error) {
	transfers, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}
	if transfers == nil {
		transfers = []models.Transfer{}
	}
	return transfers, nil
}

func (s *TransferService) GetByID(id int64) (models.Transfer, error) {
	if err := validateID(id); err != nil {
		return models.Transfer{}, err
	}
	return s.repo.GetByID(id)
}

func (s *TransferService) Update(id int64, transfer models.Transfer) (models.Transfer, error) {
	if err := validateID(id); err != nil {
		return models.Transfer{}, err
	}
	transfer = normalizeTransfer(transfer)
	if err := transfer.Validate(); err != nil {
		return models.Transfer{}, err
	}
	existing, err := s.repo.GetByID(id)
	if err != nil {
		return models.Transfer{}, err
	}
	if err := s.validateAccounts(transfer); err != nil {
		return models.Transfer{}, err
	}
	transfer.ID = id
	transfer.CreatedAt = existing.CreatedAt
	transfer.UpdatedAt = time.Now().UTC()
	if err := s.repo.Update(transfer); err != nil {
		return models.Transfer{}, err
	}
	return s.repo.GetByID(id)
}

func (s *TransferService) Delete(id int64) error {
	if err := validateID(id); err != nil {
		return err
	}
	return s.repo.Delete(id)
}

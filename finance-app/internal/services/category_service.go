package services

import (
	"strings"
	"time"

	"finance-app/internal/models"
	"finance-app/internal/repository"
)

type CategoryService struct {
	repo repository.CategoryRepository
}

func NewCategoryService(repo repository.CategoryRepository) *CategoryService {
	return &CategoryService{repo: repo}
}

func normalizeCategory(category models.Category) models.Category {
	category.Name = strings.TrimSpace(category.Name)
	category.Type = strings.TrimSpace(category.Type)
	return category
}

func (s *CategoryService) Create(category models.Category) (models.Category, error) {
	category = normalizeCategory(category)
	if err := category.Validate(); err != nil {
		return models.Category{}, err
	}
	category.ID = 0
	now := time.Now().UTC()
	category.CreatedAt = now
	category.UpdatedAt = now
	return s.repo.Create(category)
}

func (s *CategoryService) GetAll() ([]models.Category, error) {
	categories, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}
	if categories == nil {
		categories = []models.Category{}
	}
	return categories, nil
}

func (s *CategoryService) GetByID(id int64) (models.Category, error) {
	if err := validateID(id); err != nil {
		return models.Category{}, err
	}
	return s.repo.GetByID(id)
}

func (s *CategoryService) Update(id int64, category models.Category) (models.Category, error) {
	if err := validateID(id); err != nil {
		return models.Category{}, err
	}
	category = normalizeCategory(category)
	if err := category.Validate(); err != nil {
		return models.Category{}, err
	}
	existing, err := s.repo.GetByID(id)
	if err != nil {
		return models.Category{}, err
	}
	category.ID = id
	category.CreatedAt = existing.CreatedAt
	category.UpdatedAt = time.Now().UTC()
	if err := s.repo.Update(category); err != nil {
		return models.Category{}, err
	}
	// Return the timestamp actually persisted by the Excel repository.
	return s.repo.GetByID(id)
}

func (s *CategoryService) Delete(id int64) error {
	if err := validateID(id); err != nil {
		return err
	}
	return s.repo.Delete(id)
}

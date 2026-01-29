package storage

import (
	"tunorth-brms-backend/internal/core/domain"
	"tunorth-brms-backend/internal/core/ports"

	"gorm.io/gorm"
)

type resourceRepository struct {
	db *gorm.DB
}

func NewResourceRepository(db *gorm.DB) ports.ResourceRepository {
	return &resourceRepository{db: db}
}

func (r *resourceRepository) Create(resource *domain.Resource) error {
	return r.db.Create(resource).Error
}

func (r *resourceRepository) GetAll() ([]domain.Resource, error) {
	var resources []domain.Resource
	err := r.db.Find(&resources).Error
	return resources, err
}

func (r *resourceRepository) GetByID(id uint) (*domain.Resource, error) {
	var resource domain.Resource
	err := r.db.First(&resource, id).Error
	return &resource, err
}

func (r *resourceRepository) Update(resource *domain.Resource) error {
	return r.db.Save(resource).Error
}

func (r *resourceRepository) Delete(id uint) error {
	return r.db.Delete(&domain.Resource{}, id).Error
}
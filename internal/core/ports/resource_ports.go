package ports

import "tunorth-brms-backend/internal/core/domain"

type ResourceRepository interface {
	Create(resource *domain.Resource) error
	GetAll() ([]domain.Resource, error)
	GetByID(id uint) (*domain.Resource, error)
	Update(resource *domain.Resource) error
	Delete(id uint) error
}

type ResourceService interface {
	CreateResource(resource *domain.Resource) error
	GetAllResources() ([]domain.Resource, error)
	GetResourceByID(id uint) (*domain.Resource, error)
	UpdateResource(id uint, resource *domain.Resource) error
	DeleteResource(id uint) error
}
package services

import (
	"errors"
	"fmt"
	"tunorth-brms-backend/internal/core/domain"
	"tunorth-brms-backend/internal/core/ports"
)

type resourceService struct {
	repo       ports.ResourceRepository
	logService ports.LogService
}

func NewResourceService(repo ports.ResourceRepository, logService ports.LogService) ports.ResourceService {
	return &resourceService{repo: repo, logService: logService}
}

func (s *resourceService) CreateResource(resource *domain.Resource) error {
	if resource.ResourceName == "" {
		return errors.New("resource name is required")
	}
	if err := s.repo.Create(resource); err != nil {
		return err
	}
	go s.logService.LogAction(0, "CREATE_RESOURCE", fmt.Sprintf("Created resource: %s", resource.ResourceName), "", "")
	return nil
}

func (s *resourceService) GetAllResources() ([]domain.Resource, error) {
	return s.repo.GetAll()
}

func (s *resourceService) GetResourceByID(id uint) (*domain.Resource, error) {
	return s.repo.GetByID(id)
}

func (s *resourceService) UpdateResource(id uint, input *domain.Resource) error {
	existing, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	// Update fields
	existing.ResourceName = input.ResourceName
	existing.Type = input.Type
	
	if err := s.repo.Update(existing); err != nil {
		return err
	}
	go s.logService.LogAction(0, "UPDATE_RESOURCE", fmt.Sprintf("Updated resource ID: %d", id), "", "")
	return nil
}

func (s *resourceService) DeleteResource(id uint) error {
	if err := s.repo.Delete(id); err != nil {
		return err
	}
	go s.logService.LogAction(0, "DELETE_RESOURCE", fmt.Sprintf("Deleted resource ID: %d", id), "", "")
	return nil
}
package services

import (
	"tunorth-brms-backend/internal/core/domain"
	"tunorth-brms-backend/internal/core/ports"
)

type logService struct {
	repo ports.LogRepository
}

func NewLogService(repo ports.LogRepository) ports.LogService {
	return &logService{repo: repo}
}

func (s *logService) LogAction(userID uint, action, description, ip, userAgent string) error {
	log := &domain.Log{
		UserID:      userID,
		Action:      action,
		Description: description,
		IPAddress:   ip,
		UserAgent:   userAgent,
	}
	return s.repo.Create(log)
}

func (s *logService) GetLogs(limit int) ([]domain.Log, error) {
	// Default limit if not provided
	if limit <= 0 {
		limit = 100
	}
	return s.repo.GetAll(limit)
}

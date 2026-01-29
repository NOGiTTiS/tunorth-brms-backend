package ports

import "tunorth-brms-backend/internal/core/domain"

type LogRepository interface {
	Create(log *domain.Log) error
	GetAll(limit int) ([]domain.Log, error)
	// Add filter method later if needed
}

type LogService interface {
	LogAction(userID uint, action, description, ip, userAgent string) error
	GetLogs(limit int) ([]domain.Log, error)
}

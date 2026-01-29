package storage

import (
	"tunorth-brms-backend/internal/core/domain"
	"tunorth-brms-backend/internal/core/ports"

	"gorm.io/gorm"
)

type logRepository struct {
	db *gorm.DB
}

func NewLogRepository(db *gorm.DB) ports.LogRepository {
	return &logRepository{db: db}
}

func (r *logRepository) Create(log *domain.Log) error {
	return r.db.Create(log).Error
}

func (r *logRepository) GetAll(limit int) ([]domain.Log, error) {
	var logs []domain.Log
	// Preload User to show who did the action
	err := r.db.Preload("User").Order("created_at desc").Limit(limit).Find(&logs).Error
	return logs, err
}

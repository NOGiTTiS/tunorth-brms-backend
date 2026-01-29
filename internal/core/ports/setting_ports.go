package ports

import "tunorth-brms-backend/internal/core/domain"

type SettingRepository interface {
	GetAll() ([]domain.Setting, error)
	GetByName(name string) (*domain.Setting, error)
	Update(setting *domain.Setting) error
	UpdateBatch(settings []domain.Setting) error
}

type SettingService interface {
	GetAllSettings() (map[string]interface{}, []domain.Setting, error)
	UpdateSettings(updates []domain.Setting, actorID uint) error
	GetSettingValue(key string) string
	InitializeDefaults() error
}

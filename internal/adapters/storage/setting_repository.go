package storage

import (
	"tunorth-brms-backend/internal/core/domain"
	"tunorth-brms-backend/internal/core/ports"

	"gorm.io/gorm"
)

type settingRepository struct {
	db *gorm.DB
}

func NewSettingRepository(db *gorm.DB) ports.SettingRepository {
	return &settingRepository{db: db}
}

func (r *settingRepository) GetAll() ([]domain.Setting, error) {
	var settings []domain.Setting
	err := r.db.Find(&settings).Error
	return settings, err
}

func (r *settingRepository) GetByName(name string) (*domain.Setting, error) {
	var setting domain.Setting
	err := r.db.First(&setting, "setting_name = ?", name).Error
	return &setting, err
}

func (r *settingRepository) Update(setting *domain.Setting) error {
	return r.db.Save(setting).Error
}

func (r *settingRepository) UpdateBatch(settings []domain.Setting) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, s := range settings {
			// Update only value
			if err := tx.Model(&domain.Setting{}).Where("setting_name = ?", s.SettingName).Update("setting_value", s.SettingValue).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

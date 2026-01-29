package domain

import (
	"time"
	"gorm.io/gorm"
)

// Room struct แทนตาราง rooms
type Room struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	RoomName    string         `gorm:"not null" json:"room_name"`
	Description string         `json:"description"`
	Capacity    int            `json:"capacity"`
	ImagePath   string         `json:"image_path"`
	Color       string         `gorm:"type:varchar(7)" json:"color"` // Hex Code เช่น #FF5733
	Status      string         `gorm:"default:'active'" json:"status"` // active, maintenance
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}
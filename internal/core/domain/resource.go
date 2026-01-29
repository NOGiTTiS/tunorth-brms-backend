package domain

import "time"

// Resource struct แทนตาราง resources
type Resource struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	ResourceName string    `gorm:"not null" json:"resource_name"`
	Type         string    `json:"type"` // equipment, catering
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
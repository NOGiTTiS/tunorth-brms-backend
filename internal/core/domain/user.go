package domain

import (
	"time"
	"gorm.io/gorm"
)

// User struct แทนตาราง users
type User struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Username     string         `gorm:"unique;not null" json:"username"`
	Password     string         `gorm:"not null" json:"-"` // json:"-" เพื่อไม่ให้ส่ง password กลับไปหน้าเว็บ
	FullName     string         `gorm:"not null" json:"full_name"`
	Department   string         `json:"department"`
	Role         string         `gorm:"type:varchar(20);default:'user'" json:"role"` // admin, approver, user
	Email        string         `gorm:"unique" json:"email"`
	Tel          string         `json:"tel"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"` // Soft Delete (ลบแบบกู้คืนได้)
}
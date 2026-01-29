package storage

import (
	"time"
	"tunorth-brms-backend/internal/core/domain"
	"tunorth-brms-backend/internal/core/ports"

	"gorm.io/gorm"
)

type bookingRepository struct {
	db *gorm.DB
}

func NewBookingRepository(db *gorm.DB) ports.BookingRepository {
	return &bookingRepository{db: db}
}

func (r *bookingRepository) Create(booking *domain.Booking) error {
	return r.db.Create(booking).Error
}

func (r *bookingRepository) GetAll() ([]domain.Booking, error) {
	var bookings []domain.Booking
	// Preload Room และ User เพื่อเอาไปโชว์
	err := r.db.Preload("Room").Preload("User").Find(&bookings).Error
	return bookings, err
}

func (r *bookingRepository) GetByID(id uint) (*domain.Booking, error) {
	var booking domain.Booking
	err := r.db.Preload("Room").Preload("User").First(&booking, id).Error
	return &booking, err
}

// GetByDateRange: ดึงข้อมูลเฉพาะช่วงวันที่กำหนด (เช่น ดึงทีละเดือน)
func (r *bookingRepository) GetByDateRange(start, end time.Time) ([]domain.Booking, error) {
	var bookings []domain.Booking
	err := r.db.Preload("Room").Preload("User").
		Where("start_time >= ? AND start_time <= ?", start, end).
		Find(&bookings).Error
	return bookings, err
}

// CountOverlapping: นับจำนวนการจองที่เวลาทับซ้อนกัน
// Logic: (StartA < EndB) AND (EndA > StartB)
func (r *bookingRepository) CountOverlapping(roomID uint, start, end time.Time) (int64, error) {
	var count int64
	err := r.db.Model(&domain.Booking{}).
		Where("room_id = ? AND status != 'cancelled' AND start_time < ? AND end_time > ?", roomID, end, start).
		Count(&count).Error
	return count, err
}

func (r *bookingRepository) CountOverlappingExcludingID(roomID uint, start, end time.Time, excludeID uint) (int64, error) {
	var count int64
	err := r.db.Model(&domain.Booking{}).
		Where("room_id = ? AND status != 'cancelled' AND start_time < ? AND end_time > ? AND id != ?", roomID, end, start, excludeID).
		Count(&count).Error
	return count, err
}

func (r *bookingRepository) Update(booking *domain.Booking) error {
	return r.db.Save(booking).Error
}

func (r *bookingRepository) Delete(id uint) error {
	return r.db.Delete(&domain.Booking{}, id).Error
}
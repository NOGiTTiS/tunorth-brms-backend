package ports

import (
	"time"
	"tunorth-brms-backend/internal/core/domain"
)

type BookingRepository interface {
	Create(booking *domain.Booking) error
	GetAll() ([]domain.Booking, error)
	GetByID(id uint) (*domain.Booking, error)
	// ดึงเฉพาะช่วงเวลา (สำหรับปฏิทิน)
	GetByDateRange(start, end time.Time) ([]domain.Booking, error)
	// เช็คว่าห้องนี้ เวลานี้ มีใครจองหรือยัง (เพื่อป้องกันจองซ้ำ)
	CountOverlapping(roomID uint, start, end time.Time) (int64, error)
	// เช็คซ้ำแต่นับข้าม ID ตัวเอง (สำหรับ Update)
	CountOverlappingExcludingID(roomID uint, start, end time.Time, excludeID uint) (int64, error)
	Update(booking *domain.Booking) error
	Delete(id uint) error
}

type BookingService interface {
	CreateBooking(booking *domain.Booking) error
	GetAllBookings() ([]domain.Booking, error)
	GetBookingsByRange(start, end string) ([]domain.Booking, error) // รับ string แล้วแปลงเป็น time ใน service
	GetBookingByID(id uint) (*domain.Booking, error)
	UpdateBookingStatus(id uint, status string, approverID uint) error
	UpdateBooking(id uint, booking *domain.Booking, actorID uint) error
	// DeleteBooking(id uint) error -> เปลี่ยนเป็น รับ actorID ด้วย
	DeleteBooking(id uint, actorID uint) error
}
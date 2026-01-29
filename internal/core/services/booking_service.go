package services

import (
	"errors"
	"fmt"
	"strconv"
	"time"
	"tunorth-brms-backend/internal/core/domain"
	"tunorth-brms-backend/internal/core/ports"
)

type bookingService struct {
	repo       ports.BookingRepository
	roomRepo   ports.RoomRepository // Add RoomRepo
	settings   ports.SettingService
	userRepo   ports.UserRepository
	notifier   ports.NotificationService
	logService ports.LogService
}

func NewBookingService(repo ports.BookingRepository, roomRepo ports.RoomRepository, settings ports.SettingService, userRepo ports.UserRepository, notifier ports.NotificationService, logService ports.LogService) ports.BookingService {
	return &bookingService{
		repo:       repo,
		roomRepo:   roomRepo,
		settings:   settings,
		userRepo:   userRepo,
		notifier:   notifier,
		logService: logService,
	}
}

func (s *bookingService) CreateBooking(booking *domain.Booking) error {
	// 1. Validation พื้นฐาน
	if booking.RoomID == 0 {
		return errors.New("room_id is required")
	}
	if booking.StartTime.After(booking.EndTime) || booking.StartTime.Equal(booking.EndTime) {
		return errors.New("start time must be before end time")
	}

	// 1.2 Check Room Status
	room, err := s.roomRepo.GetByID(booking.RoomID)
	if err != nil {
		return errors.New("room not found")
	}
	if room.Status != "active" {
		return fmt.Errorf("room is not available (Status: %s)", room.Status)
	}

	// 1.5. Advance Booking Check
	// Get User Role
	user, err := s.userRepo.GetByID(booking.UserID)
	isAdmin := false
	if err == nil && user.Role == "admin" {
		isAdmin = true
	}

	if !isAdmin {
		advanceDaysStr := s.settings.GetSettingValue("advance_booking_days")
		advanceDays, _ := strconv.Atoi(advanceDaysStr)

		if advanceDays > 0 {
			now := time.Now()
			// Use server local time logic
			// Create dates with 00:00:00 time part for comparison
			bookingDate := time.Date(booking.StartTime.Year(), booking.StartTime.Month(), booking.StartTime.Day(), 0, 0, 0, 0, now.Location())
			minDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, advanceDays)

			if bookingDate.Before(minDate) {
				return fmt.Errorf("must book at least %d days in advance", advanceDays)
			}
		}
	}

	// 2. Weekend Check
	allowWeekend := s.settings.GetSettingValue("allow_weekend")
	// ถ้าเป็น "false" หรือปิดอยู่ (default) ห้ามจองเสาร์-อาทิตย์
	if allowWeekend == "false" {
		if booking.StartTime.Weekday() == time.Saturday || booking.StartTime.Weekday() == time.Sunday ||
			booking.EndTime.Weekday() == time.Saturday || booking.EndTime.Weekday() == time.Sunday {
			return errors.New("booking is not allowed on weekends")
		}
	}

	// 3. Conflict Check (ป้องกันจองซ้ำ)
	count, err := s.repo.CountOverlapping(booking.RoomID, booking.StartTime, booking.EndTime)
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("room is not available at this time")
	}

	// 3. กำหนดสถานะเริ่มต้น
	defaultStatus := s.settings.GetSettingValue("default_booking_status")
	if defaultStatus == "" {
		defaultStatus = "pending"
	}
	booking.Status = defaultStatus

	// 4. บันทึก
	if err := s.repo.Create(booking); err != nil {
		return err
	}

	// 5. Notify Admin
	// เรียกแบบ Async (go func) เพื่อไม่ให้ User ต้องรอ
	go s.notifier.NotifyAdminNewBooking(booking)

	// 6. Log Activity
	go s.logService.LogAction(booking.UserID, "CREATE_BOOKING", fmt.Sprintf("จองห้อง ID: %d วันที่: %s", booking.RoomID, booking.StartTime.Format("02/01/2006")), "", "")

	return nil
}

func (s *bookingService) GetAllBookings() ([]domain.Booking, error) {
	return s.repo.GetAll()
}

// GetBookingsByRange รับ string มาแปลงเป็น time ก่อนส่งให้ repo
func (s *bookingService) GetBookingsByRange(startStr, endStr string) ([]domain.Booking, error) {
	// FullCalendar ส่งมา format ISO8601 (2026-01-01T00:00:00Z)
	layout := time.RFC3339 
	
	start, err := time.Parse(layout, startStr)
	if err != nil {
		// ถ้า parse ไม่ได้ ลอง format ง่ายๆ (เผื่อส่งมาแค่ YYYY-MM-DD)
		start, err = time.Parse("2006-01-02", startStr)
		if err != nil {
			return nil, errors.New("invalid start date format")
		}
	}

	end, err := time.Parse(layout, endStr)
	if err != nil {
		end, err = time.Parse("2006-01-02", endStr)
		if err != nil {
			return nil, errors.New("invalid end date format")
		}
	}

	return s.repo.GetByDateRange(start, end)
}

func (s *bookingService) GetBookingByID(id uint) (*domain.Booking, error) {
	return s.repo.GetByID(id)
}

func (s *bookingService) UpdateBookingStatus(id uint, status string, approverID uint) error {
	// 1. หา Booking เดิมมาก่อน
	booking, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	// 2. อัปเดตสถานะ
	booking.Status = status
	booking.ApproverID = &approverID // บันทึกว่าใครเป็นคนกดอนุมัติ
	
	// 3. บันทึก
	if err := s.repo.Update(booking); err != nil {
		return err
	}
	
	// 4. Notify User
	go s.notifier.NotifyUserStatusChange(booking)
	
	// 5. Log
	action := "UPDATE"
	if status == "approved" {
		action = "APPROVE"
	} else if status == "rejected" {
		action = "REJECT"
	} else if status == "cancelled" {
		action = "CANCEL"
	}
	go s.logService.LogAction(approverID, action, fmt.Sprintf("%s รายการจอง ID: %d", status, booking.ID), "", "")

	return nil
}

func (s *bookingService) UpdateBooking(id uint, updatedBooking *domain.Booking, actorID uint) error {
    existing, err := s.repo.GetByID(id)
    if err != nil {
        return err
    }
    
    // Update fields
    existing.Subject = updatedBooking.Subject
    existing.RoomID = updatedBooking.RoomID
    existing.StartTime = updatedBooking.StartTime
    existing.EndTime = updatedBooking.EndTime
    existing.Note = updatedBooking.Note
    // Add other fields if necessary
    
    // Validate Time again?
	if existing.StartTime.After(existing.EndTime) || existing.StartTime.Equal(existing.EndTime) {
		return errors.New("start time must be before end time")
	}

    // Check conflict? If room or time changed.
    // For simplicity, let's assume conflict check is skipped or basic re-check
    // s.repo.CountOverlapping(...)
    count, err := s.repo.CountOverlappingExcludingID(existing.RoomID, existing.StartTime, existing.EndTime, id)
    if err != nil {
        return err
    }
    if count > 0 {
        return errors.New("room is not available at this time")
    }
    
    // IMPORTANT: Clear associations to prevent Gorm from trying to update/create them
    // or causing issues with the foreign key update
    existing.Room = domain.Room{}
    existing.User = domain.User{}
    existing.Approver = nil

    if err := s.repo.Update(existing); err != nil {
		return err
	}

	// Log
	go s.logService.LogAction(actorID, "UPDATE_BOOKING", fmt.Sprintf("Updated booking ID: %d", id), "", "")

	return nil
}

func (s *bookingService) DeleteBooking(id uint, actorID uint) error {
	booking, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	// Determine if actor is admin
	// Note: Ideally we pass user role or full user obj, but for now lets fetch user.
	// Optimization: Pass isAdmin flag or fetch actor role.
	actor, err := s.userRepo.GetByID(actorID)
	if err != nil {
		return errors.New("unauthorized")
	}

	// Check permission: Owner OR Admin
	if booking.UserID != actorID && actor.Role != "admin" {
		return errors.New("you do not have permission to delete this booking")
	}

    if err := s.repo.Delete(id); err != nil {
		return err
	}

	// Log
	go s.logService.LogAction(actorID, "DELETE_BOOKING", fmt.Sprintf("Deleted booking ID: %d", id), "", "")

	return nil
}
package storage

import (
	"fmt"
	"time"
	"tunorth-brms-backend/internal/core/domain"
	"tunorth-brms-backend/internal/core/ports"

	"gorm.io/gorm"
)

type reportRepository struct {
	db *gorm.DB
}

func NewReportRepository(db *gorm.DB) ports.ReportRepository {
	return &reportRepository{db: db}
}

func (r *reportRepository) GetStats(start, end time.Time, status string) (*ports.ReportStats, error) {
	stats := &ports.ReportStats{}

	// 1. Total Booking (All statuses in date range)
	// Use Model to utilize soft delete automatically
	baseQuery := r.db.Model(&domain.Booking{}).Where("start_time >= ? AND start_time <= ?", start, end)
	
	if err := baseQuery.Count(&stats.TotalBookings).Error; err != nil {
		fmt.Printf("Error counting total: %v\n", err)
		return nil, err
	}

	// 2. Status Counts
	// 2. Status Counts (Use distinct queries to avoid contamination)
	// Approved
	r.db.Model(&domain.Booking{}).Where("start_time >= ? AND start_time <= ?", start, end).
		Where("LOWER(status) = ?", "approved").Count(&stats.ApprovedCount)

	// Pending
	r.db.Model(&domain.Booking{}).Where("start_time >= ? AND start_time <= ?", start, end).
		Where("LOWER(status) = ?", "pending").Count(&stats.PendingCount)

	// Rejected
	r.db.Model(&domain.Booking{}).Where("start_time >= ? AND start_time <= ?", start, end).
		Where("LOWER(status) = ?", "rejected").Count(&stats.RejectedCount)

	// Create a query based on Status Filter for Charts
	chartQuery := r.db.Model(&domain.Booking{}).Where("start_time >= ? AND start_time <= ?", start, end)
	if status != "" && status != "all" {
		// Use lowercase for comparison to be safe
		chartQuery = chartQuery.Where("LOWER(status) = LOWER(?)", status)
	}

	// 3. Room Usage (Pie Chart)
	// Use Scan instead of Rows for simplicity and robustness
	err := chartQuery.Session(&gorm.Session{}).
		Joins("JOIN rooms ON rooms.id = bookings.room_id").
		Select("rooms.room_name, rooms.color, count(*) as count").
		Group("rooms.room_name, rooms.color").
		Order("count DESC").
		Scan(&stats.RoomUsage).Error

	if err != nil {
		fmt.Printf("Error fetching room usage: %v\n", err)
	}

	// 4. Daily Trends (Line Chart)
	// Try a safer date format logic that works for both MySQL/PG and SQLite (usually)
	err = chartQuery.Session(&gorm.Session{}).
		Select("DATE(start_time) as date, count(*) as count").
		Group("DATE(start_time)").
		Order("date ASC").
		Scan(&stats.DailyTrends).Error

	if err != nil {
		fmt.Printf("Error fetching daily trends: %v\n", err)
	}

	return stats, nil
}

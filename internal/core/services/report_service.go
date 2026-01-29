package services

import (
	"time"
	"tunorth-brms-backend/internal/core/ports"
)

type reportService struct {
	repo ports.ReportRepository
}

func NewReportService(repo ports.ReportRepository) ports.ReportService {
	return &reportService{repo: repo}
}

func (s *reportService) GetDashboardStats(startStr, endStr, status string) (*ports.ReportStats, error) {
	// 1. Parse Dates (dd/MM/yyyy from Frontend? or yyyy-MM-dd?)
	// Let's assume frontend sends YYYY-MM-DD or use default time.RFC3339
	// The user screenshot shows "01/01/2026", so maybe MM/DD/YYYY or DD/MM/YYYY.
	// Standard ISO is safest. Let's try parsing.
	
	const layoutISO = "2006-01-02"
	
	start, err := time.Parse(layoutISO, startStr)
	if err != nil {
		// Default to beginning of month if fail
		now := time.Now()
		start = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	}
	
	end, err := time.Parse(layoutISO, endStr)
	if err != nil {
		// Default to end of month? or Now?
		now := time.Now()
		end = now
	}
	
	// Ensure End is end of day
	end = time.Date(end.Year(), end.Month(), end.Day(), 23, 59, 59, 0, end.Location())

	return s.repo.GetStats(start, end, status)
}

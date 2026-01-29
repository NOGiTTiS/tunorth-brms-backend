package ports

import "time"

type ReportStats struct {
	TotalBookings int64        `json:"total_bookings"`
	ApprovedCount int64        `json:"approved_count"`
	PendingCount  int64        `json:"pending_count"`
	RejectedCount int64        `json:"rejected_count"`
	RoomUsage     []RoomUsage  `json:"room_usage"`
	DailyTrends   []DailyTrend `json:"daily_trends"`
}

type RoomUsage struct {
	RoomName string `json:"room_name"`
	Color    string `json:"color"`
	Count    int64  `json:"count"`
}

type DailyTrend struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

type ReportRepository interface {
	GetStats(start, end time.Time, status string) (*ReportStats, error)
}

type ReportService interface {
	GetDashboardStats(startStr, endStr, status string) (*ReportStats, error)
}

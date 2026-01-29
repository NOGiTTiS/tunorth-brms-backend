package http

import (
	"tunorth-brms-backend/internal/core/ports"

	"github.com/gofiber/fiber/v2"
)

type ReportHandler struct {
	service ports.ReportService
}

func NewReportHandler(service ports.ReportService) *ReportHandler {
	return &ReportHandler{service: service}
}

// GET /api/reports/dashboard
func (h *ReportHandler) GetDashboardStats(c *fiber.Ctx) error {
	startDate := c.Query("start") // YYYY-MM-DD
	endDate := c.Query("end")     // YYYY-MM-DD
	status := c.Query("status")   // all, approved, pending, rejected

	stats, err := h.service.GetDashboardStats(startDate, endDate, status)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(stats)
}

package http

import (
	"strconv"
	"tunorth-brms-backend/internal/core/domain"
	"tunorth-brms-backend/internal/core/ports"

	"github.com/gofiber/fiber/v2"
)

type RoomHandler struct {
	service ports.RoomService
}

// NewRoomHandler สร้าง Handler และรับ Service เข้ามา
func NewRoomHandler(service ports.RoomService) *RoomHandler {
	return &RoomHandler{service: service}
}

// CreateRoom: [POST] /api/rooms
func (h *RoomHandler) CreateRoom(c *fiber.Ctx) error {
	var room domain.Room
	// 1. แปลง JSON จาก Body เป็น Struct
	if err := c.BodyParser(&room); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// 2. เรียกใช้ Service
	if err := h.service.CreateRoom(&room); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// 3. ส่งข้อมูลห้องที่สร้างเสร็จกลับไป
	return c.Status(fiber.StatusCreated).JSON(room)
}

// GetAllRooms: [GET] /api/rooms
func (h *RoomHandler) GetAllRooms(c *fiber.Ctx) error {
	rooms, err := h.service.GetAllRooms()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

    // Filter by status if provided (e.g. ?status=active)
    statusFilter := c.Query("status")
    if statusFilter != "" {
        var filtered []domain.Room
        for _, r := range rooms {
            if r.Status == statusFilter {
                filtered = append(filtered, r)
            }
        }
        return c.JSON(filtered)
    }

	return c.JSON(rooms)
}

// GetRoom: [GET] /api/rooms/:id
func (h *RoomHandler) GetRoom(c *fiber.Ctx) error {
	// ดึง ID จาก URL Parameter
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
	}

	room, err := h.service.GetRoomByID(uint(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Room not found"})
	}
	return c.JSON(room)
}

// UpdateRoom: [PUT] /api/rooms/:id
func (h *RoomHandler) UpdateRoom(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
	}

	var room domain.Room
	if err := c.BodyParser(&room); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if err := h.service.UpdateRoom(uint(id), &room); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Room updated successfully"})
}

// DeleteRoom: [DELETE] /api/rooms/:id
func (h *RoomHandler) DeleteRoom(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
	}

	if err := h.service.DeleteRoom(uint(id)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Room deleted successfully"})
}
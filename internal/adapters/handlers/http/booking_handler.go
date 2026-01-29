package http

import (
	"fmt"
	"strconv"
	"time"
	"tunorth-brms-backend/internal/adapters/storage"
	"tunorth-brms-backend/internal/core/domain"
	"tunorth-brms-backend/internal/core/ports"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

type BookingHandler struct {
	service        ports.BookingService
	settingService ports.SettingService
}

func NewBookingHandler(service ports.BookingService, settingService ports.SettingService) *BookingHandler {
	return &BookingHandler{service: service, settingService: settingService}
}

// [GET] /api/bookings?start=...&end=...
// ในฟังก์ชัน GetBookings
func (h *BookingHandler) GetBookings(c *fiber.Ctx) error {
	start := c.Query("start")
	end := c.Query("end")
	userIdStr := c.Query("user_id") // รับค่ามาเป็น string ก่อน

	// 1. กรณีดึงข้อมูลปฏิทิน (กรองตามวันที่ start/end)
	if start != "" && end != "" {
		bookings, err := h.service.GetBookingsByRange(start, end)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}

		// Filter by status if provided (e.g. ?status=approved)
		statusFilter := c.Query("status")
		if statusFilter != "" {
			var filtered []domain.Booking
			for _, b := range bookings {
				if b.Status == statusFilter {
					filtered = append(filtered, b)
				}
			}
			return c.JSON(filtered)
		}

		return c.JSON(bookings)
	}

	// 2. ดึงข้อมูลทั้งหมดมาก่อน (เพื่อเตรียมกรอง)
	bookings, err := h.service.GetAllBookings()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// 3. กรณีดูประวัติส่วนตัว (กรองตาม User ID)
	if userIdStr != "" {
		// แปลง user_id จาก string เป็น int
		targetID, err := strconv.Atoi(userIdStr)
		if err != nil {
			// ถ้าส่งมาไม่ใช่ตัวเลข ให้แจ้ง error กลับไป
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user_id format"})
		}

		// สร้าง slice ใหม่เพื่อเก็บเฉพาะของ User คนนั้น
		var myBookings []domain.Booking
		for _, b := range bookings {
			if int(b.UserID) == targetID {
				myBookings = append(myBookings, b)
			}
		}
		return c.JSON(myBookings)
	}

	// 4. กรณี Admin หรือไม่ส่งอะไรมาเลย -> คืนค่าทั้งหมด
	// Filter by status if provided
    statusFilter := c.Query("status")
	if statusFilter != "" {
		var filtered []domain.Booking
		for _, b := range bookings {
			if b.Status == statusFilter {
				filtered = append(filtered, b)
			}
		}
		return c.JSON(filtered)
	}

	return c.JSON(bookings)
}

// POST /api/bookings (รองรับ File Upload)
func (h *BookingHandler) CreateBooking(c *fiber.Ctx) error {
	// 1. รับค่าจาก Form Data (ไม่ใช่ JSON แล้ว)
	// เราต้องแปลง string เป็น type ที่ถูกต้องเอง
	roomID, _ := strconv.Atoi(c.FormValue("room_id"))
	attendees, _ := strconv.Atoi(c.FormValue("attendees"))
	
	// แปลงเวลา (Time string -> Time object)
	layout := "2006-01-02T15:04:05.000Z" // Format ISO8601
	startTime, _ := time.Parse(layout, c.FormValue("start_time"))
	endTime, _ := time.Parse(layout, c.FormValue("end_time"))

	// สร้าง Object Booking
	booking := domain.Booking{
		// UserID จะรับจาก Form หรือ Token ก็ได้ (ในที่นี้รับจาก Form เพื่อความง่ายตาม Code เดิม)
		UserID:      uint(1), // Default ไว้ก่อน หรือแปลง c.FormValue("user_id")
		RoomID:      uint(roomID),
		Subject:     c.FormValue("subject"),
		Department:  c.FormValue("department"),
		Phone:       c.FormValue("phone"),
		Attendees:   attendees,
		StartTime:   startTime,
		EndTime:     endTime,
		Note:        c.FormValue("note"),
		ResourceText: c.FormValue("resource_text"),
		Status:      "pending",
	}
	
	// แก้ UserID ให้ถูกต้อง (ถ้าส่งมา)
	if uid, err := strconv.Atoi(c.FormValue("user_id")); err == nil {
		booking.UserID = uint(uid)
	}

	// 2. จัดการไฟล์อัปโหลด (Layout Image)
	file, err := c.FormFile("layout_image")
	if err == nil {
		// Retrieve Cloudinary Settings
		cloudName := h.settingService.GetSettingValue("cloudinary_cloud_name")
		apiKey := h.settingService.GetSettingValue("cloudinary_api_key")
		apiSecret := h.settingService.GetSettingValue("cloudinary_api_secret")

		// Check if Cloudinary is configured
		if cloudName != "" && apiKey != "" && apiSecret != "" {
			adapter, err := storage.NewCloudinaryAdapter(cloudName, apiKey, apiSecret)
			if err == nil {
				// Upload to Cloudinary
				url, err := adapter.Upload(file, fmt.Sprintf("booking_%d", time.Now().UnixNano()))
				if err == nil {
					booking.LayoutImage = url
				} else {
					// Join error if upload fails, or fallback? For safety let's return error
					return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Cloudinary Upload Failed: " + err.Error()})
				}
			} else {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to init Cloudinary: " + err.Error()})
			}
		} else {
			// Fallback: Local Storage
			// ถ้ามีการส่งไฟล์มา
			// ตั้งชื่อไฟล์ใหม่กันซ้ำ (เช่น booking_timestamp.jpg)
			filename := fmt.Sprintf("booking_%d_%s", time.Now().Unix(), file.Filename)
			path := fmt.Sprintf("./uploads/%s", filename)

			// บันทึกลงเครื่อง
			if err := c.SaveFile(file, path); err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to save file"})
			}

			// บันทึก Path ลง DB (เพื่อให้ Frontend เรียกใช้ได้)
			booking.LayoutImage = "/uploads/" + filename
		}
	}

	// 3. เรียก Service บันทึกข้อมูล
	if err := h.service.CreateBooking(&booking); err != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(booking)
}

// PATCH /api/bookings/:id/status
func (h *BookingHandler) UpdateStatus(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
	}

	// รับค่า status จาก Body เช่น { "status": "approved" }
	var input struct {
		Status string `json:"status"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	// ตรวจสอบว่า status ถูกต้องไหม
	if input.Status != "approved" && input.Status != "rejected" && input.Status != "pending" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid status"})
	}

	// ดึง User ID ของคนกด (Admin) จาก Token
	// (ใน Workshop นี้เราสมมติว่า Middleware แปะ user_id มาให้ หรือเราจะใช้จาก Claims ก็ได้)
	// เพื่อความง่ายตอนนี้เราจะ Hardcode หรือดึงจาก Token ถ้าทำ Middleware แล้ว
	// สมมติ admin_id = 1 ไปก่อนสำหรับการทดสอบ
	adminID := uint(1)

	if err := h.service.UpdateBookingStatus(uint(id), input.Status, adminID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Status updated successfully"})
}

// PUT /api/bookings/:id
func (h *BookingHandler) UpdateBooking(c *fiber.Ctx) error {
    id, err := c.ParamsInt("id")
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
    }

    var input struct {
        Subject   string    `json:"subject"`
        RoomID    uint      `json:"room_id"` // Note: JSON uses string for uint in some cases, but here assumes number
        StartTime time.Time `json:"start_time"`
        EndTime   time.Time `json:"end_time"`
        Note      string    `json:"note"`
    }

    if err := c.BodyParser(&input); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
    }

    if input.Subject == "" {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Subject cannot be empty"})
    }
    
    // Manual mapping or use domain.Booking directly if body matches
    // Note: The frontend sends JSON with room_id as string? Check BookingEditModal
    // BookingEditModal sends: room_id (string), start_time (ISO string)
    // Go Fiber BodyParser handles ISO string to time.Time automatically? Yes usually.
    // But room_id string to uint might fail if strict.
    // Let's use flexible struct or check frontend.
    // Frontend sends room_id as string. Backend expects uint or int.
    // Fiber parser is smart enough usually, but let's be safe.
    // Actually, `c.BodyParser` decodes based on the struct tag.
    // If I use `RoomID string` I can convert manually.

    // Let's try direct mapping to domain.Booking struct first, usually easier.
    // But domain.Booking has many fields.
    
    fmt.Printf("Updating Booking ID: %d with data: %+v\n", id, input) // Debug Log

    // Use domain booking for simplicity
    booking := domain.Booking{
		Subject:   input.Subject,
        // RoomID will need handling if frontend sends string
        // StartTime, EndTime handled
        Note: input.Note,
	}
    // Handle RoomID manually if needed, but if input.RoomID is uint, frontend MUST send number or string-number.
    booking.RoomID = input.RoomID
    booking.StartTime = input.StartTime
    booking.EndTime = input.EndTime
    
    // Get User ID from Token (Assuming Middleware puts it in Locals "user_id" or similar)
	userCtx := c.Locals("user")
    actorID := uint(0)
	if userCtx != nil {
        userToken := userCtx.(*jwt.Token)
        claims := userToken.Claims.(jwt.MapClaims)
        if idFloat, ok := claims["user_id"].(float64); ok {
            actorID = uint(idFloat)
        }
    }

    if err := h.service.UpdateBooking(uint(id), &booking, actorID); err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
    }
    
    return c.JSON(fiber.Map{"message": "Booking updated successfully"})
}

// DELETE /api/bookings/:id
// DELETE /api/bookings/:id
func (h *BookingHandler) DeleteBooking(c *fiber.Ctx) error {
    id, err := c.ParamsInt("id")
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
    }

	// Get User ID from Token
	// Note: Check if Locals("user") exists to avoid panic in case of middleware failure
	userCtx := c.Locals("user")
	if userCtx == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}
	
	userToken := userCtx.(*jwt.Token)
	claims := userToken.Claims.(jwt.MapClaims)
	
	// Safely retrieve user_id
	idInterface := claims["user_id"]
	var actorID uint
	if idFloat, ok := idInterface.(float64); ok {
		actorID = uint(idFloat)
	} else {
		// Fallback or Error
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid token claims"})
	}

    if err := h.service.DeleteBooking(uint(id), actorID); err != nil {
		if err.Error() == "unauthorized" || err.Error() == "you do not have permission to delete this booking" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
		}
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
    }
    
    return c.SendStatus(fiber.StatusOK)
}

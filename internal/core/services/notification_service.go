package services

import (
	"fmt"
	"html"
	"net/http"
	"net/url"
	"tunorth-brms-backend/internal/core/domain"
	"tunorth-brms-backend/internal/core/ports"
)

type notificationService struct {
	settings ports.SettingService
	roomRepo ports.RoomRepository
	userRepo ports.UserRepository
}

func NewNotificationService(settings ports.SettingService, roomRepo ports.RoomRepository, userRepo ports.UserRepository) ports.NotificationService {
	return &notificationService{
		settings: settings,
		roomRepo: roomRepo,
		userRepo: userRepo,
	}
}

func (s *notificationService) SendTelegram(chatID, message string) error {
	token := s.settings.GetSettingValue("telegram_bot_token")
	if token == "" || chatID == "" {
		return nil // ไม่ error แต่ไม่ส่ง
	}

	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token)
	formData := url.Values{
		"chat_id": {chatID},
		"text":    {message},
		"parse_mode": {"HTML"},
	}

	resp, err := http.PostForm(apiURL, formData)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send telegram message, status: %d", resp.StatusCode)
	}

	return nil
}

func (s *notificationService) NotifyAdminNewBooking(booking *domain.Booking) error {
	// เช็คว่าเปิดแจ้งเตือนไหม
	if s.settings.GetSettingValue("notify_admin") != "true" {
		return nil
	}

	adminChatID := s.settings.GetSettingValue("telegram_admin_chat_id")
	if adminChatID == "" {
		return nil
	}

	// Fetch Room Name
	roomName := fmt.Sprintf("ID %d", booking.RoomID)
	if room, err := s.roomRepo.GetByID(booking.RoomID); err == nil {
		roomName = room.RoomName
	}

	// Fetch User Name
	userName := fmt.Sprintf("ID %d", booking.UserID)
	if user, err := s.userRepo.GetByID(booking.UserID); err == nil {
		userName = user.FullName
	}

	// Link (Use 127.0.0.1 instead of localhost which Telegram often strips)
	link := `<a href="http://127.0.0.1:3000/admin/bookings">คลิกเพื่อดูรายละเอียดและอนุมัติ</a>`

	// Escape strings to prevent HTML parse errors
	subject := html.EscapeString(booking.Subject)
	rName := html.EscapeString(roomName)
	uName := html.EscapeString(userName)

	msg := fmt.Sprintf(
		"🔔 <b>มีการจองห้องประชุมใหม่</b> 🔔\n\n"+
			"📝 <b>หัวข้อ:</b> %s\n"+
			"🏢 <b>ห้อง:</b> %s\n"+
			"📅 <b>เวลา:</b> %s\n"+
			"👤 <b>ผู้จอง:</b> %s\n\n"+
			"🔗 <b>Link :</b> %s",
		subject,
		rName,
		booking.StartTime.Format("02/01/2006 15:04"),
		uName,
		link,
	)

	return s.SendTelegram(adminChatID, msg)
}

func (s *notificationService) NotifyUserStatusChange(booking *domain.Booking) error {
	// เช็คว่าเปิดแจ้งเตือนไหม
	if s.settings.GetSettingValue("notify_user") != "true" {
		return nil
	}

	// อันนี้อาจต้อง mapping User ID -> Telegram Chat ID ของ user คนนั้น
	// แต่ใน requirement ตอนนี้อาจจะส่งเข้า Group รวม หรือถ้ามี field telegram_id ใน user ก็ใช้ได้
	// สมมติส่งเข้า User Chat ID กลางที่ตั้งไว้ใน Settings ก่อน
	userChatID := s.settings.GetSettingValue("telegram_user_chat_id")
	
	// TODO: ถ้าอนาคต Users มี telegram_id ส่วนตัว ให้ดึงจาก userRepo.GetByID(booking.UserID).TelegramID
	
	if userChatID == "" {
		return nil
	}
	
	statusText := "รออนุมัติ"
	if booking.Status == "approved" {
		statusText = "✅ อนุมัติแล้ว"
	} else if booking.Status == "rejected" {
		statusText = "❌ ไม่อนุมัติ"
	}

	subject := html.EscapeString(booking.Subject)

	msg := fmt.Sprintf(
		"🔔 <b>สถานะการจองอัปเดต</b>\n\n"+
			"📝 <b>หัวข้อ:</b> %s\n"+
			"สถานะใหม่: <b>%s</b>",
		subject,
		statusText,
	)

	return s.SendTelegram(userChatID, msg)
}

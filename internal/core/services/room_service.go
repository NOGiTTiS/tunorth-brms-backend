package services

import (
	"errors"
	"fmt"
	"tunorth-brms-backend/internal/core/domain"
	"tunorth-brms-backend/internal/core/ports"
)

type roomService struct {
	repo       ports.RoomRepository
	logService ports.LogService
}

// NewRoomService รับ Repository เข้ามาเพื่อใช้งานต่อ
func NewRoomService(repo ports.RoomRepository, logService ports.LogService) ports.RoomService {
	return &roomService{repo: repo, logService: logService}
}

func (s *roomService) CreateRoom(room *domain.Room) error {
	// ตัวอย่าง Validation: ถ้าไม่ได้กรอกชื่อห้องมา ให้แจ้ง Error
	if room.RoomName == "" {
		return errors.New("room name is required")
	}
	
	// ถ้าผ่าน ก็ส่งต่อให้ Repo บันทึก
	if err := s.repo.Create(room); err != nil {
		return err
	}

	// Log
	go s.logService.LogAction(0, "CREATE_ROOM", fmt.Sprintf("Created room: %s", room.RoomName), "", "")
	
	return nil
}

func (s *roomService) GetAllRooms() ([]domain.Room, error) {
	return s.repo.GetAll()
}

func (s *roomService) GetRoomByID(id uint) (*domain.Room, error) {
	return s.repo.GetByID(id)
}

func (s *roomService) UpdateRoom(id uint, input *domain.Room) error {
	// 1. หาข้อมูลเก่าก่อนว่ามีไหม
	existingRoom, err := s.repo.GetByID(id)
	if err != nil {
		return errors.New("room not found")
	}

	// 2. อัปเดตข้อมูลเฉพาะส่วนที่มีการแก้ไข
	existingRoom.RoomName = input.RoomName
	existingRoom.Description = input.Description
	existingRoom.Capacity = input.Capacity
	existingRoom.Color = input.Color
	existingRoom.Status = input.Status
	// (ถ้ามีรูปภาพ image_path ก็อัปเดตตรงนี้)
	
	// 3. บันทึกลง DB
	if err := s.repo.Update(existingRoom); err != nil {
		return err
	}

	// Log
	go s.logService.LogAction(0, "UPDATE_ROOM", fmt.Sprintf("Updated room ID: %d", id), "", "")

	return nil
}

func (s *roomService) DeleteRoom(id uint) error {
	err := s.repo.Delete(id)
	if err == nil {
		go s.logService.LogAction(0, "DELETE_ROOM", fmt.Sprintf("Deleted room ID: %d", id), "", "")
	}
	return err
}
package ports

import "tunorth-brms-backend/internal/core/domain"

// RoomRepositoryInterface: บอกว่าต้องคุยกับ Database เรื่องห้องยังไง
type RoomRepository interface {
	Create(room *domain.Room) error
	GetAll() ([]domain.Room, error)
	GetByID(id uint) (*domain.Room, error)
	Update(room *domain.Room) error
	Delete(id uint) error
}

// RoomServiceInterface: บอกว่า Business Logic ของห้องมีอะไรบ้าง
type RoomService interface {
	CreateRoom(room *domain.Room) error
	GetAllRooms() ([]domain.Room, error)
	GetRoomByID(id uint) (*domain.Room, error)
	UpdateRoom(id uint, room *domain.Room) error
	DeleteRoom(id uint) error
}
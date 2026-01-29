package services

import (
	"errors"
	"os"
	"time"
	"tunorth-brms-backend/internal/core/domain"
	"tunorth-brms-backend/internal/core/ports"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type authService struct {
	userRepo ports.UserRepository
}

func NewAuthService(userRepo ports.UserRepository) ports.AuthService {
	return &authService{userRepo: userRepo}
}

// Register: สมัครสมาชิก (Hash Password ก่อนบันทึก)
func (s *authService) Register(user *domain.User) error {
	// 1. ตรวจสอบว่ามี username นี้หรือยัง
	existingUser, _ := s.userRepo.GetByUsername(user.Username)
	if existingUser != nil && existingUser.ID != 0 {
		return errors.New("username already exists")
	}

	// 2. ตรวจสอบว่ามี Email นี้หรือยัง
	existingEmail, _ := s.userRepo.GetByEmail(user.Email)
	if existingEmail != nil && existingEmail.ID != 0 {
		return errors.New("email already exists")
	}

	// 3. Hash Password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hashedPassword)

	// 3. บันทึก
	return s.userRepo.Create(user)
}

// Login: ตรวจสอบรหัสและออก Token (รองรับ Username หรือ Email)
func (s *authService) Login(identifier, password string) (string, uint, error) {
	// 1. หา User (By Username or Email)
	user, err := s.userRepo.GetByUsernameOrEmail(identifier)
	if err != nil {
		return "", 0, errors.New("invalid username/email or password")
	}

	// 2. ตรวจสอบรหัสผ่าน (Hash vs Plain)
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", 0, errors.New("invalid username or password")
	}

	// 3. สร้าง JWT Token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"role":     user.Role,
		"exp":      time.Now().Add(time.Hour * 24).Unix(), // หมดอายุใน 24 ชม.
	})

	// เซ็นลายเซ็นด้วย Secret Key (จาก .env)
	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return "", 0, err
	}

	return tokenString, user.ID, nil
}

func (s *authService) GetMe(userID uint) (*domain.User, error) {
	return s.userRepo.GetByID(userID)
}

func (s *authService) UpdateMe(userID uint, updates *domain.User) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}

	// Update fields
	user.FullName = updates.FullName
	user.Email = updates.Email
	user.Tel = updates.Tel
    user.Department = updates.Department

	// If password provided, hash it
	if updates.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(updates.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		user.Password = string(hashedPassword)
	}

	return s.userRepo.Update(user)
}
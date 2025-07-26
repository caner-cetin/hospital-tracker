package services

import (
	"crypto/rand"
	"errors"
	"fmt"
	"time"

	"github.com/caner-cetin/hospital-tracker/internal/models"
	"gorm.io/gorm"
)

type PasswordResetService struct {
	db          *gorm.DB
	authService *AuthService
}

func NewPasswordResetService(db *gorm.DB, authService *AuthService) *PasswordResetService {
	return &PasswordResetService{
		db:          db,
		authService: authService,
	}
}

func (s *PasswordResetService) RequestPasswordReset(phone string) (string, error) {
	var user models.User
	if err := s.db.Where("phone = ?", phone).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errors.New("phone number not found")
		}
		return "", err
	}

	s.db.Where("phone = ? AND used = false", phone).Delete(&models.PasswordReset{})

	code := s.generateCode()
	passwordReset := &models.PasswordReset{
		Phone:     phone,
		Code:      code,
		ExpiresAt: time.Now().Add(15 * time.Minute),
		Used:      false,
	}

	if err := s.db.Create(passwordReset).Error; err != nil {
		return "", err
	}

	return code, nil
}

func (s *PasswordResetService) ResetPassword(phone, code, newPassword, confirmPassword string) error {
	if newPassword != confirmPassword {
		return errors.New("passwords do not match")
	}

	var passwordReset models.PasswordReset
	err := s.db.Where("phone = ? AND code = ? AND used = false AND expires_at > ?",
		phone, code, time.Now()).First(&passwordReset).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("invalid or expired reset code")
		}
		return err
	}

	var user models.User
	if err := s.db.Where("phone = ?", phone).First(&user).Error; err != nil {
		return err
	}

	hashedPassword, err := s.authService.HashPassword(newPassword)
	if err != nil {
		return err
	}

	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Model(&user).Update("password", hashedPassword).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Model(&passwordReset).Update("used", true).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (s *PasswordResetService) generateCode() string {
	b := make([]byte, 4)
	_, err := rand.Read(b)
	if err != nil {
		return "000000"
	}
	n := uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])
	return fmt.Sprintf("%06d", n%1000000)
}

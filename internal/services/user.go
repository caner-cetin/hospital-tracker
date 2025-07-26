package services

import (
	"errors"

	userErrors "github.com/caner-cetin/hospital-tracker/internal/errors"
	"github.com/caner-cetin/hospital-tracker/internal/models"
	"gorm.io/gorm"
)

type UserService struct {
	db          *gorm.DB
	authService *AuthService
}

func NewUserService(db *gorm.DB, authService *AuthService) *UserService {
	return &UserService{
		db:          db,
		authService: authService,
	}
}

func (s *UserService) CreateUser(req *models.CreateUserRequest, createdByID uint, hospitalID uint) (*models.User, error) {
	if err := s.validateUserUniqueness(req.NationalID, req.Email, req.Phone); err != nil {
		return nil, err
	}

	hashedPassword, err := s.authService.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		NationalID:  req.NationalID,
		Email:       req.Email,
		Phone:       req.Phone,
		Password:    hashedPassword,
		UserType:    req.UserType,
		HospitalID:  hospitalID,
		CreatedByID: &createdByID,
	}

	if err := s.db.Create(user).Error; err != nil {
		return nil, err
	}

	if err := s.db.Preload("Hospital").Preload("CreatedBy").First(user, user.ID).Error; err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) UpdateUser(userID uint, req *models.UpdateUserRequest, hospitalID uint) (*models.User, error) {
	var user models.User
	if err := s.db.Where("id = ? AND hospital_id = ?", userID, hospitalID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, userErrors.NewUserNotFoundError()
		}
		return nil, err
	}

	if req.NationalID != "" && req.NationalID != user.NationalID {
		if err := s.checkUniqueness("national_id", req.NationalID, userID); err != nil {
			return nil, err
		}
		user.NationalID = req.NationalID
	}

	if req.Email != "" && req.Email != user.Email {
		if err := s.checkUniqueness("email", req.Email, userID); err != nil {
			return nil, err
		}
		user.Email = req.Email
	}

	if req.Phone != "" && req.Phone != user.Phone {
		if err := s.checkUniqueness("phone", req.Phone, userID); err != nil {
			return nil, err
		}
		user.Phone = req.Phone
	}

	if req.FirstName != "" {
		user.FirstName = req.FirstName
	}

	if req.LastName != "" {
		user.LastName = req.LastName
	}

	if req.UserType != "" {
		user.UserType = req.UserType
	}

	if err := s.db.Save(&user).Error; err != nil {
		return nil, err
	}

	if err := s.db.Preload("Hospital").Preload("CreatedBy").First(&user, user.ID).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *UserService) DeleteUser(userID uint, hospitalID uint) error {
	var user models.User
	if err := s.db.Where("id = ? AND hospital_id = ?", userID, hospitalID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return userErrors.NewUserNotFoundError()
		}
		return userErrors.NewDatabaseError("user operation", err)
	}

	if err := s.db.Delete(&user).Error; err != nil {
		return userErrors.NewDatabaseError("delete user", err)
	}
	return nil
}

func (s *UserService) GetUsers(hospitalID uint) ([]models.User, error) {
	var users []models.User
	err := s.db.Where("hospital_id = ?", hospitalID).
		Preload("Hospital").
		Preload("CreatedBy").
		Find(&users).Error

	if err != nil {
		return nil, userErrors.NewDatabaseError("get users", err)
	}
	return users, nil
}

func (s *UserService) GetUser(userID uint, hospitalID uint) (*models.User, error) {
	var user models.User
	err := s.db.Where("id = ? AND hospital_id = ?", userID, hospitalID).
		Preload("Hospital").
		Preload("CreatedBy").
		First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, userErrors.NewUserNotFoundError()
		}
		return nil, err
	}

	return &user, nil
}

func (s *UserService) validateUserUniqueness(nationalID, email, phone string) error {
	var count int64

	if err := s.db.Model(&models.User{}).Where("national_id = ?", nationalID).Count(&count).Error; err != nil {
		return userErrors.NewDatabaseError("check national ID uniqueness", err)
	}
	if count > 0 {
		return userErrors.NewDuplicateNationalIDError(nationalID)
	}

	if err := s.db.Model(&models.User{}).Where("email = ?", email).Count(&count).Error; err != nil {
		return userErrors.NewDatabaseError("check email uniqueness", err)
	}
	if count > 0 {
		return userErrors.NewDuplicateEmailError(email)
	}

	if err := s.db.Model(&models.User{}).Where("phone = ?", phone).Count(&count).Error; err != nil {
		return userErrors.NewDatabaseError("check phone uniqueness", err)
	}
	if count > 0 {
		return userErrors.NewDuplicatePhoneError(phone)
	}

	return nil
}

func (s *UserService) checkUniqueness(field, value string, excludeID uint) error {
	var count int64
	err := s.db.Model(&models.User{}).Where(field+" = ? AND id != ?", value, excludeID).Count(&count).Error
	if err != nil {
		return userErrors.NewDatabaseError("user operation", err)
	}
	if count > 0 {
		return userErrors.NewConflictError(field, value, "user")
	}
	return nil
}

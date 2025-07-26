package services

import (
	"errors"

	hospitalErrors "github.com/caner-cetin/hospital-tracker/internal/errors"
	"github.com/caner-cetin/hospital-tracker/internal/models"
	"gorm.io/gorm"
)

type HospitalService struct {
	db          *gorm.DB
	authService *AuthService
}

func NewHospitalService(db *gorm.DB, authService *AuthService) *HospitalService {
	return &HospitalService{
		db:          db,
		authService: authService,
	}
}

func (s *HospitalService) RegisterHospital(req *models.HospitalRegistrationRequest) (*models.Hospital, *models.User, error) {
	if err := s.validateHospitalUniqueness(req); err != nil {
		return nil, nil, err
	}

	if err := s.validateUserUniqueness(req); err != nil {
		return nil, nil, err
	}

	if err := s.validateProvinceDistrict(req.ProvinceID, req.DistrictID); err != nil {
		return nil, nil, err
	}

	hashedPassword, err := s.authService.HashPassword(req.Password)
	if err != nil {
		return nil, nil, hospitalErrors.NewInternalError("password hashing failed", err)
	}

	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	hospital := &models.Hospital{
		Name:       req.HospitalName,
		TaxID:      req.TaxID,
		Email:      req.Email,
		Phone:      req.Phone,
		ProvinceID: req.ProvinceID,
		DistrictID: req.DistrictID,
		Address:    req.Address,
	}

	if err := tx.Create(hospital).Error; err != nil {
		tx.Rollback()
		return nil, nil, hospitalErrors.NewDatabaseError("create hospital", err)
	}

	user := &models.User{
		FirstName:  req.FirstName,
		LastName:   req.LastName,
		NationalID: req.NationalID,
		Email:      req.UserEmail,
		Phone:      req.UserPhone,
		Password:   hashedPassword,
		UserType:   models.UserTypeAuthorized,
		HospitalID: hospital.ID,
	}

	if err := tx.Create(user).Error; err != nil {
		tx.Rollback()
		return nil, nil, hospitalErrors.NewDatabaseError("create user", err)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, nil, hospitalErrors.NewDatabaseError("commit transaction", err)
	}

	if err := s.db.Preload("Province").Preload("District").First(hospital, hospital.ID).Error; err != nil {
		return nil, nil, hospitalErrors.NewDatabaseError("load hospital data", err)
	}

	if err := s.db.Preload("Hospital").First(user, user.ID).Error; err != nil {
		return nil, nil, hospitalErrors.NewDatabaseError("load user data", err)
	}

	return hospital, user, nil
}

func (s *HospitalService) validateHospitalUniqueness(req *models.HospitalRegistrationRequest) error {
	var count int64

	if err := s.db.Model(&models.Hospital{}).Where("tax_id = ?", req.TaxID).Count(&count).Error; err != nil {
		return hospitalErrors.NewDatabaseError("check tax ID uniqueness", err)
	}
	if count > 0 {
		return hospitalErrors.NewConflictError("tax_id", req.TaxID, "hospital")
	}

	if err := s.db.Model(&models.Hospital{}).Where("email = ?", req.Email).Count(&count).Error; err != nil {
		return hospitalErrors.NewDatabaseError("check hospital email uniqueness", err)
	}
	if count > 0 {
		return hospitalErrors.NewDuplicateEmailError(req.Email)
	}

	if err := s.db.Model(&models.Hospital{}).Where("phone = ?", req.Phone).Count(&count).Error; err != nil {
		return hospitalErrors.NewDatabaseError("check hospital phone uniqueness", err)
	}
	if count > 0 {
		return hospitalErrors.NewDuplicatePhoneError(req.Phone)
	}

	return nil
}

func (s *HospitalService) validateUserUniqueness(req *models.HospitalRegistrationRequest) error {
	var count int64

	if err := s.db.Model(&models.User{}).Where("national_id = ?", req.NationalID).Count(&count).Error; err != nil {
		return hospitalErrors.NewDatabaseError("check national ID uniqueness", err)
	}
	if count > 0 {
		return hospitalErrors.NewDuplicateNationalIDError(req.NationalID)
	}

	if err := s.db.Model(&models.User{}).Where("email = ?", req.UserEmail).Count(&count).Error; err != nil {
		return hospitalErrors.NewDatabaseError("check user email uniqueness", err)
	}
	if count > 0 {
		return hospitalErrors.NewDuplicateEmailError(req.UserEmail)
	}

	if err := s.db.Model(&models.User{}).Where("phone = ?", req.UserPhone).Count(&count).Error; err != nil {
		return hospitalErrors.NewDatabaseError("check user phone uniqueness", err)
	}
	if count > 0 {
		return hospitalErrors.NewDuplicatePhoneError(req.UserPhone)
	}

	return nil
}

func (s *HospitalService) validateProvinceDistrict(provinceID, districtID uint) error {
	var district models.District
	err := s.db.Where("id = ? AND province_id = ?", districtID, provinceID).First(&district).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return hospitalErrors.NewValidationError("province_district", "invalid province or district")
		}
		return hospitalErrors.NewDatabaseError("validate province/district", err)
	}
	return nil
}

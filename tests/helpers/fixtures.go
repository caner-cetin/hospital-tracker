package helpers

import (
	"github.com/caner-cetin/hospital-tracker/internal/models"
	"github.com/caner-cetin/hospital-tracker/internal/services"
	"github.com/go-faker/faker/v4"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

func CreateTestProvince(db *gorm.DB) (*models.Province, *models.District, error) {
	var province models.Province
	if err := db.First(&province).Error; err != nil {
		province = models.Province{Name: faker.Word()}
		if err := db.Create(&province).Error; err != nil {
			return nil, nil, err
		}
	}

	var district models.District
	if err := db.Where("province_id = ?", province.ID).First(&district).Error; err != nil {
		district = models.District{
			Name:       faker.Word(),
			ProvinceID: province.ID,
		}
		if err := db.Create(&district).Error; err != nil {
			return nil, nil, err
		}
	}

	return &province, &district, nil
}

func CreateTestHospital(db *gorm.DB, authService *services.AuthService) (*models.Hospital, *models.User, string, error) {
	var province models.Province
	if err := db.First(&province).Error; err != nil {
		province = models.Province{Name: faker.Word()}
		if err := db.Create(&province).Error; err != nil {
			return nil, nil, "", err
		}
	}

	var district models.District
	if err := db.Where("province_id = ?", province.ID).First(&district).Error; err != nil {
		district = models.District{
			Name:       faker.Word(),
			ProvinceID: province.ID,
		}
		if err := db.Create(&district).Error; err != nil {
			return nil, nil, "", err
		}
	}

	req := &models.HospitalRegistrationRequest{
		HospitalName: faker.Name() + " Hospital",
		TaxID:        faker.CCNumber(),
		Email:        faker.Email(),
		Phone:        faker.Phonenumber(),
		ProvinceID:   province.ID,
		DistrictID:   district.ID,
		Address:      faker.Sentence(),
		FirstName:    faker.FirstName(),
		LastName:     faker.LastName(),
		NationalID:   faker.UUIDDigit()[:11],
		UserEmail:    faker.Email(),
		UserPhone:    faker.Phonenumber(),
		Password:     faker.Password(),
	}

	hospitalService := services.NewHospitalService(db, authService)
	hospital, user, err := hospitalService.RegisterHospital(req)
	if err != nil {
		return nil, nil, "", errors.Wrap(err, "failed to register hospital")
	}
	return hospital, user, req.Password, nil
}

func CreateTestUser(db *gorm.DB, authService *services.AuthService, hospitalID uint, userType models.UserType) (*models.User, error) {
	var createdBy models.User
	if err := db.Where("hospital_id = ? AND user_type = ?", hospitalID, models.UserTypeAuthorized).First(&createdBy).Error; err != nil {
		return nil, err
	}

	req := &models.CreateUserRequest{
		FirstName:  faker.FirstName(),
		LastName:   faker.LastName(),
		NationalID: faker.UUIDDigit()[:11],
		Email:      faker.Email(),
		Phone:      faker.Phonenumber(),
		Password:   faker.Password(),
		UserType:   userType,
	}

	userService := services.NewUserService(db, authService)
	user, err := userService.CreateUser(req, createdBy.ID, hospitalID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create user")
	}
	return user, nil
}

func CreateTestClinic(db *gorm.DB, hospitalID uint) (*models.Clinic, error) {
	var clinicType models.ClinicType
	if err := db.First(&clinicType).Error; err != nil {
		clinicType = models.ClinicType{Name: faker.Word()}
		if err := db.Create(&clinicType).Error; err != nil {
			return nil, err
		}
	}

	req := &models.CreateClinicRequest{
		ClinicTypeID: clinicType.ID,
	}

	clinicService := services.NewClinicService(db)
	clinic, err := clinicService.CreateClinic(req, hospitalID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create clinic")
	}
	return clinic, nil
}

func CreateTestStaff(db *gorm.DB, hospitalID uint, clinicID *uint) (*models.Staff, error) {
	var professionGroup models.ProfessionGroup
	if err := db.Where("name = ?", "Doktor").First(&professionGroup).Error; err != nil {
		return nil, err
	}

	var title models.Title
	if err := db.Where("name = ? AND profession_group_id = ?", "Asistan", professionGroup.ID).First(&title).Error; err != nil {
		return nil, err
	}

	req := &models.CreateStaffRequest{
		FirstName:         faker.FirstName(),
		LastName:          faker.LastName(),
		NationalID:        faker.UUIDDigit()[:11],
		Phone:             faker.Phonenumber(),
		ProfessionGroupID: professionGroup.ID,
		TitleID:           title.ID,
		ClinicID:          clinicID,
		WorkingDays:       []models.WorkingDay{models.Monday, models.Tuesday, models.Wednesday},
	}

	staffService := services.NewStaffService(db, nil)
	staff, err := staffService.CreateStaff(req, hospitalID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create staff")
	}
	return staff, nil
}

func GetTestJWTToken(authService *services.AuthService, user *models.User) (string, error) {
	token, err := authService.GenerateToken(user)
	if err != nil {
		return "", errors.Wrap(err, "failed to generate token")
	}
	return token, nil
}

func CreateTestPasswordReset(db *gorm.DB, phone string) (*models.PasswordReset, string, error) {
	if phone == "" {
		phone = faker.Phonenumber()
	}

	passwordResetService := services.NewPasswordResetService(db, nil)
	code, err := passwordResetService.RequestPasswordReset(phone)
	if err != nil {
		return nil, "", errors.Wrap(err, "failed to request password reset")
	}

	var passwordReset models.PasswordReset
	err = db.Where("phone = ? AND code = ?", phone, code).First(&passwordReset).Error
	if err != nil {
		return nil, "", errors.Wrap(err, "failed to find password reset")
	}
	return &passwordReset, code, nil
}

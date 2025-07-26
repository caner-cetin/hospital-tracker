package unit

import (
	"context"
	"testing"

	"github.com/caner-cetin/hospital-tracker/internal/errors"
	"github.com/caner-cetin/hospital-tracker/internal/models"
	"github.com/caner-cetin/hospital-tracker/internal/services"
	"github.com/caner-cetin/hospital-tracker/tests/helpers"
	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/suite"
)

type HospitalServiceTestSuite struct {
	suite.Suite
	containers      *helpers.TestContainers
	hospitalService *services.HospitalService
	authService     *services.AuthService
}

func (suite *HospitalServiceTestSuite) SetupSuite() {
	ctx := context.Background()
	containers, err := helpers.SetupTestContainers(ctx)
	suite.Require().NoError(err)

	suite.containers = containers
	suite.authService = services.NewAuthService(containers.DB, containers.Config)
	suite.hospitalService = services.NewHospitalService(containers.DB, suite.authService)
}

func (suite *HospitalServiceTestSuite) TearDownSuite() {
	ctx := context.Background()
	if suite.containers != nil {
		_ = suite.containers.Cleanup(ctx)
	}
}

func (suite *HospitalServiceTestSuite) SetupTest() {
	err := suite.containers.CleanDatabase()
	suite.Require().NoError(err)
}

func (suite *HospitalServiceTestSuite) TestRegisterHospital() {
	province, district, err := helpers.CreateTestProvince(suite.containers.DB)
	suite.Require().NoError(err)

	req := &models.HospitalRegistrationRequest{
		HospitalName: faker.Name(),
		TaxID:        faker.UUIDDigit(),
		Email:        faker.Email(),
		Phone:        faker.Phonenumber(),
		ProvinceID:   province.ID,
		DistrictID:   district.ID,
		Address:      faker.Sentence(),
		FirstName:    faker.FirstName(),
		LastName:     faker.LastName(),
		NationalID:   faker.UUIDDigit(),
		UserEmail:    faker.Email(),
		UserPhone:    faker.Phonenumber(),
		Password:     faker.Password(),
	}

	hospital, user, err := suite.hospitalService.RegisterHospital(req)

	suite.NoError(err)
	suite.NotNil(hospital)
	suite.NotNil(user)

	suite.Equal(req.HospitalName, hospital.Name)
	suite.Equal(req.TaxID, hospital.TaxID)
	suite.Equal(req.Email, hospital.Email)
	suite.Equal(req.Phone, hospital.Phone)
	suite.Equal(req.ProvinceID, hospital.ProvinceID)
	suite.Equal(req.DistrictID, hospital.DistrictID)
	suite.Equal(req.Address, hospital.Address)

	suite.Equal(req.FirstName, user.FirstName)
	suite.Equal(req.LastName, user.LastName)
	suite.Equal(req.NationalID, user.NationalID)
	suite.Equal(req.UserEmail, user.Email)
	suite.Equal(req.UserPhone, user.Phone)
	suite.Equal(models.UserTypeAuthorized, user.UserType)
	suite.Equal(hospital.ID, user.HospitalID)
}

func (suite *HospitalServiceTestSuite) TestRegisterHospitalDuplicateTaxID() {
	province, district, err := helpers.CreateTestProvince(suite.containers.DB)
	suite.Require().NoError(err)

	taxID := faker.UUIDDigit()
	req1 := &models.HospitalRegistrationRequest{
		HospitalName: faker.Name(),
		TaxID:        taxID,
		Email:        faker.Email(),
		Phone:        faker.Phonenumber(),
		ProvinceID:   province.ID,
		DistrictID:   district.ID,
		Address:      faker.Sentence(),
		FirstName:    faker.FirstName(),
		LastName:     faker.LastName(),
		NationalID:   faker.UUIDDigit(),
		UserEmail:    faker.Email(),
		UserPhone:    faker.Phonenumber(),
		Password:     faker.Password(),
	}

	hospital1, user1, err := suite.hospitalService.RegisterHospital(req1)
	suite.Require().NoError(err)
	suite.Require().NotNil(hospital1)
	suite.Require().NotNil(user1)

	req2 := &models.HospitalRegistrationRequest{
		HospitalName: faker.Name(),
		TaxID:        taxID,
		Email:        faker.Email(),
		Phone:        faker.Phonenumber(),
		ProvinceID:   province.ID,
		DistrictID:   district.ID,
		Address:      faker.Sentence(),
		FirstName:    faker.FirstName(),
		LastName:     faker.LastName(),
		NationalID:   faker.UUIDDigit(),
		UserEmail:    faker.Email(),
		UserPhone:    faker.Phonenumber(),
		Password:     faker.Password(),
	}

	hospital2, user2, err := suite.hospitalService.RegisterHospital(req2)
	suite.Error(err)
	suite.Nil(hospital2)
	suite.Nil(user2)

	// Check that it's a conflict error for tax ID
	if appErr, ok := errors.IsAppError(err); ok {
		suite.Equal(errors.ErrCodeConflict, appErr.Code)
	} else {
		suite.T().Errorf("Expected AppError, got %T", err)
	}
}

func (suite *HospitalServiceTestSuite) TestRegisterHospitalDuplicateHospitalEmail() {
	province, district, err := helpers.CreateTestProvince(suite.containers.DB)
	suite.Require().NoError(err)

	hospitalEmail := faker.Email()
	req1 := &models.HospitalRegistrationRequest{
		HospitalName: faker.Name(),
		TaxID:        faker.UUIDDigit(),
		Email:        hospitalEmail,
		Phone:        faker.Phonenumber(),
		ProvinceID:   province.ID,
		DistrictID:   district.ID,
		Address:      faker.Sentence(),
		FirstName:    faker.FirstName(),
		LastName:     faker.LastName(),
		NationalID:   faker.UUIDDigit(),
		UserEmail:    faker.Email(),
		UserPhone:    faker.Phonenumber(),
		Password:     faker.Password(),
	}

	hospital1, user1, err := suite.hospitalService.RegisterHospital(req1)
	suite.Require().NoError(err)
	suite.Require().NotNil(hospital1)
	suite.Require().NotNil(user1)

	req2 := &models.HospitalRegistrationRequest{
		HospitalName: faker.Name(),
		TaxID:        faker.UUIDDigit(),
		Email:        hospitalEmail,
		Phone:        faker.Phonenumber(),
		ProvinceID:   province.ID,
		DistrictID:   district.ID,
		Address:      faker.Sentence(),
		FirstName:    faker.FirstName(),
		LastName:     faker.LastName(),
		NationalID:   faker.UUIDDigit(),
		UserEmail:    faker.Email(),
		UserPhone:    faker.Phonenumber(),
		Password:     faker.Password(),
	}

	hospital2, user2, err := suite.hospitalService.RegisterHospital(req2)
	suite.Error(err)
	suite.Nil(hospital2)
	suite.Nil(user2)

	// check that it's a duplicate email error
	if appErr, ok := errors.IsAppError(err); ok {
		suite.Equal(errors.ErrCodeDuplicateEmail, appErr.Code)
	} else {
		suite.T().Errorf("Expected AppError, got %T", err)
	}
}

func (suite *HospitalServiceTestSuite) TestRegisterHospitalDuplicateUserNationalID() {
	province, district, err := helpers.CreateTestProvince(suite.containers.DB)
	suite.Require().NoError(err)

	nationalID := faker.UUIDDigit()
	req1 := &models.HospitalRegistrationRequest{
		HospitalName: faker.Name(),
		TaxID:        faker.UUIDDigit(),
		Email:        faker.Email(),
		Phone:        faker.Phonenumber(),
		ProvinceID:   province.ID,
		DistrictID:   district.ID,
		Address:      faker.Sentence(),
		FirstName:    faker.FirstName(),
		LastName:     faker.LastName(),
		NationalID:   nationalID,
		UserEmail:    faker.Email(),
		UserPhone:    faker.Phonenumber(),
		Password:     faker.Password(),
	}

	hospital1, user1, err := suite.hospitalService.RegisterHospital(req1)
	suite.Require().NoError(err)
	suite.Require().NotNil(hospital1)
	suite.Require().NotNil(user1)

	req2 := &models.HospitalRegistrationRequest{
		HospitalName: faker.Name(),
		TaxID:        faker.UUIDDigit(),
		Email:        faker.Email(),
		Phone:        faker.Phonenumber(),
		ProvinceID:   province.ID,
		DistrictID:   district.ID,
		Address:      faker.Sentence(),
		FirstName:    faker.FirstName(),
		LastName:     faker.LastName(),
		NationalID:   nationalID,
		UserEmail:    faker.Email(),
		UserPhone:    faker.Phonenumber(),
		Password:     faker.Password(),
	}

	hospital2, user2, err := suite.hospitalService.RegisterHospital(req2)
	suite.Error(err)
	suite.Nil(hospital2)
	suite.Nil(user2)
}

func (suite *HospitalServiceTestSuite) TestRegisterHospitalInvalidProvinceDistrict() {
	req := &models.HospitalRegistrationRequest{
		HospitalName: faker.Name(),
		TaxID:        faker.UUIDDigit(),
		Email:        faker.Email(),
		Phone:        faker.Phonenumber(),
		ProvinceID:   1,
		DistrictID:   999,
		Address:      faker.Sentence(),
		FirstName:    faker.FirstName(),
		LastName:     faker.LastName(),
		NationalID:   faker.UUIDDigit(),
		UserEmail:    faker.Email(),
		UserPhone:    faker.Phonenumber(),
		Password:     faker.Password(),
	}

	hospital, user, err := suite.hospitalService.RegisterHospital(req)
	suite.Error(err)
	suite.Nil(hospital)
	suite.Nil(user)

	// check that it's a validation error
	if appErr, ok := errors.IsAppError(err); ok {
		suite.Equal(errors.ErrCodeValidation, appErr.Code)
	} else {
		suite.T().Errorf("Expected AppError, got %T", err)
	}
}

func TestHospitalServiceTestSuite(t *testing.T) {
	suite.Run(t, new(HospitalServiceTestSuite))
}

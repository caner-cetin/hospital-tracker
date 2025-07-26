package integration

import (
	"context"
	"testing"

	"github.com/caner-cetin/hospital-tracker/internal/models"
	"github.com/caner-cetin/hospital-tracker/internal/services"
	"github.com/caner-cetin/hospital-tracker/tests/helpers"
	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/suite"
)

type StaffServiceIntegrationTestSuite struct {
	suite.Suite
	containers    *helpers.TestContainers
	staffService  *services.StaffService
	clinicService *services.ClinicService
	authService   *services.AuthService
	hospitalID    uint
	clinicID      uint
}

func (suite *StaffServiceIntegrationTestSuite) SetupSuite() {
	ctx := context.Background()
	containers, err := helpers.SetupTestContainers(ctx)
	suite.Require().NoError(err)

	suite.containers = containers
	suite.authService = services.NewAuthService(containers.DB, containers.Config)
	suite.staffService = services.NewStaffService(containers.DB, containers.Redis)
	suite.clinicService = services.NewClinicService(containers.DB)
}

func (suite *StaffServiceIntegrationTestSuite) TearDownSuite() {
	ctx := context.Background()
	if suite.containers != nil {
		_ = suite.containers.Cleanup(ctx)
	}
}

func (suite *StaffServiceIntegrationTestSuite) SetupTest() {
	err := suite.containers.CleanDatabase()
	suite.Require().NoError(err)

	hospital, _, _, err := helpers.CreateTestHospital(suite.containers.DB, suite.authService)
	suite.Require().NoError(err)
	suite.hospitalID = hospital.ID

	clinic, err := helpers.CreateTestClinic(suite.containers.DB, suite.hospitalID)
	suite.Require().NoError(err)
	suite.clinicID = clinic.ID
}

func (suite *StaffServiceIntegrationTestSuite) getProfessionGroupAndTitle(professionGroupName, titleName string) (uint, uint) {
	var professionGroup models.ProfessionGroup
	err := suite.containers.DB.Where("name = ?", professionGroupName).First(&professionGroup).Error
	suite.Require().NoError(err, "Failed to find profession group: %s", professionGroupName)

	var title models.Title
	err = suite.containers.DB.Where("name = ? AND profession_group_id = ?", titleName, professionGroup.ID).First(&title).Error
	suite.Require().NoError(err, "Failed to find title: %s for profession group: %s", titleName, professionGroupName)

	return professionGroup.ID, title.ID
}

func (suite *StaffServiceIntegrationTestSuite) TestCreateStaff() {
	professionGroupID, titleID := suite.getProfessionGroupAndTitle("Doktor", "Asistan")

	req := &models.CreateStaffRequest{
		FirstName:         "John",
		LastName:          "Smith",
		NationalID:        "12345678901",
		Phone:             "+905551234567",
		ProfessionGroupID: professionGroupID,
		TitleID:           titleID,
		ClinicID:          &suite.clinicID,
		WorkingDays:       []models.WorkingDay{models.Monday, models.Tuesday, models.Wednesday},
	}

	staff, err := suite.staffService.CreateStaff(req, suite.hospitalID)

	suite.NoError(err)
	suite.NotNil(staff)
	suite.Equal(req.FirstName, staff.FirstName)
	suite.Equal(req.LastName, staff.LastName)
	suite.Equal(req.NationalID, staff.NationalID)
	suite.Equal(req.Phone, staff.Phone)
	suite.Equal(req.ProfessionGroupID, staff.ProfessionGroupID)
	suite.Equal(req.TitleID, staff.TitleID)
	suite.Equal(suite.hospitalID, staff.HospitalID)
	suite.Equal(suite.clinicID, *staff.ClinicID)
}

func (suite *StaffServiceIntegrationTestSuite) TestCreateStaffWithoutClinic() {
	professionGroupID, titleID := suite.getProfessionGroupAndTitle("Hizmet Personeli", "Güvenlik")

	req := &models.CreateStaffRequest{
		FirstName:         "Security",
		LastName:          "Guard",
		NationalID:        "12345678902",
		Phone:             "+905551234568",
		ProfessionGroupID: professionGroupID,
		TitleID:           titleID,
		ClinicID:          nil,
		WorkingDays:       []models.WorkingDay{models.Monday, models.Tuesday, models.Wednesday, models.Thursday, models.Friday},
	}

	staff, err := suite.staffService.CreateStaff(req, suite.hospitalID)

	suite.NoError(err)
	suite.NotNil(staff)
	suite.Nil(staff.ClinicID)
}

func (suite *StaffServiceIntegrationTestSuite) TestCreateStaffChiefPhysicianRestriction() {
	professionGroupID, titleID := suite.getProfessionGroupAndTitle("İdari Personel", "Başhekim")

	req1 := &models.CreateStaffRequest{
		FirstName:         "Chief",
		LastName:          "Physician",
		NationalID:        "12345678901",
		Phone:             "+905551234567",
		ProfessionGroupID: professionGroupID,
		TitleID:           titleID,
		ClinicID:          nil,
		WorkingDays:       []models.WorkingDay{models.Monday, models.Tuesday, models.Wednesday},
	}

	staff1, err := suite.staffService.CreateStaff(req1, suite.hospitalID)
	suite.Require().NoError(err)
	suite.Require().NotNil(staff1)

	req2 := &models.CreateStaffRequest{
		FirstName:         "Another",
		LastName:          "Chief",
		NationalID:        "12345678902",
		Phone:             "+905551234568",
		ProfessionGroupID: professionGroupID,
		TitleID:           titleID,
		ClinicID:          nil,
		WorkingDays:       []models.WorkingDay{models.Monday, models.Tuesday, models.Wednesday},
	}

	staff2, err := suite.staffService.CreateStaff(req2, suite.hospitalID)
	suite.Error(err)
	suite.Nil(staff2)
}

func (suite *StaffServiceIntegrationTestSuite) TestGetStaffWithPagination() {
	professionGroupID, titleID := suite.getProfessionGroupAndTitle("Doktor", "Asistan")

	for i := 0; i < 15; i++ {
		req := &models.CreateStaffRequest{
			FirstName:         faker.Name(),
			LastName:          faker.LastName(),
			NationalID:        "1234567890" + string(rune('0'+i)),
			Phone:             "+9055512345" + string(rune('0'+i)) + string(rune('0'+i)),
			ProfessionGroupID: professionGroupID,
			TitleID:           titleID,
			ClinicID:          &suite.clinicID,
			WorkingDays:       []models.WorkingDay{models.Monday},
		}

		_, err := suite.staffService.CreateStaff(req, suite.hospitalID)
		suite.Require().NoError(err)
	}

	filter := &models.StaffFilterRequest{
		Page:  1,
		Limit: 10,
	}

	result, err := suite.staffService.GetStaff(filter, suite.hospitalID)

	suite.NoError(err)
	suite.NotNil(result)
	suite.Equal(int64(15), result.TotalCount)
	suite.Equal(1, result.Page)
	suite.Equal(10, result.Limit)
	suite.Equal(2, result.TotalPages)
	suite.Len(result.Data, 10)

	filter.Page = 2
	result, err = suite.staffService.GetStaff(filter, suite.hospitalID)

	suite.NoError(err)
	suite.Equal(2, result.Page)
	suite.Len(result.Data, 5) // remaining 5 records
}

func (suite *StaffServiceIntegrationTestSuite) TestGetStaffWithFilters() {
	doctorProfessionGroupID, doctorTitleID := suite.getProfessionGroupAndTitle("Doktor", "Asistan")
	serviceProfessionGroupID, serviceTitleID := suite.getProfessionGroupAndTitle("Hizmet Personeli", "Danışman")

	req1 := &models.CreateStaffRequest{
		FirstName:         "John",
		LastName:          "Doctor",
		NationalID:        "12345678901",
		Phone:             "+905551234567",
		ProfessionGroupID: doctorProfessionGroupID,
		TitleID:           doctorTitleID,
		ClinicID:          &suite.clinicID,
		WorkingDays:       []models.WorkingDay{models.Monday},
	}

	req2 := &models.CreateStaffRequest{
		FirstName:         "Jane",
		LastName:          "Nurse",
		NationalID:        "12345678902",
		Phone:             "+905551234568",
		ProfessionGroupID: serviceProfessionGroupID,
		TitleID:           serviceTitleID,
		ClinicID:          &suite.clinicID,
		WorkingDays:       []models.WorkingDay{models.Monday},
	}

	_, err := suite.staffService.CreateStaff(req1, suite.hospitalID)
	suite.Require().NoError(err)

	_, err = suite.staffService.CreateStaff(req2, suite.hospitalID)
	suite.Require().NoError(err)

	filter := &models.StaffFilterRequest{
		FirstName: "John",
		Page:      1,
		Limit:     10,
	}

	result, err := suite.staffService.GetStaff(filter, suite.hospitalID)
	suite.NoError(err)
	suite.Equal(int64(1), result.TotalCount)
	suite.Len(result.Data, 1)

	filter = &models.StaffFilterRequest{
		ProfessionGroupID: doctorProfessionGroupID,
		Page:              1,
		Limit:             10,
	}

	result, err = suite.staffService.GetStaff(filter, suite.hospitalID)

	suite.NoError(err)
	suite.Equal(int64(1), result.TotalCount)
	suite.Len(result.Data, 1)
}

func (suite *StaffServiceIntegrationTestSuite) TestUpdateStaff() {
	staff, err := helpers.CreateTestStaff(suite.containers.DB, suite.hospitalID, &suite.clinicID)
	suite.Require().NoError(err)

	req := &models.UpdateStaffRequest{
		FirstName: "Updated",
		LastName:  "Name",
		Phone:     "+905559999999",
	}

	updatedStaff, err := suite.staffService.UpdateStaff(staff.ID, req, suite.hospitalID)

	suite.NoError(err)
	suite.NotNil(updatedStaff)
	suite.Equal(req.FirstName, updatedStaff.FirstName)
	suite.Equal(req.LastName, updatedStaff.LastName)
	suite.Equal(req.Phone, updatedStaff.Phone)
	suite.Equal(staff.NationalID, updatedStaff.NationalID)
}

func (suite *StaffServiceIntegrationTestSuite) TestDeleteStaff() {
	staff, err := helpers.CreateTestStaff(suite.containers.DB, suite.hospitalID, &suite.clinicID)
	suite.Require().NoError(err)

	err = suite.staffService.DeleteStaff(staff.ID, suite.hospitalID)
	suite.NoError(err)

	deletedStaff, err := suite.staffService.GetStaffByID(staff.ID, suite.hospitalID)
	suite.Error(err)
	suite.Nil(deletedStaff)
	suite.Contains(err.Error(), "staff not found")
}
func TestStaffServiceIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(StaffServiceIntegrationTestSuite))
}

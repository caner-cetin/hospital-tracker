package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/caner-cetin/hospital-tracker/internal/handlers"
	"github.com/caner-cetin/hospital-tracker/internal/middleware"
	"github.com/caner-cetin/hospital-tracker/internal/models"
	"github.com/caner-cetin/hospital-tracker/internal/services"
	"github.com/caner-cetin/hospital-tracker/tests/helpers"
	"github.com/gin-gonic/gin"
	"github.com/go-faker/faker/v4"

	"github.com/stretchr/testify/suite"
)

type APITestSuite struct {
	suite.Suite
	containers *helpers.TestContainers
	router     *gin.Engine
	authToken  string
	hospitalID uint
	userID     uint
	userEmail  string
	userPhone  string
	password   string
}

func (suite *APITestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)

	ctx := context.Background()
	containers, err := helpers.SetupTestContainers(ctx)
	suite.Require().NoError(err)

	suite.containers = containers
	suite.setupRouter()
}

func (suite *APITestSuite) setupRouter() {
	r := gin.New()
	r.Use(middleware.CORS())

	api := r.Group("/api")
	handlers.SetupRoutes(api, suite.containers.DB, suite.containers.Redis, suite.containers.Config)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	suite.router = r
}

func (suite *APITestSuite) TearDownSuite() {
	ctx := context.Background()
	if suite.containers != nil {
		_ = suite.containers.Cleanup(ctx)
	}
}

func (suite *APITestSuite) SetupTest() {
	err := suite.containers.CleanDatabase()
	suite.Require().NoError(err)

	authService := services.NewAuthService(suite.containers.DB, suite.containers.Config)
	hospital, user, password, err := helpers.CreateTestHospital(suite.containers.DB, authService)
	suite.Require().NoError(err)

	token, err := helpers.GetTestJWTToken(authService, user)
	suite.Require().NoError(err)

	suite.authToken = token
	suite.hospitalID = hospital.ID
	suite.userID = user.ID
	suite.userEmail = user.Email
	suite.userPhone = user.Phone
	suite.password = password
}

func (suite *APITestSuite) makeRequest(method, url string, body interface{}, headers map[string]string) *httptest.ResponseRecorder {
	var jsonBody []byte
	var err error

	if body != nil {
		jsonBody, err = json.Marshal(body)
		suite.Require().NoError(err)
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(jsonBody))
	suite.Require().NoError(err)

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	return w
}

func (suite *APITestSuite) makeAuthenticatedRequest(method, url string, body interface{}) *httptest.ResponseRecorder {
	headers := map[string]string{
		"Authorization": "Bearer " + suite.authToken,
	}
	return suite.makeRequest(method, url, body, headers)
}

func (suite *APITestSuite) TestHealthEndpoint() {
	w := suite.makeRequest("GET", "/health", nil, nil)

	suite.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal("ok", response["status"])
}

func (suite *APITestSuite) TestHospitalRegistration() {
	registrationData := models.HospitalRegistrationRequest{
		HospitalName: "New Test Hospital",
		TaxID:        "9876543210",
		Email:        "newhospital@test.com",
		Phone:        "+905558765432",
		ProvinceID:   1,
		DistrictID:   1,
		Address:      "New Hospital Address",
		FirstName:    "Jane",
		LastName:     "Doe",
		NationalID:   "98765432101",
		UserEmail:    "jane.doe@test.com",
		UserPhone:    "+905556543210",
		Password:     faker.Password(),
	}

	w := suite.makeRequest("POST", "/api/register", registrationData, nil)

	suite.Equal(http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Contains(response, "hospital")
	suite.Contains(response, "user")
	suite.Equal("Hospital registered successfully", response["message"])
}

func (suite *APITestSuite) TestLogin() {
	loginData := models.LoginRequest{
		Identifier: suite.userEmail,
		Password:   suite.password,
	}

	w := suite.makeRequest("POST", "/api/login", loginData, nil)

	suite.Equal(http.StatusOK, w.Code)

	var response models.LoginResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.NotEmpty(response.Token)
	suite.Equal(string(models.UserTypeAuthorized), response.UserType)
}

func (suite *APITestSuite) TestLoginInvalidCredentials() {
	loginData := models.LoginRequest{
		Identifier: suite.userEmail,
		Password:   "wrongpassword",
	}

	w := suite.makeRequest("POST", "/api/login", loginData, nil)

	suite.Equal(http.StatusUnauthorized, w.Code)

	var response models.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal("Unauthorized", response.Error)
}

func (suite *APITestSuite) TestPasswordResetFlow() {
	resetRequest := models.PasswordResetRequest{
		Phone: suite.userPhone,
	}

	w := suite.makeRequest("POST", "/api/password-reset/request", resetRequest, nil)

	suite.Equal(http.StatusOK, w.Code)

	var resetResponse models.PasswordResetResponse
	err := json.Unmarshal(w.Body.Bytes(), &resetResponse)
	suite.NoError(err)
	suite.NotEmpty(resetResponse.Code)

	confirmRequest := models.PasswordResetConfirmRequest{
		Phone:           suite.userPhone,
		Code:            resetResponse.Code,
		NewPassword:     "newpassword123",
		ConfirmPassword: "newpassword123",
	}

	w = suite.makeRequest("POST", "/api/password-reset/confirm", confirmRequest, nil)
	suite.Equal(http.StatusNoContent, w.Code, fmt.Sprintf("expected 204, got %d", w.Code))
}

func (suite *APITestSuite) TestCreateUser() {
	userData := models.CreateUserRequest{
		FirstName:  "Test",
		LastName:   "User",
		NationalID: "11111111111",
		Email:      "testuser@test.com",
		Phone:      "+905557777777",
		Password:   faker.Password(),
		UserType:   models.UserTypeEmployee,
	}

	w := suite.makeAuthenticatedRequest("POST", "/api/users", userData)

	suite.Equal(http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Contains(response, "user")
	suite.Equal("User created successfully", response["message"])
}

func (suite *APITestSuite) TestCreateUserUnauthorized() {
	userData := models.CreateUserRequest{
		FirstName:  "Test",
		LastName:   "User",
		NationalID: "11111111111",
		Email:      "testuser@test.com",
		Phone:      "+905557777777",
		Password:   faker.Password(),
		UserType:   models.UserTypeEmployee,
	}

	w := suite.makeRequest("POST", "/api/users", userData, nil)

	suite.Equal(http.StatusUnauthorized, w.Code)
}

func (suite *APITestSuite) TestGetUsers() {
	w := suite.makeAuthenticatedRequest("GET", "/api/users", nil)

	suite.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Contains(response, "users")

	users := response["users"].([]interface{})
	suite.Len(users, 1) // Should have the hospital registration user
}

func (suite *APITestSuite) TestCreateClinic() {
	clinicData := models.CreateClinicRequest{
		ClinicTypeID: 1,
	}

	w := suite.makeAuthenticatedRequest("POST", "/api/clinics", clinicData)

	suite.Equal(http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Contains(response, "clinic")
	suite.Equal("Clinic created successfully", response["message"])
}

func (suite *APITestSuite) TestGetClinics() {
	clinicData := models.CreateClinicRequest{
		ClinicTypeID: 1,
	}
	w := suite.makeAuthenticatedRequest("POST", "/api/clinics", clinicData)
	suite.Require().Equal(http.StatusCreated, w.Code)

	w = suite.makeAuthenticatedRequest("GET", "/api/clinics", nil)

	suite.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Contains(response, "clinics")

	clinics := response["clinics"].([]interface{})
	suite.Len(clinics, 1)
}

func (suite *APITestSuite) TestCreateStaff() {
	// First create a clinic
	clinicData := models.CreateClinicRequest{
		ClinicTypeID: 1,
	}
	w := suite.makeAuthenticatedRequest("POST", "/api/clinics", clinicData)
	suite.Require().Equal(http.StatusCreated, w.Code)

	var clinicResponse map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &clinicResponse)
	suite.Require().NoError(err)

	clinic := clinicResponse["clinic"].(map[string]interface{})
	clinicID := uint(clinic["id"].(float64))

	staffData := models.CreateStaffRequest{
		FirstName:         "Staff",
		LastName:          "Member",
		NationalID:        "22222222222",
		Phone:             "+905558888888",
		ProfessionGroupID: 1,
		TitleID:           1,
		ClinicID:          &clinicID,
		WorkingDays:       []models.WorkingDay{models.Monday, models.Tuesday},
	}

	w = suite.makeAuthenticatedRequest("POST", "/api/staff", staffData)

	suite.Equal(http.StatusCreated, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Contains(response, "staff")
	suite.Equal("Staff created successfully", response["message"])
}

func (suite *APITestSuite) TestGetStaff() {
	w := suite.makeAuthenticatedRequest("GET", "/api/staff", nil)

	suite.Equal(http.StatusOK, w.Code)

	var response models.StaffPaginatedResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal(int64(0), response.TotalCount) // no staff initially
	suite.Equal(1, response.Page)
	suite.Equal(10, response.Limit)
}

func (suite *APITestSuite) TestGetStaffWithFilters() {
	url := "/api/staff?first_name=test&page=1&limit=5"
	w := suite.makeAuthenticatedRequest("GET", url, nil)

	suite.Equal(http.StatusOK, w.Code)

	var response models.StaffPaginatedResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal(1, response.Page)
	suite.Equal(5, response.Limit)
}

func (suite *APITestSuite) TestGetProvinces() {
	w := suite.makeRequest("GET", "/api/provinces", nil, nil)

	suite.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Contains(response, "provinces")

	provinces := response["provinces"].([]interface{})
	suite.GreaterOrEqual(len(provinces), 1)
}

func (suite *APITestSuite) TestGetDistricts() {
	w := suite.makeRequest("GET", "/api/districts", nil, nil)

	suite.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Contains(response, "districts")

	districts := response["districts"].([]interface{})
	suite.GreaterOrEqual(len(districts), 1)
}

func (suite *APITestSuite) TestGetDistrictsByProvince() {
	w := suite.makeRequest("GET", "/api/districts?province_id=1", nil, nil)

	suite.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Contains(response, "districts")
}

func (suite *APITestSuite) TestGetClinicTypes() {
	w := suite.makeRequest("GET", "/api/clinic-types", nil, nil)

	suite.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Contains(response, "clinic_types")

	clinicTypes := response["clinic_types"].([]interface{})
	suite.GreaterOrEqual(len(clinicTypes), 1)
}

func (suite *APITestSuite) TestGetProfessionGroups() {
	w := suite.makeRequest("GET", "/api/profession-groups", nil, nil)

	suite.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Contains(response, "profession_groups")

	professionGroups := response["profession_groups"].([]interface{})
	suite.GreaterOrEqual(len(professionGroups), 1)
}

func TestAPITestSuite(t *testing.T) {
	suite.Run(t, new(APITestSuite))
}

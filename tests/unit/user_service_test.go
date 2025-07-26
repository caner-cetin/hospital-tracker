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

type UserServiceTestSuite struct {
	suite.Suite
	containers     *helpers.TestContainers
	userService    *services.UserService
	authService    *services.AuthService
	hospitalID     uint
	authorizedUser *models.User
}

func (suite *UserServiceTestSuite) SetupSuite() {
	ctx := context.Background()
	containers, err := helpers.SetupTestContainers(ctx)
	suite.Require().NoError(err)

	suite.containers = containers
	suite.authService = services.NewAuthService(containers.DB, containers.Config)
	suite.userService = services.NewUserService(containers.DB, suite.authService)
}

func (suite *UserServiceTestSuite) TearDownSuite() {
	ctx := context.Background()
	if suite.containers != nil {
		_ = suite.containers.Cleanup(ctx)
	}
}

func (suite *UserServiceTestSuite) SetupTest() {
	err := suite.containers.CleanDatabase()
	suite.Require().NoError(err)

	// Create a test hospital for each test
	hospital, authorizedUser, _, err := helpers.CreateTestHospital(suite.containers.DB, suite.authService)
	suite.Require().NoError(err)
	suite.hospitalID = hospital.ID
	suite.authorizedUser = authorizedUser
}

func (suite *UserServiceTestSuite) TestCreateUser() {
	req := &models.CreateUserRequest{
		FirstName:  "Jane",
		LastName:   "Smith",
		NationalID: "98765432109",
		Email:      "jane.smith@test.com",
		Phone:      "+905551111111",
		Password:   faker.Password(),
		UserType:   models.UserTypeEmployee,
	}

	user, err := suite.userService.CreateUser(req, suite.authorizedUser.ID, suite.hospitalID)

	suite.NoError(err)
	suite.NotNil(user)
	suite.Equal(req.FirstName, user.FirstName)
	suite.Equal(req.LastName, user.LastName)
	suite.Equal(req.NationalID, user.NationalID)
	suite.Equal(req.Email, user.Email)
	suite.Equal(req.Phone, user.Phone)
	suite.Equal(req.UserType, user.UserType)
	suite.Equal(suite.hospitalID, user.HospitalID)
	suite.NotEqual(req.Password, user.Password)
}

func (suite *UserServiceTestSuite) TestCreateUserDuplicateNationalID() {
	req1 := &models.CreateUserRequest{
		FirstName:  "Jane",
		LastName:   "Smith",
		NationalID: "98765432109",
		Email:      "jane.smith@test.com",
		Phone:      "+905551111111",
		Password:   faker.Password(),
		UserType:   models.UserTypeEmployee,
	}

	user1, err := suite.userService.CreateUser(req1, suite.authorizedUser.ID, suite.hospitalID)
	suite.Require().NoError(err)
	suite.Require().NotNil(user1)

	req2 := &models.CreateUserRequest{
		FirstName:  "John",
		LastName:   "Doe",
		NationalID: "98765432109",
		Email:      "john.doe@test.com",
		Phone:      "+905552222222",
		Password:   faker.Password(),
		UserType:   models.UserTypeEmployee,
	}

	user2, err := suite.userService.CreateUser(req2, suite.authorizedUser.ID, suite.hospitalID)
	suite.Error(err)
	suite.Nil(user2)

	if appErr, ok := errors.IsAppError(err); ok {
		suite.Equal(errors.ErrCodeDuplicateNationalID, appErr.Code)
	} else {
		suite.T().Errorf("Expected AppError, got %T", err)
	}
}

func (suite *UserServiceTestSuite) TestUpdateUser() {
	user, err := helpers.CreateTestUser(suite.containers.DB, suite.authService, suite.hospitalID, models.UserTypeEmployee)
	suite.Require().NoError(err)

	req := &models.UpdateUserRequest{
		FirstName: "UpdatedName",
		LastName:  "UpdatedLastName",
		Email:     "updated@test.com",
		UserType:  models.UserTypeAuthorized,
	}

	updatedUser, err := suite.userService.UpdateUser(user.ID, req, suite.hospitalID)

	suite.NoError(err)
	suite.NotNil(updatedUser)
	suite.Equal(req.FirstName, updatedUser.FirstName)
	suite.Equal(req.LastName, updatedUser.LastName)
	suite.Equal(req.Email, updatedUser.Email)
	suite.Equal(req.UserType, updatedUser.UserType)
	suite.Equal(user.NationalID, updatedUser.NationalID)
}

func (suite *UserServiceTestSuite) TestUpdateUserDuplicateEmail() {
	user1, err := helpers.CreateTestUser(suite.containers.DB, suite.authService, suite.hospitalID, models.UserTypeEmployee)
	suite.Require().NoError(err)

	req := &models.CreateUserRequest{
		FirstName:  "John",
		LastName:   "Doe",
		NationalID: "11111111111",
		Email:      "john.doe@test.com",
		Phone:      "+905553333333",
		Password:   faker.Password(),
		UserType:   models.UserTypeEmployee,
	}

	user2, err := suite.userService.CreateUser(req, suite.authorizedUser.ID, suite.hospitalID)
	suite.Require().NoError(err)

	updateReq := &models.UpdateUserRequest{
		Email: user1.Email,
	}

	updatedUser, err := suite.userService.UpdateUser(user2.ID, updateReq, suite.hospitalID)
	suite.Error(err)
	suite.Nil(updatedUser)
	suite.Contains(err.Error(), "email already exists")
}

func (suite *UserServiceTestSuite) TestDeleteUser() {
	user, err := helpers.CreateTestUser(suite.containers.DB, suite.authService, suite.hospitalID, models.UserTypeEmployee)
	suite.Require().NoError(err)

	err = suite.userService.DeleteUser(user.ID, suite.hospitalID)
	suite.NoError(err)

	deletedUser, err := suite.userService.GetUser(user.ID, suite.hospitalID)
	suite.Error(err)
	suite.Nil(deletedUser)

	if appErr, ok := errors.IsAppError(err); ok {
		suite.Equal(errors.ErrCodeUserNotFound, appErr.Code)
	} else {
		suite.T().Errorf("expected AppError, got %T", err)
	}
}

func (suite *UserServiceTestSuite) TestGetUsers() {
	user1, err := helpers.CreateTestUser(suite.containers.DB, suite.authService, suite.hospitalID, models.UserTypeEmployee)
	suite.Require().NoError(err)

	req := &models.CreateUserRequest{
		FirstName:  "John",
		LastName:   "Doe",
		NationalID: "11111111111",
		Email:      "john.doe@test.com",
		Phone:      "+905553333333",
		Password:   faker.Password(),
		UserType:   models.UserTypeAuthorized,
	}

	user2, err := suite.userService.CreateUser(req, suite.authorizedUser.ID, suite.hospitalID)
	suite.Require().NoError(err)

	users, err := suite.userService.GetUsers(suite.hospitalID)

	suite.NoError(err)
	// 2 created + 1 application
	suite.Len(users, 3)

	userIDs := make(map[uint]bool)
	for _, u := range users {
		userIDs[u.ID] = true
	}
	suite.True(userIDs[user1.ID])
	suite.True(userIDs[user2.ID])
}

func (suite *UserServiceTestSuite) TestGetUser() {
	user, err := helpers.CreateTestUser(suite.containers.DB, suite.authService, suite.hospitalID, models.UserTypeEmployee)
	suite.Require().NoError(err)

	retrievedUser, err := suite.userService.GetUser(user.ID, suite.hospitalID)

	suite.NoError(err)
	suite.NotNil(retrievedUser)
	suite.Equal(user.ID, retrievedUser.ID)
	suite.Equal(user.FirstName, retrievedUser.FirstName)
	suite.Equal(user.LastName, retrievedUser.LastName)
	suite.Equal(user.Email, retrievedUser.Email)
}

func (suite *UserServiceTestSuite) TestGetUserNotFound() {
	retrievedUser, err := suite.userService.GetUser(999, suite.hospitalID)

	suite.Error(err)
	suite.Nil(retrievedUser)

	if appErr, ok := errors.IsAppError(err); ok {
		suite.Equal(errors.ErrCodeUserNotFound, appErr.Code)
	} else {
		suite.T().Errorf("Expected AppError, got %T", err)
	}
}

func TestUserServiceTestSuite(t *testing.T) {
	suite.Run(t, new(UserServiceTestSuite))
}

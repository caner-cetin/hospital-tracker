package unit

import (
	"context"
	"testing"

	"github.com/caner-cetin/hospital-tracker/internal/errors"
	"github.com/caner-cetin/hospital-tracker/internal/services"
	"github.com/caner-cetin/hospital-tracker/tests/helpers"
	"github.com/stretchr/testify/suite"
)

type AuthServiceTestSuite struct {
	suite.Suite
	containers  *helpers.TestContainers
	authService *services.AuthService
}

func (suite *AuthServiceTestSuite) SetupSuite() {
	ctx := context.Background()
	containers, err := helpers.SetupTestContainers(ctx)
	suite.Require().NoError(err)

	suite.containers = containers
	suite.authService = services.NewAuthService(containers.DB, containers.Config)
}

func (suite *AuthServiceTestSuite) TearDownSuite() {
	ctx := context.Background()
	if suite.containers != nil {
		_ = suite.containers.Cleanup(ctx)
	}
}

func (suite *AuthServiceTestSuite) SetupTest() {
	err := suite.containers.CleanDatabase()
	suite.Require().NoError(err)
}

func (suite *AuthServiceTestSuite) TestHashPassword() {
	password := "testpassword123"

	hashedPassword, err := suite.authService.HashPassword(password)

	suite.NoError(err)
	suite.NotEmpty(hashedPassword)
	suite.NotEqual(password, hashedPassword)
}

func (suite *AuthServiceTestSuite) TestCheckPassword() {
	password := "testpassword123"
	hashedPassword, err := suite.authService.HashPassword(password)
	suite.Require().NoError(err)

	// Valid password
	err = suite.authService.CheckPassword(hashedPassword, password)
	suite.NoError(err)

	// Invalid password
	err = suite.authService.CheckPassword(hashedPassword, "wrongpassword")
	suite.Error(err)
}

func (suite *AuthServiceTestSuite) TestGenerateToken() {
	hospital, user, _, err := helpers.CreateTestHospital(suite.containers.DB, suite.authService)
	suite.Require().NoError(err)
	suite.Require().NotNil(hospital)
	suite.Require().NotNil(user)

	token, err := suite.authService.GenerateToken(user)

	suite.NoError(err)
	suite.NotEmpty(token)
}

func (suite *AuthServiceTestSuite) TestValidateToken() {
	hospital, user, _, err := helpers.CreateTestHospital(suite.containers.DB, suite.authService)
	suite.Require().NoError(err)
	suite.Require().NotNil(hospital)
	suite.Require().NotNil(user)

	token, err := suite.authService.GenerateToken(user)
	suite.Require().NoError(err)

	claims, err := suite.authService.ValidateToken(token)

	suite.NoError(err)
	suite.NotNil(claims)
	suite.Equal(user.ID, claims.UserID)
	suite.Equal(user.HospitalID, claims.HospitalID)
	suite.Equal(user.UserType, claims.UserType)
}

func (suite *AuthServiceTestSuite) TestValidateInvalidToken() {
	invalidToken := "invalid.token.here"

	claims, err := suite.authService.ValidateToken(invalidToken)

	suite.Error(err)
	suite.Nil(claims)
}

func (suite *AuthServiceTestSuite) TestLogin() {
	hospital, user, password, err := helpers.CreateTestHospital(suite.containers.DB, suite.authService)
	suite.Require().NoError(err)
	suite.Require().NotNil(hospital)
	suite.Require().NotNil(user)

	loginUser, token, err := suite.authService.Login(user.Email, password)
	suite.NoError(err)
	suite.NotNil(loginUser)
	suite.NotEmpty(token)
	suite.Equal(user.ID, loginUser.ID)

	loginUser, token, err = suite.authService.Login(user.Phone, password)
	suite.NoError(err)
	suite.NotNil(loginUser)
	suite.NotEmpty(token)
	suite.Equal(user.ID, loginUser.ID)
}

func (suite *AuthServiceTestSuite) TestLoginInvalidCredentials() {
	hospital, user, password, err := helpers.CreateTestHospital(suite.containers.DB, suite.authService)
	suite.Require().NoError(err)
	suite.Require().NotNil(hospital)
	suite.Require().NotNil(user)

	loginUser, token, err := suite.authService.Login(user.Email, password + "wrong")
	suite.Error(err)
	suite.Nil(loginUser)
	suite.Empty(token)

	if appErr, ok := errors.IsAppError(err); ok {
		suite.Equal(errors.ErrCodeInvalidCredentials, appErr.Code)
	} else {
		suite.T().Errorf("Expected AppError, got %T", err)
	}

	loginUser, token, err = suite.authService.Login("nonexistent@test.com", password)
	suite.Error(err)
	suite.Nil(loginUser)
	suite.Empty(token)

	if appErr, ok := errors.IsAppError(err); ok {
		suite.Equal(errors.ErrCodeInvalidCredentials, appErr.Code)
	} else {
		suite.T().Errorf("Expected AppError, got %T", err)
	}
}

func TestAuthServiceTestSuite(t *testing.T) {
	suite.Run(t, new(AuthServiceTestSuite))
}

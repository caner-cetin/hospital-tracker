package unit

import (
	"context"
	"testing"
	"time"

	"github.com/caner-cetin/hospital-tracker/internal/models"
	"github.com/caner-cetin/hospital-tracker/internal/services"
	"github.com/caner-cetin/hospital-tracker/tests/helpers"
	"github.com/stretchr/testify/suite"
)

type PasswordResetServiceTestSuite struct {
	suite.Suite
	containers           *helpers.TestContainers
	passwordResetService *services.PasswordResetService
	authService          *services.AuthService
}

func (suite *PasswordResetServiceTestSuite) SetupSuite() {
	ctx := context.Background()
	containers, err := helpers.SetupTestContainers(ctx)
	suite.Require().NoError(err)

	suite.containers = containers
	suite.authService = services.NewAuthService(containers.DB, containers.Config)
	suite.passwordResetService = services.NewPasswordResetService(containers.DB, suite.authService)
}

func (suite *PasswordResetServiceTestSuite) TearDownSuite() {
	ctx := context.Background()
	if suite.containers != nil {
		_ = suite.containers.Cleanup(ctx)
	}
}

func (suite *PasswordResetServiceTestSuite) SetupTest() {
	err := suite.containers.CleanDatabase()
	suite.Require().NoError(err)
}

func (suite *PasswordResetServiceTestSuite) TestRequestPasswordReset() {
	_, user, _, err := helpers.CreateTestHospital(suite.containers.DB, suite.authService)
	suite.Require().NoError(err)

	code, err := suite.passwordResetService.RequestPasswordReset(user.Phone)

	suite.NoError(err)
	suite.NotEmpty(code)
	suite.Len(code, 6)

	var passwordReset models.PasswordReset
	err = suite.containers.DB.Where("phone = ? AND code = ?", user.Phone, code).First(&passwordReset).Error
	suite.NoError(err)
	suite.Equal(user.Phone, passwordReset.Phone)
	suite.Equal(code, passwordReset.Code)
	suite.False(passwordReset.Used)
	suite.True(passwordReset.ExpiresAt.After(time.Now()))
}

func (suite *PasswordResetServiceTestSuite) TestRequestPasswordResetNonExistentPhone() {
	code, err := suite.passwordResetService.RequestPasswordReset("+905559999999")

	suite.Error(err)
	suite.Empty(code)
	suite.Contains(err.Error(), "phone number not found")
}

func (suite *PasswordResetServiceTestSuite) TestRequestPasswordResetReplacesOldCodes() {
	_, user, _, err := helpers.CreateTestHospital(suite.containers.DB, suite.authService)
	suite.Require().NoError(err)

	code1, err := suite.passwordResetService.RequestPasswordReset(user.Phone)
	suite.Require().NoError(err)

	code2, err := suite.passwordResetService.RequestPasswordReset(user.Phone)
	suite.Require().NoError(err)

	suite.NotEqual(code1, code2)

	var count int64
	suite.containers.DB.Model(&models.PasswordReset{}).Where("phone = ? AND code = ? AND used = false", user.Phone, code1).Count(&count)
	suite.Equal(int64(0), count)

	suite.containers.DB.Model(&models.PasswordReset{}).Where("phone = ? AND code = ? AND used = false", user.Phone, code2).Count(&count)
	suite.Equal(int64(1), count)
}

func (suite *PasswordResetServiceTestSuite) TestResetPassword() {
	_, user, _, err := helpers.CreateTestHospital(suite.containers.DB, suite.authService)
	suite.Require().NoError(err)

	code, err := suite.passwordResetService.RequestPasswordReset(user.Phone)
	suite.Require().NoError(err)

	newPassword := "newpassword123"
	err = suite.passwordResetService.ResetPassword(user.Phone, code, newPassword, newPassword)

	suite.NoError(err)

	var updatedUser models.User
	err = suite.containers.DB.First(&updatedUser, user.ID).Error
	suite.Require().NoError(err)

	err = suite.authService.CheckPassword(updatedUser.Password, newPassword)
	suite.NoError(err)

	var passwordReset models.PasswordReset
	err = suite.containers.DB.Where("phone = ? AND code = ?", user.Phone, code).First(&passwordReset).Error
	suite.NoError(err)
	suite.True(passwordReset.Used)
}

func (suite *PasswordResetServiceTestSuite) TestResetPasswordMismatch() {
	_, user, _, err := helpers.CreateTestHospital(suite.containers.DB, suite.authService)
	suite.Require().NoError(err)

	code, err := suite.passwordResetService.RequestPasswordReset(user.Phone)
	suite.Require().NoError(err)

	err = suite.passwordResetService.ResetPassword(user.Phone, code, "password1", "password2")

	suite.Error(err)
}

func (suite *PasswordResetServiceTestSuite) TestResetPasswordInvalidCode() {
	_, user, _, err := helpers.CreateTestHospital(suite.containers.DB, suite.authService)
	suite.Require().NoError(err)

	err = suite.passwordResetService.ResetPassword(user.Phone, "invalid", "newpassword", "newpassword")

	suite.Error(err)
}

func (suite *PasswordResetServiceTestSuite) TestResetPasswordExpiredCode() {
	_, user, _, err := helpers.CreateTestHospital(suite.containers.DB, suite.authService)
	suite.Require().NoError(err)

	code, err := suite.passwordResetService.RequestPasswordReset(user.Phone)
	suite.Require().NoError(err)

	expiredTime := time.Now().Add(-1 * time.Hour)
	suite.containers.DB.Model(&models.PasswordReset{}).Where("phone = ? AND code = ?", user.Phone, code).Update("expires_at", expiredTime)

	err = suite.passwordResetService.ResetPassword(user.Phone, code, "newpassword", "newpassword")

	suite.Error(err)
}

func (suite *PasswordResetServiceTestSuite) TestResetPasswordUsedCode() {
	_, user, _, err := helpers.CreateTestHospital(suite.containers.DB, suite.authService)
	suite.Require().NoError(err)

	code, err := suite.passwordResetService.RequestPasswordReset(user.Phone)
	suite.Require().NoError(err)

	err = suite.passwordResetService.ResetPassword(user.Phone, code, "newpassword", "newpassword")
	suite.Require().NoError(err)

	err = suite.passwordResetService.ResetPassword(user.Phone, code, "anotherpassword", "anotherpassword")

	suite.Error(err)
	suite.Contains(err.Error(), "invalid or expired reset code")
}

func TestPasswordResetServiceTestSuite(t *testing.T) {
	suite.Run(t, new(PasswordResetServiceTestSuite))
}
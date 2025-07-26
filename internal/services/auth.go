package services

import (
	"errors"
	"time"

	"github.com/caner-cetin/hospital-tracker/internal/config"
	apperrors "github.com/caner-cetin/hospital-tracker/internal/errors"
	"github.com/caner-cetin/hospital-tracker/internal/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	db  *gorm.DB
	cfg *config.Config
}

func NewAuthService(db *gorm.DB, cfg *config.Config) *AuthService {
	return &AuthService{
		db:  db,
		cfg: cfg,
	}
}

type Claims struct {
	UserID     uint            `json:"user_id"`
	HospitalID uint            `json:"hospital_id"`
	UserType   models.UserType `json:"user_type"`
	jwt.RegisteredClaims
}

func (s *AuthService) HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", apperrors.NewInternalError("failed to hash password", err)
	}
	return string(hashedPassword), nil
}

func (s *AuthService) CheckPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

func (s *AuthService) GenerateToken(user *models.User) (string, error) {
	claims := Claims{
		UserID:     user.ID,
		HospitalID: user.HospitalID,
		UserType:   user.UserType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(s.cfg.JWT.ExpireHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.cfg.JWT.Secret))
	if err != nil {
		return "", apperrors.NewInternalError("failed to generate token", err)
	}
	return tokenString, nil
}

func (s *AuthService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(_ *jwt.Token) (interface{}, error) {
		return []byte(s.cfg.JWT.Secret), nil
	})

	if err != nil {
		if err.Error() == "token is expired" {
			return nil, apperrors.NewExpiredTokenError()
		}
		return nil, apperrors.NewInvalidTokenError()
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, apperrors.NewInvalidTokenError()
}

func (s *AuthService) Login(identifier, password string) (*models.User, string, error) {
	log.Info().Str("identifier", identifier).Msg("Login attempt")
	
	var user models.User

	err := s.db.Where("email = ? OR phone = ?", identifier, identifier).
		Preload("Hospital").First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Warn().Str("identifier", identifier).Msg("Login failed: user not found")
			return nil, "", apperrors.NewInvalidCredentialsError()
		}
		log.Error().Err(err).Str("identifier", identifier).Msg("Database error during login")
		return nil, "", apperrors.NewDatabaseError("user lookup", err)
	}

	if err := s.CheckPassword(user.Password, password); err != nil {
		log.Warn().Uint("user_id", user.ID).Str("identifier", identifier).Msg("Login failed: invalid password")
		return nil, "", apperrors.NewInvalidCredentialsError()
	}

	token, err := s.GenerateToken(&user)
	if err != nil {
		log.Error().Err(err).Uint("user_id", user.ID).Msg("Failed to generate token")
		return nil, "", err
	}

	log.Info().
		Uint("user_id", user.ID).
		Uint("hospital_id", user.HospitalID).
		Str("user_type", string(user.UserType)).
		Msg("Login successful")

	return &user, token, nil
}

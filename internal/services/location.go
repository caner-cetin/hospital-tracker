package services

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/caner-cetin/hospital-tracker/internal/models"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type LocationService struct {
	db          *gorm.DB
	redisClient *redis.Client
}

func NewLocationService(db *gorm.DB, redisClient *redis.Client) *LocationService {
	return &LocationService{
		db:          db,
		redisClient: redisClient,
	}
}

func (s *LocationService) GetProvinces() ([]models.Province, error) {
	ctx := context.Background()
	cacheKey := "provinces"

	cachedData, err := s.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		var provinces []models.Province
		if err := json.Unmarshal([]byte(cachedData), &provinces); err == nil {
			return provinces, nil
		}
	}

	var provinces []models.Province
	err = s.db.Find(&provinces).Error
	if err != nil {
		return nil, err
	}

	if data, err := json.Marshal(provinces); err == nil {
		s.redisClient.Set(ctx, cacheKey, data, 24*time.Hour)
	}

	return provinces, nil
}

func (s *LocationService) GetAllDistricts() ([]models.District, error) {
	ctx := context.Background()
	cacheKey := "districts:all"

	cachedData, err := s.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		var districts []models.District
		if err := json.Unmarshal([]byte(cachedData), &districts); err == nil {
			return districts, nil
		}
	}

	var districts []models.District
	err = s.db.Preload("Province").Find(&districts).Error
	if err != nil {
		return nil, err
	}

	if data, err := json.Marshal(districts); err == nil {
		s.redisClient.Set(ctx, cacheKey, data, 24*time.Hour)
	}

	return districts, nil
}

func (s *LocationService) GetDistrictsByProvince(provinceIDStr string) ([]models.District, error) {
	provinceID, err := strconv.ParseUint(provinceIDStr, 10, 32)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	cacheKey := "districts:province:" + provinceIDStr

	cachedData, err := s.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		var districts []models.District
		if err := json.Unmarshal([]byte(cachedData), &districts); err == nil {
			return districts, nil
		}
	}

	var districts []models.District
	err = s.db.Where("province_id = ?", uint(provinceID)).Preload("Province").Find(&districts).Error
	if err != nil {
		return nil, err
	}

	if data, err := json.Marshal(districts); err == nil {
		s.redisClient.Set(ctx, cacheKey, data, 24*time.Hour)
	}

	return districts, err
}

package services

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"time"

	"github.com/caner-cetin/hospital-tracker/internal/models"
	pkgerrors "github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type StaffService struct {
	db          *gorm.DB
	redisClient *redis.Client
}

func NewStaffService(db *gorm.DB, redisClient *redis.Client) *StaffService {
	return &StaffService{
		db:          db,
		redisClient: redisClient,
	}
}

func (s *StaffService) CreateStaff(req *models.CreateStaffRequest, hospitalID uint) (*models.Staff, error) {
	if err := s.validateStaffUniqueness(req.NationalID, req.Phone, 0); err != nil {
		return nil, err
	}

	if err := s.validateTitleProfessionGroup(req.TitleID, req.ProfessionGroupID); err != nil {
		return nil, err
	}

	if req.ClinicID != nil {
		if err := s.validateClinicBelongsToHospital(*req.ClinicID, hospitalID); err != nil {
			return nil, err
		}
	}

	var title models.Title
	if err := s.db.Preload("ProfessionGroup").First(&title, req.TitleID).Error; err != nil {
		return nil, errors.New("title not found")
	}

	if title.ProfessionGroup.Name == "İdari Personel" && title.Name == "Başhekim" {
		var count int64
		s.db.Model(&models.Staff{}).
			Joins("JOIN titles ON staffs.title_id = titles.id").
			Where("titles.name = ? AND staffs.hospital_id = ?", "Başhekim", hospitalID).
			Count(&count)
		if count > 0 {
			return nil, errors.New("hospital can only have one chief physician (Başhekim)")
		}
	}

	workingDaysJSON, err := json.Marshal(req.WorkingDays)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "failed to marshal working days")
	}

	staff := &models.Staff{
		FirstName:         req.FirstName,
		LastName:          req.LastName,
		NationalID:        req.NationalID,
		Phone:             req.Phone,
		ProfessionGroupID: req.ProfessionGroupID,
		TitleID:           req.TitleID,
		HospitalID:        hospitalID,
		ClinicID:          req.ClinicID,
		WorkingDays:       string(workingDaysJSON),
	}

	if err := s.db.Create(staff).Error; err != nil {
		return nil, err
	}

	if err := s.db.Preload("ProfessionGroup").Preload("Title").
		Preload("Hospital").Preload("Clinic.ClinicType").
		First(staff, staff.ID).Error; err != nil {
		return nil, err
	}

	return staff, nil
}

func (s *StaffService) UpdateStaff(staffID uint, req *models.UpdateStaffRequest, hospitalID uint) (*models.Staff, error) {
	var staff models.Staff
	if err := s.db.Where("id = ? AND hospital_id = ?", staffID, hospitalID).First(&staff).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("staff not found")
		}
		return nil, err
	}

	if req.NationalID != "" && req.NationalID != staff.NationalID {
		if err := s.validateStaffUniqueness(req.NationalID, "", staffID); err != nil {
			return nil, err
		}
		staff.NationalID = req.NationalID
	}

	if req.Phone != "" && req.Phone != staff.Phone {
		if err := s.validateStaffUniqueness("", req.Phone, staffID); err != nil {
			return nil, err
		}
		staff.Phone = req.Phone
	}

	if req.TitleID != 0 && req.ProfessionGroupID != 0 {
		if err := s.validateTitleProfessionGroup(req.TitleID, req.ProfessionGroupID); err != nil {
			return nil, err
		}

		var title models.Title
		if err := s.db.Preload("ProfessionGroup").First(&title, req.TitleID).Error; err != nil {
			return nil, errors.New("title not found")
		}

		if title.ProfessionGroup.Name == "İdari Personel" && title.Name == "Başhekim" {
			var count int64
			s.db.Model(&models.Staff{}).
				Joins("JOIN titles ON staffs.title_id = titles.id").
				Where("titles.name = ? AND staffs.hospital_id = ? AND staffs.id != ?", "Başhekim", hospitalID, staffID).
				Count(&count)
			if count > 0 {
				return nil, errors.New("hospital can only have one chief physician (Başhekim)")
			}
		}

		staff.TitleID = req.TitleID
		staff.ProfessionGroupID = req.ProfessionGroupID
	}

	if req.ClinicID != nil {
		if err := s.validateClinicBelongsToHospital(*req.ClinicID, hospitalID); err != nil {
			return nil, err
		}
		staff.ClinicID = req.ClinicID
	}

	if req.FirstName != "" {
		staff.FirstName = req.FirstName
	}

	if req.LastName != "" {
		staff.LastName = req.LastName
	}

	if len(req.WorkingDays) > 0 {
		workingDaysJSON, err := json.Marshal(req.WorkingDays)
		if err != nil {
			return nil, pkgerrors.Wrap(err, "failed to marshal working days")
		}
		staff.WorkingDays = string(workingDaysJSON)
	}

	if err := s.db.Save(&staff).Error; err != nil {
		return nil, err
	}

	if err := s.db.Preload("ProfessionGroup").Preload("Title").
		Preload("Hospital").Preload("Clinic.ClinicType").
		First(&staff, staff.ID).Error; err != nil {
		return nil, err
	}

	return &staff, nil
}

func (s *StaffService) DeleteStaff(staffID uint, hospitalID uint) error {
	var staff models.Staff
	if err := s.db.Where("id = ? AND hospital_id = ?", staffID, hospitalID).First(&staff).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("staff not found")
		}
		return err
	}

	return s.db.Delete(&staff).Error
}

func (s *StaffService) GetStaff(filter *models.StaffFilterRequest, hospitalID uint) (*models.StaffPaginatedResponse, error) {
	query := s.db.Model(&models.Staff{}).Where("hospital_id = ?", hospitalID)

	if filter.FirstName != "" {
		query = query.Where("first_name ILIKE ?", "%"+filter.FirstName+"%")
	}

	if filter.LastName != "" {
		query = query.Where("last_name ILIKE ?", "%"+filter.LastName+"%")
	}

	if filter.NationalID != "" {
		query = query.Where("national_id ILIKE ?", "%"+filter.NationalID+"%")
	}

	if filter.ProfessionGroupID != 0 {
		query = query.Where("profession_group_id = ?", filter.ProfessionGroupID)
	}

	if filter.TitleID != 0 {
		query = query.Where("title_id = ?", filter.TitleID)
	}

	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, err
	}

	if filter.Limit <= 0 {
		filter.Limit = 10
	}
	if filter.Page <= 0 {
		filter.Page = 1
	}

	offset := (filter.Page - 1) * filter.Limit
	totalPages := int(math.Ceil(float64(totalCount) / float64(filter.Limit)))

	var staff []models.Staff
	err := query.Preload("ProfessionGroup").Preload("Title").
		Preload("Hospital").Preload("Clinic.ClinicType").
		Offset(offset).Limit(filter.Limit).
		Find(&staff).Error
	if err != nil {
		return nil, err
	}

	return &models.StaffPaginatedResponse{
		Data: staff,
		BasePagination: models.BasePagination{
			TotalCount: totalCount,
			Page:       filter.Page,
			Limit:      filter.Limit,
			TotalPages: totalPages,
		},
	}, nil
}

func (s *StaffService) GetStaffByID(staffID uint, hospitalID uint) (*models.Staff, error) {
	var staff models.Staff
	err := s.db.Where("id = ? AND hospital_id = ?", staffID, hospitalID).
		Preload("ProfessionGroup").Preload("Title").
		Preload("Hospital").Preload("Clinic.ClinicType").
		First(&staff).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("staff not found")
		}
		return nil, err
	}

	return &staff, nil
}

func (s *StaffService) GetProfessionGroups() ([]models.ProfessionGroupResponse, error) {
	ctx := context.Background()
	cacheKey := "profession_groups"

	cachedData, err := s.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		var professionGroups []models.ProfessionGroupResponse
		if err := json.Unmarshal([]byte(cachedData), &professionGroups); err == nil {
			return professionGroups, nil
		}
	}

	var professionGroups []models.ProfessionGroup
	err = s.db.Preload("Titles").Find(&professionGroups).Error
	if err != nil {
		return nil, err
	}

	// Convert to response DTOs
	response := make([]models.ProfessionGroupResponse, 0, len(professionGroups))
	for _, pg := range professionGroups {
		var titles []models.TitleResponse
		for _, title := range pg.Titles {
			titles = append(titles, models.TitleResponse{
				ID:                title.ID,
				Name:              title.Name,
				ProfessionGroupID: title.ProfessionGroupID,
				CreatedAt:         title.CreatedAt,
				UpdatedAt:         title.UpdatedAt,
			})
		}

		response = append(response, models.ProfessionGroupResponse{
			ID:        pg.ID,
			Name:      pg.Name,
			Titles:    titles,
			CreatedAt: pg.CreatedAt,
			UpdatedAt: pg.UpdatedAt,
		})
	}

	if data, err := json.Marshal(response); err == nil {
		s.redisClient.Set(ctx, cacheKey, data, 24*time.Hour)
	}

	return response, nil
}

func (s *StaffService) validateStaffUniqueness(nationalID, phone string, excludeID uint) error {
	var count int64

	if nationalID != "" {
		query := s.db.Model(&models.Staff{}).Where("national_id = ?", nationalID)
		if excludeID > 0 {
			query = query.Where("id != ?", excludeID)
		}
		if err := query.Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return errors.New("national ID already exists")
		}
	}

	if phone != "" {
		query := s.db.Model(&models.Staff{}).Where("phone = ?", phone)
		if excludeID > 0 {
			query = query.Where("id != ?", excludeID)
		}
		if err := query.Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return errors.New("phone already exists")
		}
	}

	return nil
}

func (s *StaffService) validateTitleProfessionGroup(titleID, professionGroupID uint) error {
	var title models.Title
	err := s.db.Where("id = ? AND profession_group_id = ?", titleID, professionGroupID).First(&title).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("title does not belong to the specified profession group")
		}
		return err
	}
	return nil
}

func (s *StaffService) validateClinicBelongsToHospital(clinicID, hospitalID uint) error {
	var clinic models.Clinic
	err := s.db.Where("id = ? AND hospital_id = ?", clinicID, hospitalID).First(&clinic).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("clinic not found or does not belong to hospital")
		}
		return err
	}
	return nil
}

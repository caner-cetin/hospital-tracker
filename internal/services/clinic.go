package services

import (
	"errors"

	"github.com/caner-cetin/hospital-tracker/internal/models"
	"gorm.io/gorm"
)

type ClinicService struct {
	db *gorm.DB
}

func NewClinicService(db *gorm.DB) *ClinicService {
	return &ClinicService{
		db: db,
	}
}

func (s *ClinicService) CreateClinic(req *models.CreateClinicRequest, hospitalID uint) (*models.Clinic, error) {
	var clinicType models.ClinicType
	if err := s.db.First(&clinicType, req.ClinicTypeID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("clinic type not found")
		}
		return nil, err
	}

	var existingClinic models.Clinic
	err := s.db.Where("hospital_id = ? AND clinic_type_id = ?", hospitalID, req.ClinicTypeID).First(&existingClinic).Error
	if err == nil {
		return nil, errors.New("clinic type already exists for this hospital")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	clinic := &models.Clinic{
		HospitalID:   hospitalID,
		ClinicTypeID: req.ClinicTypeID,
	}

	if err := s.db.Create(clinic).Error; err != nil {
		return nil, err
	}

	if err := s.db.Preload("Hospital").Preload("ClinicType").First(clinic, clinic.ID).Error; err != nil {
		return nil, err
	}

	return clinic, nil
}

func (s *ClinicService) GetClinics(hospitalID uint) ([]models.ClinicSummary, error) {
	var clinics []models.Clinic
	err := s.db.Where("hospital_id = ?", hospitalID).
		Preload("ClinicType").
		Find(&clinics).Error
	if err != nil {
		return nil, err
	}

	summaries := make([]models.ClinicSummary, 0, len(clinics))
	for _, clinic := range clinics {
		summary := models.ClinicSummary{
			ID:         clinic.ID,
			ClinicType: clinic.ClinicType,
		}

		var totalStaff int64
		s.db.Model(&models.Staff{}).Where("clinic_id = ? AND hospital_id = ?", clinic.ID, hospitalID).Count(&totalStaff)
		summary.TotalStaff = totalStaff

		var professionCounts []struct {
			ProfessionGroupName string `json:"profession_group_name"`
			Count               int64  `json:"count"`
		}

		s.db.Raw(`
			SELECT pg.name as profession_group_name, COUNT(*) as count
			FROM staffs s
			JOIN profession_groups pg ON s.profession_group_id = pg.id
			WHERE s.clinic_id = ? AND s.hospital_id = ?
			GROUP BY pg.id, pg.name
		`, clinic.ID, hospitalID).Scan(&professionCounts)

		for _, pc := range professionCounts {
			summary.StaffByProfession = append(summary.StaffByProfession, models.StaffProfessionSummary{
				ProfessionGroup: pc.ProfessionGroupName,
				Count:           pc.Count,
			})
		}

		summaries = append(summaries, summary)
	}

	return summaries, nil
}

func (s *ClinicService) GetClinicTypes() ([]models.ClinicType, error) {
	var clinicTypes []models.ClinicType
	err := s.db.Find(&clinicTypes).Error
	return clinicTypes, err
}

func (s *ClinicService) DeleteClinic(clinicID uint, hospitalID uint) error {
	var clinic models.Clinic
	if err := s.db.Where("id = ? AND hospital_id = ?", clinicID, hospitalID).First(&clinic).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("clinic not found")
		}
		return err
	}

	var staffCount int64
	s.db.Model(&models.Staff{}).Where("clinic_id = ?", clinicID).Count(&staffCount)
	if staffCount > 0 {
		return errors.New("cannot delete clinic with assigned staff")
	}

	return s.db.Delete(&clinic).Error
}

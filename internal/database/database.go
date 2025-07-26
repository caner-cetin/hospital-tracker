package database

import (
	"fmt"

	"github.com/caner-cetin/hospital-tracker/internal/config"
	"github.com/caner-cetin/hospital-tracker/internal/models"
	"github.com/rs/zerolog/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Initialize(cfg config.DatabaseConfig) (*gorm.DB, error) {
	log.Info().Str("host", cfg.Host).Str("database", cfg.Name).Msg("Connecting to database")
	
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		cfg.Host, cfg.User, cfg.Password, cfg.Name, cfg.Port, cfg.SSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Error().Err(err).Msg("Failed to connect to database")
		return nil, err
	}

	log.Info().Msg("Database connection established")

	log.Info().Msg("Running database migrations")
	err = migrate(db)
	if err != nil {
		log.Error().Err(err).Msg("Database migration failed")
		return nil, err
	}

	log.Info().Msg("Seeding database")
	err = seedData(db)
	if err != nil {
		log.Error().Err(err).Msg("Database seeding failed")
		return nil, err
	}

	return db, nil
}

func migrate(db *gorm.DB) error {
	// migrate base tables without foreign key dependencies
	err := db.AutoMigrate(
		&models.Province{},
		&models.District{},
		&models.ProfessionGroup{},
		&models.Title{},
		&models.ClinicType{},
		&models.PasswordReset{},
	)
	if err != nil {
		return err
	}

	// migrate tables with dependencies
	err = db.AutoMigrate(
		&models.Hospital{},
		&models.User{},
	)
	if err != nil {
		return err
	}

	// migrate tables with circular dependencies
	err = db.AutoMigrate(
		&models.Clinic{},
		&models.Staff{},
	)
	return err
}

func seedData(db *gorm.DB) error {
	var count int64

	db.Model(&models.Province{}).Count(&count)
	if count > 0 {
		return nil
	}

	provinces := []models.Province{
		{Name: "İstanbul"},
		{Name: "Ankara"},
		{Name: "İzmir"},
	}

	if err := db.Create(&provinces).Error; err != nil {
		return err
	}

	districts := []models.District{
		{Name: "Kadıköy", ProvinceID: 1},
		{Name: "Beşiktaş", ProvinceID: 1},
		{Name: "Çankaya", ProvinceID: 2},
		{Name: "Keçiören", ProvinceID: 2},
		{Name: "Konak", ProvinceID: 3},
		{Name: "Bornova", ProvinceID: 3},
	}

	if err := db.Create(&districts).Error; err != nil {
		return err
	}

	professionGroups := []models.ProfessionGroup{
		{Name: "Doktor"},
		{Name: "İdari Personel"},
		{Name: "Hizmet Personeli"},
	}

	if err := db.Create(&professionGroups).Error; err != nil {
		return err
	}

	titles := []models.Title{
		{Name: "Asistan", ProfessionGroupID: 1},
		{Name: "Uzman", ProfessionGroupID: 1},
		{Name: "Başhekim", ProfessionGroupID: 2},
		{Name: "Müdür", ProfessionGroupID: 2},
		{Name: "Danışman", ProfessionGroupID: 3},
		{Name: "Temizlik", ProfessionGroupID: 3},
		{Name: "Güvenlik", ProfessionGroupID: 3},
	}

	if err := db.Create(&titles).Error; err != nil {
		return err
	}

	clinicTypes := []models.ClinicType{
		{Name: "Dahiliye"},
		{Name: "Kardiyoloji"},
		{Name: "Nöroloji"},
		{Name: "Ortopedi"},
		{Name: "Göz Hastalıkları"},
		{Name: "Kulak Burun Boğaz"},
	}

	return db.Create(&clinicTypes).Error
}

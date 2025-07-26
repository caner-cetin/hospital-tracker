package models

import (
	"time"

	"gorm.io/gorm"
)

type DeletedAt *time.Time

type Province struct {
	ID        uint       `json:"id" gorm:"primaryKey"`
	Name      string     `json:"name" gorm:"not null;unique"`
	Districts []District `json:"districts,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type District struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	Name       string    `json:"name" gorm:"not null"`
	ProvinceID uint      `json:"province_id" gorm:"not null"`
	Province   Province  `json:"province,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type ProfessionGroup struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" gorm:"not null;unique"`
	Titles    []Title   `json:"titles,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Title struct {
	ID                uint            `json:"id" gorm:"primaryKey"`
	Name              string          `json:"name" gorm:"not null"`
	ProfessionGroupID uint            `json:"profession_group_id" gorm:"not null"`
	ProfessionGroup   ProfessionGroup `json:"profession_group,omitempty"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
}

type ClinicType struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" gorm:"not null;unique"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Hospital struct {
	ID         uint           `json:"id" gorm:"primaryKey"`
	Name       string         `json:"name" gorm:"not null"`
	TaxID      string         `json:"tax_id" gorm:"not null;unique"`
	Email      string         `json:"email" gorm:"not null;unique"`
	Phone      string         `json:"phone" gorm:"not null;unique"`
	ProvinceID uint           `json:"province_id" gorm:"not null"`
	DistrictID uint           `json:"district_id" gorm:"not null"`
	Address    string         `json:"address" gorm:"not null"`
	Province   Province       `json:"province,omitempty"`
	District   District       `json:"district,omitempty"`
	Users      []User         `json:"users,omitempty"`
	Clinics    []Clinic       `json:"clinics,omitempty"`
	Staff      []Staff        `json:"staff,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index" swaggertype:"string" format:"date-time"`
}

type UserType string

const (
	UserTypeAuthorized UserType = "authorized"
	UserTypeEmployee   UserType = "employee"
)

type User struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	FirstName   string         `json:"first_name" gorm:"not null"`
	LastName    string         `json:"last_name" gorm:"not null"`
	NationalID  string         `json:"national_id" gorm:"not null;unique"`
	Email       string         `json:"email" gorm:"not null;unique"`
	Phone       string         `json:"phone" gorm:"not null;unique"`
	Password    string         `json:"-" gorm:"not null"`
	UserType    UserType       `json:"user_type" gorm:"not null;default:'employee'"`
	HospitalID  uint           `json:"hospital_id" gorm:"not null"`
	Hospital    Hospital       `json:"hospital,omitempty"`
	CreatedByID *uint          `json:"created_by_id,omitempty"`
	CreatedBy   *User          `json:"created_by,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index" swaggertype:"string" format:"date-time"`
}

type Clinic struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	HospitalID   uint           `json:"hospital_id" gorm:"not null"`
	ClinicTypeID uint           `json:"clinic_type_id" gorm:"not null"`
	Hospital     Hospital       `json:"hospital,omitempty"`
	ClinicType   ClinicType     `json:"clinic_type,omitempty"`
	Staff        []Staff        `json:"staff,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index" swaggertype:"string" format:"date-time"`
}

type WorkingDay string

const (
	Monday    WorkingDay = "monday"
	Tuesday   WorkingDay = "tuesday"
	Wednesday WorkingDay = "wednesday"
	Thursday  WorkingDay = "thursday"
	Friday    WorkingDay = "friday"
	Saturday  WorkingDay = "saturday"
	Sunday    WorkingDay = "sunday"
)

type Staff struct {
	ID                uint            `json:"id" gorm:"primaryKey"`
	FirstName         string          `json:"first_name" gorm:"not null"`
	LastName          string          `json:"last_name" gorm:"not null"`
	NationalID        string          `json:"national_id" gorm:"not null;unique"`
	Phone             string          `json:"phone" gorm:"not null;unique"`
	ProfessionGroupID uint            `json:"profession_group_id" gorm:"not null"`
	TitleID           uint            `json:"title_id" gorm:"not null"`
	HospitalID        uint            `json:"hospital_id" gorm:"not null"`
	ClinicID          *uint           `json:"clinic_id,omitempty"`
	WorkingDays       string          `json:"working_days" gorm:"type:text"`
	ProfessionGroup   ProfessionGroup `json:"profession_group,omitempty"`
	Title             Title           `json:"title,omitempty"`
	Hospital          Hospital        `json:"hospital,omitempty"`
	Clinic            *Clinic         `json:"clinic,omitempty"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
	DeletedAt         gorm.DeletedAt  `json:"deleted_at,omitempty" gorm:"index" swaggertype:"string" format:"date-time"`
}

type PasswordReset struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Phone     string    `json:"phone" gorm:"not null"`
	Code      string    `json:"code" gorm:"not null"`
	ExpiresAt time.Time `json:"expires_at" gorm:"not null"`
	Used      bool      `json:"used" gorm:"default:false"`
	CreatedAt time.Time `json:"created_at"`
}

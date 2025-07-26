package models

import "time"

type HospitalRegistrationRequest struct {
	HospitalName string `json:"hospital_name" binding:"required"`
	TaxID        string `json:"tax_id" binding:"required"`
	Email        string `json:"email" binding:"required,email"`
	Phone        string `json:"phone" binding:"required"`
	ProvinceID   uint   `json:"province_id" binding:"required"`
	DistrictID   uint   `json:"district_id" binding:"required"`
	Address      string `json:"address" binding:"required"`

	FirstName  string `json:"first_name" binding:"required"`
	LastName   string `json:"last_name" binding:"required"`
	NationalID string `json:"national_id" binding:"required"`
	UserEmail  string `json:"user_email" binding:"required,email"`
	UserPhone  string `json:"user_phone" binding:"required"`
	Password   string `json:"password" binding:"required,min=6"`
}

type LoginRequest struct {
	Identifier string `json:"identifier" binding:"required"`
	Password   string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token    string `json:"token"`
	UserType string `json:"user_type"`
	User     User   `json:"user"`
}

type PasswordResetRequest struct {
	Phone string `json:"phone" binding:"required"`
}

type PasswordResetConfirmRequest struct {
	Phone           string `json:"phone" binding:"required"`
	Code            string `json:"code" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=6"`
	ConfirmPassword string `json:"confirm_password" binding:"required,min=6"`
}

type PasswordResetResponse struct {
	Code string `json:"code"`
}

type CreateUserRequest struct {
	FirstName  string   `json:"first_name" binding:"required"`
	LastName   string   `json:"last_name" binding:"required"`
	NationalID string   `json:"national_id" binding:"required"`
	Email      string   `json:"email" binding:"required,email"`
	Phone      string   `json:"phone" binding:"required"`
	Password   string   `json:"password" binding:"required,min=6"`
	UserType   UserType `json:"user_type" binding:"required,oneof=authorized employee"`
}

type UpdateUserRequest struct {
	FirstName  string   `json:"first_name"`
	LastName   string   `json:"last_name"`
	NationalID string   `json:"national_id"`
	Email      string   `json:"email"`
	Phone      string   `json:"phone"`
	UserType   UserType `json:"user_type" binding:"omitempty,oneof=authorized employee"`
}

type CreateClinicRequest struct {
	ClinicTypeID uint `json:"clinic_type_id" binding:"required"`
}

type CreateStaffRequest struct {
	FirstName         string       `json:"first_name" binding:"required"`
	LastName          string       `json:"last_name" binding:"required"`
	NationalID        string       `json:"national_id" binding:"required"`
	Phone             string       `json:"phone" binding:"required"`
	ProfessionGroupID uint         `json:"profession_group_id" binding:"required"`
	TitleID           uint         `json:"title_id" binding:"required"`
	ClinicID          *uint        `json:"clinic_id,omitempty"`
	WorkingDays       []WorkingDay `json:"working_days"`
}

type UpdateStaffRequest struct {
	FirstName         string       `json:"first_name"`
	LastName          string       `json:"last_name"`
	NationalID        string       `json:"national_id"`
	Phone             string       `json:"phone"`
	ProfessionGroupID uint         `json:"profession_group_id"`
	TitleID           uint         `json:"title_id"`
	ClinicID          *uint        `json:"clinic_id"`
	WorkingDays       []WorkingDay `json:"working_days"`
}

type StaffFilterRequest struct {
	FirstName         string `form:"first_name"`
	LastName          string `form:"last_name"`
	NationalID        string `form:"national_id"`
	ProfessionGroupID uint   `form:"profession_group_id"`
	TitleID           uint   `form:"title_id"`
	Page              int    `form:"page,default=1"`
	Limit             int    `form:"limit,default=10"`
}

type BasePagination struct {
	TotalCount int64 `json:"total_count"`
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalPages int   `json:"total_pages"`
}

type StaffPaginatedResponse struct {
	Data []Staff `json:"data"`
	BasePagination
}

type UserPaginatedResponse struct {
	Data []User `json:"data"`
	BasePagination
}

type ClinicPaginatedResponse struct {
	Data []Clinic `json:"data"`
	BasePagination
}

type ClinicSummary struct {
	ID                uint                     `json:"id"`
	ClinicType        ClinicType               `json:"clinic_type"`
	TotalStaff        int64                    `json:"total_staff"`
	StaffByProfession []StaffProfessionSummary `json:"staff_by_profession"`
}

type StaffProfessionSummary struct {
	ProfessionGroup string `json:"profession_group"`
	Count           int64  `json:"count"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

type TitleResponse struct {
	ID                uint      `json:"id"`
	Name              string    `json:"name"`
	ProfessionGroupID uint      `json:"profession_group_id"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type ProfessionGroupResponse struct {
	ID        uint            `json:"id"`
	Name      string          `json:"name"`
	Titles    []TitleResponse `json:"titles,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

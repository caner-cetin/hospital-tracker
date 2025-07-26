package handlers

import (
	"github.com/caner-cetin/hospital-tracker/internal/config"
	"github.com/caner-cetin/hospital-tracker/internal/middleware"
	"github.com/caner-cetin/hospital-tracker/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func SetupRoutes(router *gin.RouterGroup, db *gorm.DB, redisClient *redis.Client, cfg *config.Config) {
	authService := services.NewAuthService(db, cfg)
	hospitalService := services.NewHospitalService(db, authService)
	passwordResetService := services.NewPasswordResetService(db, authService)
	userService := services.NewUserService(db, authService)
	clinicService := services.NewClinicService(db)
	staffService := services.NewStaffService(db, redisClient)
	locationService := services.NewLocationService(db, redisClient)

	authHandler := NewAuthHandler(authService)
	hospitalHandler := NewHospitalHandler(hospitalService)
	passwordResetHandler := NewPasswordResetHandler(passwordResetService)
	userHandler := NewUserHandler(userService)
	clinicHandler := NewClinicHandler(clinicService)
	staffHandler := NewStaffHandler(staffService)
	locationHandler := NewLocationHandler(locationService)

	router.POST("/register", hospitalHandler.Register)
	router.POST("/login", authHandler.Login)
	router.POST("/password-reset/request", passwordResetHandler.RequestReset)
	router.POST("/password-reset/confirm", passwordResetHandler.ConfirmReset)

	router.GET("/provinces", locationHandler.GetProvinces)
	router.GET("/districts", locationHandler.GetDistricts)
	router.GET("/clinic-types", clinicHandler.GetClinicTypes)
	router.GET("/profession-groups", staffHandler.GetProfessionGroups)

	protected := router.Group("/")
	protected.Use(middleware.AuthRequired(authService))
	{
		protected.GET("/users", userHandler.GetUsers)
		protected.GET("/users/:id", userHandler.GetUser)

		protected.GET("/clinics", clinicHandler.GetClinics)

		protected.GET("/staff", staffHandler.GetStaff)
		protected.GET("/staff/:id", staffHandler.GetStaffByID)
	}

	authorized := protected.Group("/")
	authorized.Use(middleware.AuthorizedOnly())
	{
		authorized.POST("/users", userHandler.CreateUser)
		authorized.PUT("/users/:id", userHandler.UpdateUser)
		authorized.DELETE("/users/:id", userHandler.DeleteUser)

		authorized.POST("/clinics", clinicHandler.CreateClinic)
		authorized.DELETE("/clinics/:id", clinicHandler.DeleteClinic)

		authorized.POST("/staff", staffHandler.CreateStaff)
		authorized.PUT("/staff/:id", staffHandler.UpdateStaff)
		authorized.DELETE("/staff/:id", staffHandler.DeleteStaff)
	}
}

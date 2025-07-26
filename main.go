package main

import (
	"net/http"

	"github.com/caner-cetin/hospital-tracker/internal/config"
	"github.com/caner-cetin/hospital-tracker/internal/database"
	"github.com/caner-cetin/hospital-tracker/internal/handlers"
	"github.com/caner-cetin/hospital-tracker/internal/middleware"
	"github.com/caner-cetin/hospital-tracker/internal/redis"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/caner-cetin/hospital-tracker/docs"
)

// @title Hospital Tracker API
// @version 1.0
// @description hospital management and tracking platform
// @termsOfService http://swagger.io/terms/

// @license.name GNU GPLv3
// @license.url https://opensource.org/license/gpl-3-0

// @host hospital.cansu.dev
// @BasePath /api

// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	cfg := config.Load()
	config.InitLogger(&cfg.Logging)

	log.Info().Msg("Starting Hospital Tracker API")

	db, err := database.Initialize(cfg.Database)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize database")
	}
	redisClient, err := redis.Initialize(cfg.Redis)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize Redis")
	}
	r := gin.Default()
	r.Use(middleware.CORS())
	api := r.Group("/api")
	handlers.SetupRoutes(api, db, redisClient, cfg)
	r.GET("/health", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL("doc.json"), ginSwagger.DocExpansion("none"), ginSwagger.PersistAuthorization(true)))

	log.Info().Str("port", cfg.Server.Port).Msg("Server starting")
	log.Info().Str("url", "http://localhost:"+cfg.Server.Port+"/swagger/index.html").Msg("Swagger documentation available")

	if err := r.Run(":" + cfg.Server.Port); err != nil {
		log.Fatal().Err(err).Msg("Failed to start server")
	}
}

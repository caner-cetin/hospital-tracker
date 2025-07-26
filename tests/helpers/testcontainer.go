package helpers

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/caner-cetin/hospital-tracker/internal/config"
	"github.com/caner-cetin/hospital-tracker/internal/database"
	redisClient "github.com/caner-cetin/hospital-tracker/internal/redis"
	"github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	redisContainer "github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/gorm"
)

type TestContainers struct {
	PostgresContainer testcontainers.Container
	RedisContainer    testcontainers.Container
	DB                *gorm.DB
	Redis             *redis.Client
	Config            *config.Config
}

func SetupTestContainers(ctx context.Context) (*TestContainers, error) {
	postgresContainer, err := postgres.Run(ctx,
		"postgres:15-alpine",
		postgres.WithDatabase("test_hospital_tracker"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Minute)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start postgres container: %w", err)
	}

	redisContainer, err := redisContainer.Run(ctx,
		"redis:7-alpine",
		testcontainers.WithWaitStrategy(wait.ForLog("Ready to accept connections")),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start redis container: %w", err)
	}

	postgresHost, err := postgresContainer.Host(ctx)
	if err != nil {
		return nil, err
	}

	postgresPort, err := postgresContainer.MappedPort(ctx, "5432")
	if err != nil {
		return nil, err
	}

	redisHost, err := redisContainer.Host(ctx)
	if err != nil {
		return nil, err
	}

	redisPort, err := redisContainer.MappedPort(ctx, "6379")
	if err != nil {
		return nil, err
	}

	cfg := &config.Config{
		Server: config.ServerConfig{
			Port: "8080",
			Env:  "test",
		},
		Database: config.DatabaseConfig{
			Host:     postgresHost,
			Port:     postgresPort.Port(),
			User:     "test",
			Password: "test",
			Name:     "test_hospital_tracker",
			SSLMode:  "disable",
		},
		Redis: config.RedisConfig{
			Host:     redisHost,
			Port:     redisPort.Port(),
			Password: "",
			DB:       0,
		},
		JWT: config.JWTConfig{
			Secret:      "test-secret",
			ExpireHours: 24,
		},
	}

	db, err := database.Initialize(cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	redisConn, err := redisClient.Initialize(cfg.Redis)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize redis: %w", err)
	}

	return &TestContainers{
		PostgresContainer: postgresContainer,
		RedisContainer:    redisContainer,
		DB:                db,
		Redis:             redisConn,
		Config:            cfg,
	}, nil
}

func (tc *TestContainers) Cleanup(ctx context.Context) error {
	var errs []error

	if err := tc.PostgresContainer.Terminate(ctx); err != nil {
		errs = append(errs, err)
		log.Printf("Failed to terminate postgres container: %v", err)
	}

	if err := tc.RedisContainer.Terminate(ctx); err != nil {
		errs = append(errs, err)
		log.Printf("Failed to terminate redis container: %v", err)
	}

	if len(errs) > 0 {
		return fmt.Errorf("cleanup errors: %v", errs)
	}

	return nil
}

func (tc *TestContainers) CleanDatabase() error {
	tc.DB.Exec("SET session_replication_role = replica")

	tables := []string{
		"staffs",
		"password_resets",
		"clinics",
		"users",
		"hospitals",
	}

	for _, table := range tables {
		if err := tc.DB.Exec(fmt.Sprintf("DELETE FROM %s", table)).Error; err != nil {
			fmt.Printf("Warning: failed to clean table %s: %v\n", table, err)
		}
	}

	tc.DB.Exec("SET session_replication_role = DEFAULT")

	return tc.Redis.FlushAll(context.Background()).Err()
}

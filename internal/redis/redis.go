package redis

import (
	"context"
	"fmt"

	"github.com/caner-cetin/hospital-tracker/internal/config"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

func Initialize(cfg config.RedisConfig) (*redis.Client, error) {
	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	log.Info().Str("addr", addr).Int("db", cfg.DB).Msg("Connecting to Redis")

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		log.Error().Err(err).Msg("Failed to connect to Redis")
		return nil, err
	}

	log.Info().Msg("Redis connection established")
	return client, nil
}

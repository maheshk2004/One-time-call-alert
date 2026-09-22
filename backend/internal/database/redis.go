package database

import (
	"context"
	"log"
	"time"

	"lead-followup-system/internal/config"

	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

type RedisService struct {
	Client      *redis.Client
	AsynqClient *asynq.Client
	RedisOpt    asynq.RedisClientOpt
}

func ConnectRedis(cfg *config.Config) (*RedisService, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.RedisAddr,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	redisOpt := asynq.RedisClientOpt{
		Addr: cfg.RedisAddr,
	}

	asynqClient := asynq.NewClient(redisOpt)

	log.Printf("Successfully connected to Redis at: %s", cfg.RedisAddr)

	return &RedisService{
		Client:      rdb,
		AsynqClient: asynqClient,
		RedisOpt:    redisOpt,
	}, nil
}

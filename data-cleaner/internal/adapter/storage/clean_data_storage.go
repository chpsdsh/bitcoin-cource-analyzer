package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"data-cleaner/internal/adapter/config"
)

const (
	setValue           = "1"
	expirationDuration = 1 * time.Hour
)

type CleanDataStorage struct {
	redis *redis.Client
}

func NewDataStorage(conf config.RedisConfig) *CleanDataStorage {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     conf.RedisAddr,
		Password: conf.RedisPassword,
		DB:       conf.RedisDB,
	})

	return &CleanDataStorage{redis: redisClient}
}

func (r CleanDataStorage) CloseRedis() error {
	if err := r.redis.Close(); err != nil {
		return fmt.Errorf("close redis connection: %w", err)
	}
	return nil
}

func (r CleanDataStorage) AddNewArticle(ctx context.Context, key string) (bool, error) {
	res, err := r.redis.SetNX(ctx, key, setValue, expirationDuration).Result()
	return res, err
}

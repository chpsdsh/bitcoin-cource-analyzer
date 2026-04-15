package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"news-gateway/internal/adapter/config"
	"news-gateway/internal/domain"
)

type NewsStorage struct {
	redis *redis.Client
}

func NewNewsStorage(conf config.RedisConfig) *NewsStorage {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     conf.RedisAddr,
		Password: conf.RedisPassword,
		DB:       conf.RedisDB,
	})

	return &NewsStorage{redis: redisClient}
}

func (r NewsStorage) CloseRedis() error {
	if err := r.redis.Close(); err != nil {
		return fmt.Errorf("close redis connection: %w", err)
	}
	return nil
}

func (r NewsStorage) AddNews(ctx context.Context, dto domain.NewsDto, key string) error {
	data, err := json.Marshal(dto)
	if err != nil {
		return fmt.Errorf("marshal news: %w", err)
	}

	score := float64(time.Now().Unix())
	r.redis.ZAdd(ctx, key, redis.Z{
		Score:  score,
		Member: string(data),
	})
	return nil
}

func (r NewsStorage) GetNews(ctx context.Context, limit int64, key string) ([]domain.NewsDto, error) {
	newsDtoArr := make([]domain.NewsDto, 0, limit)

	res, err := r.redis.ZRangeArgs(ctx, redis.ZRangeArgs{
		Key:   key,
		Start: 0,
		Stop:  limit - 1,
		Rev:   true,
	}).Result()
	if err != nil {
		return nil, fmt.Errorf("get news: %w", err)
	}

	for _, val := range res {
		var news domain.NewsDto
		if err = json.Unmarshal([]byte(val), &news); err != nil {
			return nil, fmt.Errorf("unmarshal news: %w", err)
		}
		newsDtoArr = append(newsDtoArr, news)
	}

	return newsDtoArr, nil
}

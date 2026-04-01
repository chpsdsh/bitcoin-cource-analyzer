package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"data-cleaner/internal/adapter/config"
	"data-cleaner/internal/domain"
)

type CleanDataStorage struct {
	redisNewsClient     *redis.Client
	redisArticlesClient *redis.Client
}

func NewDataStorage(conf config.RedisConfig) *CleanDataStorage {
	redisNewsClient := redis.NewClient(&redis.Options{
		Addr:     conf.RedisAddr,
		Password: conf.RedisPassword,
		DB:       conf.RedisNewsDB,
	})

	redisArticlesClient := redis.NewClient(&redis.Options{
		Addr:     conf.RedisAddr,
		Password: conf.RedisPassword,
		DB:       conf.RedisArticlesDB,
	})

	return &CleanDataStorage{redisNewsClient: redisNewsClient, redisArticlesClient: redisArticlesClient}
}

func (r CleanDataStorage) CloseRedis() error {
	if err := r.redisArticlesClient.Close(); err != nil {
		return fmt.Errorf("close redis connection: %w", err)
	}

	if err := r.redisNewsClient.Close(); err != nil {
		return fmt.Errorf("close redis connection: %w", err)
	}
	return nil
}

func (r CleanDataStorage) AddArticle(ctx context.Context, dto domain.ArticleDto) error {
	data, err := json.Marshal(dto)
	if err != nil {
		return fmt.Errorf("marshal news: %w", err)
	}
	score := float64(time.Now().Unix())

	r.redisArticlesClient.ZAdd(ctx, dto.Category, redis.Z{Score: score, Member: string(data)})
	return nil
}

func (r CleanDataStorage) AddNews(ctx context.Context, dto domain.NewsDto) error {
	data, err := json.Marshal(dto)
	if err != nil {
		return fmt.Errorf("marshal news: %w", err)
	}

	score := float64(time.Now().Unix())
	r.redisNewsClient.ZAdd(ctx, dto.Category, redis.Z{
		Score:  score,
		Member: string(data),
	})
	return nil
}

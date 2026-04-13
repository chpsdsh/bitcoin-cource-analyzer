package storage

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"

	"llm-consumer/internal/adapter/config"
	"llm-consumer/internal/domain"
)

type LLMStorage struct {
	redis *redis.Client
}

func NewLLMStorage(conf config.RedisConfig) *LLMStorage {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     conf.RedisAddr,
		Password: conf.RedisPassword,
		DB:       conf.RedisDB,
	})

	return &LLMStorage{redis: redisClient}
}

func (s LLMStorage) CloseRedis() error {
	if err := s.redis.Close(); err != nil {
		return fmt.Errorf("close redis connection: %w", err)
	}
	return nil
}

func (s LLMStorage) GetPrediction(ctx context.Context, category string) (domain.LLMResponse, error) {
	llmResponse, err := s.redis.Get(ctx, category).Result()
	if err != nil {
		return domain.LLMResponse{}, fmt.Errorf("get prediction: %w", err)
	}
	var response domain.LLMResponse
	if err = json.Unmarshal([]byte(llmResponse), &response); err != nil {
		return domain.LLMResponse{}, fmt.Errorf("unmarshal response: %w", err)
	}
	return response, nil
}

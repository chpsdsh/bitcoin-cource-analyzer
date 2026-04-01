package config

import (
	"errors"
	"os"
	"strconv"
)

var (
	ErrRedisAddrNotSet       = errors.New("redis address should be set with REDIS_ADDR env variable")
	ErrRedisPasswordNotSet   = errors.New("redis password should be set with REDIS_PASSWORD env variable")
	ErrRedisNewsDBNotSet     = errors.New("redis db should be set with REDIS_NEWS_DB env variable")
	ErrRedisArticlesDBNotSet = errors.New("redis db should be set with REDIS_ARTICLES_DB env variable")
)

type RedisConfig struct {
	RedisAddr       string
	RedisPassword   string
	RedisArticlesDB int
	RedisNewsDB     int
}

func NewRedisConfig() (RedisConfig, error) {
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		return RedisConfig{}, ErrRedisAddrNotSet
	}
	redisPassword := os.Getenv("REDIS_PASSWORD")
	if redisPassword == "" {
		return RedisConfig{}, ErrRedisPasswordNotSet
	}
	redisArticlesDB, err := strconv.Atoi(os.Getenv("REDIS_ARTICLES_DB"))
	if err != nil {
		return RedisConfig{}, errors.Join(ErrRedisNewsDBNotSet, err)
	}
	redisNewsDB, err := strconv.Atoi(os.Getenv("REDIS_NEWS_DB"))
	if err != nil {
		return RedisConfig{}, errors.Join(ErrRedisArticlesDBNotSet, err)
	}
	return RedisConfig{RedisAddr: redisAddr,
		RedisPassword:   redisPassword,
		RedisArticlesDB: redisArticlesDB,
		RedisNewsDB:     redisNewsDB}, nil
}

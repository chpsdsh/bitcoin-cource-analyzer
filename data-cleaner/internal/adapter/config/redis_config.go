package config

import (
	"errors"
	"os"
	"strconv"
)

var (
	ErrRedisAddrNotSet     = errors.New("redis address should be set with REDIS_ADDR env variable")
	ErrRedisPasswordNotSet = errors.New("redis password should be set with REDIS_PASSWORD env variable")
	ErrRedisDBNotSet       = errors.New("redis db should be set with REDIS_DB env variable")
)

type RedisConfig struct {
	RedisAddr     string
	RedisPassword string
	RedisDB       int
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
	redisDb, err := strconv.Atoi(os.Getenv("REDIS_DB"))
	if err != nil {
		return RedisConfig{}, errors.Join(ErrRedisDBNotSet, err)
	}
	return RedisConfig{RedisAddr: redisAddr, RedisPassword: redisPassword, RedisDB: redisDb}, nil
}

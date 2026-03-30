package storage

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/redis/go-redis/v9"
)

type NewsStorage struct {
	redis *redis.Client
}

func NewNewsStorage() (*NewsStorage, error) {

	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: "",
		DB:       0,
	})
	return &NewsStorage{}
}

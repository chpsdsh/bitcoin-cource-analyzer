package application

import (
	"context"
	"errors"

	"news-gateway/internal/domain"
)

var (
	ErrInternalError    = errors.New("internal error")
	ErrNotValidCategory = errors.New("not a valid category")
)

const (
	newsLimit           = 5
	politicsCategory    = "politics"
	environmentCategory = "environment"
	economyCategory     = "economy"
	technologyCategory  = "technology"
	cryptoCategory      = "crypto"
)

type Storage interface {
	GetNews(ctx context.Context, limit int64, key string) ([]domain.NewsDto, error)
}

type NewsService struct {
	Storage Storage
}

func (n NewsService) RequestNews(ctx context.Context, key string) ([]domain.NewsDto, error) {
	switch key {
	case politicsCategory, environmentCategory, economyCategory, technologyCategory, cryptoCategory:
		newsDtoArr, err := n.Storage.GetNews(ctx, newsLimit, key)
		if err != nil {
			return nil, errors.Join(err, ErrInternalError)
		}
		return newsDtoArr, nil
	default:
		return nil, ErrNotValidCategory
	}
}

package application

import (
	"context"
	"errors"

	"news-gateway/internal/domain"
)

var ErrInternalError = errors.New("internal error")

const newsLimit = 20

type Storage interface {
	GetNews(ctx context.Context, limit int64) ([]domain.NewsDto, error)
}

type NewsService struct {
	Storage Storage
}

func (n NewsService) RequestNews(ctx context.Context) ([]domain.NewsDto, error) {
	newsDtoArr, err := n.Storage.GetNews(ctx, newsLimit)
	if err != nil {
		return nil, errors.Join(err, ErrInternalError)
	}
	return newsDtoArr, nil
}

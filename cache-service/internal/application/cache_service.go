package application

import (
	"context"
	"log/slog"

	"data-cleaner/internal/domain"
)

type Storage interface {
	AddArticle(ctx context.Context, dto domain.ArticleDto) error
	AddNews(ctx context.Context, dto domain.NewsDto) error
	AddLLMResponse(ctx context.Context, dto domain.LLMResponse) error
}

type CacheService struct {
	Storage Storage
}

func (c CacheService) HandleArticle(ctx context.Context, article domain.ArticleDto) {
	if err := c.Storage.AddArticle(ctx, article); err != nil {
		slog.Error("error adding new article to storage", slog.String("error", err.Error()), slog.Any("article", article))
	}
}

func (c CacheService) HandleNews(ctx context.Context, news domain.NewsDto) {
	if err := c.Storage.AddNews(ctx, news); err != nil {
		slog.Error("error adding new news to storage", slog.String("error", err.Error()), slog.Any("news", news))
	}
}

func (c CacheService) HandleLLMResponse(ctx context.Context, response domain.LLMResponse) {
	if err := c.Storage.AddLLMResponse(ctx, response); err != nil {
		slog.Error("error adding new llm response to storage", slog.String("error", err.Error()), slog.Any("llm", response))
	}
}

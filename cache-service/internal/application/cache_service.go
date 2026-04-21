//go:generate mockgen -source cache_service.go -destination=mocks/cache_service.go -package=mocks

package application

import (
	"context"
	"log/slog"

	"data-cleaner/internal/domain"
	"data-cleaner/internal/observability"
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
	traceID := observability.TraceIDFromContext(ctx)
	if err := c.Storage.AddArticle(ctx, article); err != nil {
		slog.Error("error adding new article to storage", slog.String("trace_id", traceID), slog.String("error", err.Error()), slog.Any("article", article))
		return
	}
	slog.Info("article cached", slog.String("trace_id", traceID), slog.String("category", article.Category))
}

func (c CacheService) HandleNews(ctx context.Context, news domain.NewsDto) {
	traceID := observability.TraceIDFromContext(ctx)
	if err := c.Storage.AddNews(ctx, news); err != nil {
		slog.Error("error adding new news to storage", slog.String("trace_id", traceID), slog.String("error", err.Error()), slog.Any("news", news))
		return
	}
	slog.Info("news cached", slog.String("trace_id", traceID), slog.String("category", news.Category))
}

func (c CacheService) HandleLLMResponse(ctx context.Context, response domain.LLMResponse) {
	traceID := observability.TraceIDFromContext(ctx)
	if err := c.Storage.AddLLMResponse(ctx, response); err != nil {
		slog.Error("error adding new llm response to storage", slog.String("trace_id", traceID), slog.String("error", err.Error()), slog.Any("llm", response))
		return
	}
	slog.Info("llm response cached", slog.String("trace_id", traceID), slog.String("category", response.Category))
}

//go:generate mockgen -source app.go -destination=mocks/app_mocks.go -package=mocks

package application

import (
	"context"
	"log/slog"
	"time"

	"news-parser/internal/domain"
	"news-parser/internal/observability"
)

const NewsRequestsCount = 50

type RequestHandler interface {
	DoNewsRequest(ctx context.Context, category domain.Category) (domain.Articles, error)
	DoDataRequest(ctx context.Context, url string) ([]byte, error)
}

type LLMNotifier interface {
	StartLLMPrediction(ctx context.Context) error
}

type Application struct {
	Ticker         *time.Ticker
	RequestHandler RequestHandler
	LLMNotifier    LLMNotifier
	RequestChan    chan domain.GdeltAPIDto
	NewsChan       chan domain.NewsDto
}

func (a Application) StartParsingNews(ctx context.Context) {
	defer a.Ticker.Stop()
	defer close(a.RequestChan)

	for {
		select {
		case <-ctx.Done():
			return
		case <-a.Ticker.C:
			traceID := observability.NewTraceID()
			cycleCtx := observability.ContextWithTraceID(ctx, traceID)
			slog.Info("news parsing cycle started", slog.String("trace_id", traceID))
			a.parseNews(cycleCtx)
			if err := a.LLMNotifier.StartLLMPrediction(cycleCtx); err != nil {
				slog.Error("error starting llm", "trace_id", traceID, "err", err)
			}
		}
	}
}

func (a Application) parseNews(ctx context.Context) {
	traceID := observability.TraceIDFromContext(ctx)
	for _, category := range domain.AllCategories {
		news, err := a.RequestHandler.DoNewsRequest(ctx, category)
		if err != nil {
			slog.Error("error requesting news:", "trace_id", traceID, "err", err)
			continue
		}
		stringCategory := domain.CategoryToString(category)

		go func() {
			for _, dto := range news.Articles {
				dto.TraceID = traceID
				dto.Category = stringCategory

				a.RequestChan <- dto

				a.NewsChan <- domain.NewsDto{TraceID: traceID,
					Category:    stringCategory,
					Title:       dto.Title,
					URL:         dto.URL,
					SocialImage: dto.SocialImage}
			}
		}()
	}
}

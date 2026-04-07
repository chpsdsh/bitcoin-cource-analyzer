package application

import (
	"context"
	"log/slog"
	"time"

	"news-parser/internal/domain"
)

const NewsRequestsCount = 50

type RequestHandler interface {
	DoNewsRequest(category domain.Category) (domain.Articles, error)
	DoDataRequest(url string) ([]byte, error)
}

type LLMNotifier interface {
	StartLLMPrediction() error
}

type Application struct {
	Ticker         *time.Ticker
	RequestHandler RequestHandler
	LLMNotifier    LLMNotifier
	RequestChan    chan domain.GdeltApiDto
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
			for _, category := range domain.AllCategories {
				news, err := a.RequestHandler.DoNewsRequest(category)
				if err != nil {

				}
				stringCategory := domain.CategoryToString(category)

				go func() {
					for _, dto := range news.Articles {
						dto.Category = stringCategory

						a.RequestChan <- dto

						a.NewsChan <- domain.NewsDto{Category: stringCategory,
							Title:       dto.Title,
							URL:         dto.URL,
							SocialImage: dto.SocialImage}
					}
				}()
				if err != nil {
					slog.Error("error requesting news:", "err", err)
				}
			}

			if err := a.LLMNotifier.StartLLMPrediction(); err != nil {
				slog.Error("error starting llm", "err", err)
			}
		}
	}
}

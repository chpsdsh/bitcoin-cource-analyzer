package application

import (
	"context"
	"log/slog"
	"news-parser/internal/domain"
	"time"
)

const NewsRequestsCount = 50

type RequestHandler interface {
	DoNewsRequest(category domain.Category) (domain.NewsArticles, error)
	DoDataRequest(url string) ([]byte, error)
}

type Application struct {
	Ticker         *time.Ticker
	RequestHandler RequestHandler
	RequestChan    chan domain.GdeltApiDto
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
				stringCategory := domain.CategoryToString(category)

				go func() {
					for _, dto := range news.Articles {
						dto.Category = stringCategory
						select {
						case a.RequestChan <- dto:
						case <-ctx.Done():
							return
						}
					}
				}()
				if err != nil {
					slog.Error("error requesting news:", "err", err)
				}
			}
		}
	}
}

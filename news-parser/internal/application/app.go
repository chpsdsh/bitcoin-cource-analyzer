package application

import (
	"context"
	"log/slog"
	"news-parser/internal/domain"
	"time"
)

const NewsRequestsCount = 50

type RequestHandler interface {
	DoNewsRequest() (domain.NewsArticles, error)
	DoDataRequest(url string) ([]byte, error)
}

type Application struct {
	Ticker         *time.Ticker
	RequestHandler RequestHandler
	RequestChan    chan domain.GdeltApiDto
}

func (a Application) StartParsingNews(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				a.Ticker.Stop()
				return
			case <-a.Ticker.C:
				news, err := a.RequestHandler.DoNewsRequest()
				if err != nil {
					slog.Error("error requesting news:", "err", err)
				}

				for _, dto := range news.Articles {
					a.RequestChan <- dto
				}
			}
		}
	}()
}

package kafkaproducer

import (
	"context"
	"sync"

	"news-parser/internal/domain"
	"news-parser/internal/observability"
)

const NumKafkaSenders = 5

type Sender interface {
	SendArticle(ctx context.Context, dto domain.ArticleDto)
	SendNews(ctx context.Context, dto domain.NewsDto)
}

type KafkaSender struct {
	articles chan domain.ArticleDto
	news     chan domain.NewsDto
	producer Sender
}

func NewSenderPool(ctx context.Context, wg *sync.WaitGroup, articles chan domain.ArticleDto, news chan domain.NewsDto, producer Sender) {
	for range NumKafkaSenders {
		worker := KafkaSender{articles: articles, news: news, producer: producer}
		wg.Go(func() { worker.sendDataToKafka(ctx) })
	}
}

func (s KafkaSender) sendDataToKafka(ctx context.Context) {
	for {
		select {
		case dto, ok := <-s.articles:
			if !ok {
				return
			}
			dtoCtx := observability.ContextWithTraceID(ctx, dto.TraceID)
			s.producer.SendArticle(dtoCtx, dto)
		case news, ok := <-s.news:
			if !ok {
				return
			}
			newsCtx := observability.ContextWithTraceID(ctx, news.TraceID)
			s.producer.SendNews(newsCtx, news)
		case <-ctx.Done():
			return
		}
	}
}

package kafkaproducer

import (
	"context"
	"news-parser/internal/domain"
	"sync"
)

const NumKafkaSenders = 5

type Sender interface {
	SendArticle(dto domain.ArticleDto)
	SendNews(dto domain.NewsDto)
}

type KafkaSender struct {
	articles chan domain.ArticleDto
	news     chan domain.NewsDto
	producer Sender
}

func NewSenderPool(ctx context.Context, wg *sync.WaitGroup, articles chan domain.ArticleDto, news chan domain.NewsDto, producer Sender) {
	for i := 0; i < NumKafkaSenders; i++ {
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
			s.producer.SendArticle(dto)
		case news, ok := <-s.news:
			if !ok {
				return
			}
			s.producer.SendNews(news)
		case <-ctx.Done():
			return
		}
	}
}

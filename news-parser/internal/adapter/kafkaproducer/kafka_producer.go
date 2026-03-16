package kafkaproducer

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"news-parser/internal/domain"
	"os"
	"time"

	"github.com/segmentio/kafka-go"
)

const (
	kafkaSendTimeout = time.Second * 15
)

type KafkaProducer struct {
	articlesWriter *kafka.Writer
	newsWriter     *kafka.Writer
}

func NewKafkaProducer() KafkaProducer {
	brokers := os.Getenv("KAFKA_BROKERS")
	articlesTopic := os.Getenv("KAFKA_ARTICLES_TOPIC")
	newsTopic := os.Getenv("KAFKA_NEWS_TOPIC")

	articlesWriter := &kafka.Writer{
		Addr:                   kafka.TCP(brokers),
		Topic:                  articlesTopic,
		Balancer:               &kafka.Hash{},
		RequiredAcks:           kafka.RequireAll,
		AllowAutoTopicCreation: true,
	}

	newsWriter := &kafka.Writer{
		Addr:                   kafka.TCP(brokers),
		Topic:                  newsTopic,
		Balancer:               &kafka.Hash{},
		RequiredAcks:           kafka.RequireAll,
		AllowAutoTopicCreation: true,
	}

	return KafkaProducer{articlesWriter: articlesWriter, newsWriter: newsWriter}
}

func (p KafkaProducer) SendArticle(dto domain.ArticleDto) {
	b, err := json.Marshal(dto)
	if err != nil {
		slog.Error(err.Error())
	}
	ctx, cancel := context.WithTimeout(context.Background(), kafkaSendTimeout)
	defer cancel()
	if err = p.articlesWriter.WriteMessages(ctx, kafka.Message{Key: []byte(dto.Category), Value: b}); err != nil {
		slog.Error("error sending to kafka:",
			slog.String("topic", p.articlesWriter.Topic),
			slog.String("error", err.Error()),
		)
	}

}

func (p KafkaProducer) SendNews(dto domain.NewsDto) {
	b, err := json.Marshal(dto)
	if err != nil {
		slog.Error(err.Error())
	}
	ctx, cancel := context.WithTimeout(context.Background(), kafkaSendTimeout)
	defer cancel()
	if err = p.newsWriter.WriteMessages(ctx, kafka.Message{Key: []byte(dto.Category), Value: b}); err != nil {
		slog.Error("error sending to kafka:",
			slog.String("topic", p.newsWriter.Topic),
			slog.String("error", err.Error()),
		)
	}
}

func (p KafkaProducer) Close() error {
	slog.Info("closing kafka producer")
	errs := make([]error, 2)
	errs = append(errs, p.articlesWriter.Close())
	errs = append(errs, p.newsWriter.Close())
	return errors.Join(errs...)
}

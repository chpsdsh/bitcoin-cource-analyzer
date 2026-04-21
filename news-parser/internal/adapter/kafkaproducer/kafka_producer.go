package kafkaproducer

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"time"

	"news-parser/internal/domain"
	"news-parser/internal/observability"

	"github.com/segmentio/kafka-go"
)

const (
	kafkaSendTimeout     = time.Second * 15
	kafkaErrsArrCapacity = 2
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

func (p KafkaProducer) SendArticle(ctx context.Context, dto domain.ArticleDto) {
	b, err := json.Marshal(dto)
	if err != nil {
		slog.Error(err.Error(), slog.String("trace_id", dto.TraceID))
	}
	ctx, cancel := context.WithTimeout(ctx, kafkaSendTimeout)
	defer cancel()
	msg := kafka.Message{Key: []byte(dto.Category), Value: b, Headers: observability.KafkaHeadersFromContext(ctx)}
	if err = p.articlesWriter.WriteMessages(ctx, msg); err != nil {
		slog.Error("error sending to kafka:",
			slog.String("trace_id", dto.TraceID),
			slog.String("topic", p.articlesWriter.Topic),
			slog.String("error", err.Error()),
		)
	}

}

func (p KafkaProducer) SendNews(ctx context.Context, dto domain.NewsDto) {
	b, err := json.Marshal(dto)
	if err != nil {
		slog.Error(err.Error(), slog.String("trace_id", dto.TraceID))
	}
	ctx, cancel := context.WithTimeout(ctx, kafkaSendTimeout)
	defer cancel()
	msg := kafka.Message{Key: []byte(dto.Category), Value: b, Headers: observability.KafkaHeadersFromContext(ctx)}
	if err = p.newsWriter.WriteMessages(ctx, msg); err != nil {
		slog.Error("error sending to kafka:",
			slog.String("trace_id", dto.TraceID),
			slog.String("topic", p.newsWriter.Topic),
			slog.String("error", err.Error()),
		)
	}
}

func (p KafkaProducer) Close() error {
	slog.Info("closing kafka producer")
	errs := make([]error, 0, kafkaErrsArrCapacity)
	errs = append(errs, p.articlesWriter.Close())
	errs = append(errs, p.newsWriter.Close())
	return errors.Join(errs...)
}

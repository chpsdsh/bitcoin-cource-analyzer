package kafkaproducer

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/segmentio/kafka-go"

	"data-cleaner/internal/adapter/config"
	"data-cleaner/internal/domain"
)

const (
	kafkaSendTimeout = time.Second * 15
)

type KafkaProducer struct {
	articlesWriter *kafka.Writer
}

func NewKafkaProducer(config config.KafkaConfig) KafkaProducer {
	articlesWriter := &kafka.Writer{
		Addr:                   kafka.TCP(config.Brokers...),
		Topic:                  config.OutputTopic,
		Balancer:               &kafka.Hash{},
		RequiredAcks:           kafka.RequireAll,
		AllowAutoTopicCreation: true,
	}

	return KafkaProducer{articlesWriter: articlesWriter}
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

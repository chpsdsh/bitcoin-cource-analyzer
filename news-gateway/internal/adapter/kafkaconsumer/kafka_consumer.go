package kafkaconsumer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/segmentio/kafka-go"

	"news-gateway/internal/adapter/config"
	"news-gateway/internal/domain"
)

type StorageSender interface {
	AddNews(ctx context.Context, dto domain.NewsDto, key string) error
}

type KafkaConsumer struct {
	reader *kafka.Reader
	sender StorageSender
}

func NewKafkaConsumer(conf config.KafkaConfig, sender StorageSender) *KafkaConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: conf.Brokers,
		GroupID: conf.GroupID,
		Topic:   conf.Topic,
	})
	return &KafkaConsumer{reader: reader, sender: sender}
}

func (k *KafkaConsumer) StartReadingNews(ctx context.Context) error {
	for {
		msg, err := k.reader.ReadMessage(ctx)
		if err != nil {
			return fmt.Errorf("error reading message: %w", err)
		}

		newsDto := domain.NewsDto{}
		if err = json.Unmarshal(msg.Value, &newsDto); err != nil {
			return fmt.Errorf("error unmarshalling message: %w", err)
		}

		if err = k.sender.AddNews(ctx, newsDto, newsDto.Category); err != nil {
			return fmt.Errorf("error sending news: %w", err)
		}
	}
}

func (k *KafkaConsumer) Close() error {
	if err := k.reader.Close(); err != nil {
		return fmt.Errorf("close kafka reader: %w", err)
	}
	return nil
}

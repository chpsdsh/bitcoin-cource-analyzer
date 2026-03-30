package kafkaconsumer

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/segmentio/kafka-go"

	"news-gateway/internal/domain"
)

const (
	responseMessagesCount = 20
	kafkaRequestDuration  = 10 * time.Second
)

type KafkaConsumer struct {
	reader *kafka.Reader
}

func NewKafkaConsumer() *KafkaConsumer {
	brokers := os.Getenv("KAFKA_BROKERS")
	newsTopic := os.Getenv("KAFKA_NEWS_TOPIC")
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{brokers},
		GroupID: consumerGroupName,
		Topic:   newsTopic,
	})
	return &KafkaConsumer{reader: reader}
}

func (k *KafkaConsumer) ReadNews(ctx context.Context) ([]domain.NewsDto, error) {
	newsDtoArr := make([]domain.NewsDto, 0, responseMessagesCount)
	ctx, cancel := context.WithTimeout(ctx, kafkaRequestDuration)
	defer cancel()
	for range responseMessagesCount {
		msg, err := k.reader.ReadMessage(ctx)
		if err != nil {
			return nil, fmt.Errorf("error reading message: %w", err)
		}
		newsDto := domain.NewsDto{}
		if err = json.Unmarshal(msg.Value, &newsDto); err != nil {
			return nil, fmt.Errorf("error unmarshalling message: %w", err)
		}
		newsDtoArr = append(newsDtoArr, newsDto)
	}
	return newsDtoArr, nil
}

func (k *KafkaConsumer) Close() error {
	if err := k.reader.Close(); err != nil {
		return fmt.Errorf("close kafka reader: %w", err)
	}
	return nil
}

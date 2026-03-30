package kafkaconsumer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/segmentio/kafka-go"

	"data-cleaner/internal/adapter/config"
	"data-cleaner/internal/domain"
)

type ArticlesHandler interface {
	HandleNewArticle(ctx context.Context, article domain.ArticleDto)
}

type KafkaConsumer struct {
	reader  *kafka.Reader
	handler ArticlesHandler
}

func NewKafkaConsumer(conf config.KafkaConfig, sender ArticlesHandler) *KafkaConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: conf.Brokers,
		GroupID: conf.GroupID,
		Topic:   conf.InputTopic,
	})
	return &KafkaConsumer{reader: reader, handler: sender}
}

func (k *KafkaConsumer) StartReadingArticles(ctx context.Context) error {
	for {
		msg, err := k.reader.ReadMessage(ctx)
		if err != nil {
			return fmt.Errorf("error reading message: %w", err)
		}

		articleDto := domain.ArticleDto{}
		if err = json.Unmarshal(msg.Value, &articleDto); err != nil {
			return fmt.Errorf("error unmarshalling message: %w", err)
		}

		k.handler.HandleNewArticle(ctx, articleDto)
	}
}

func (k *KafkaConsumer) Close() error {
	if err := k.reader.Close(); err != nil {
		return fmt.Errorf("close kafka reader: %w", err)
	}
	return nil
}

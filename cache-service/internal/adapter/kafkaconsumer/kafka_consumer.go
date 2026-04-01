package kafkaconsumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/segmentio/kafka-go"

	"data-cleaner/internal/adapter/config"
	"data-cleaner/internal/domain"
)

type ArticlesHandler interface {
	HandleArticle(ctx context.Context, article domain.ArticleDto)
	HandleNews(ctx context.Context, news domain.NewsDto)
}

type KafkaConsumer struct {
	articlesReader *kafka.Reader
	newsReader     *kafka.Reader
	handler        ArticlesHandler
}

func NewKafkaConsumer(conf config.KafkaConfig, sender ArticlesHandler) *KafkaConsumer {
	articlesReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: conf.Brokers,
		GroupID: conf.GroupID,
		Topic:   conf.InputArticlesTopic,
	})
	newsReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: conf.Brokers,
		GroupID: conf.GroupID,
		Topic:   conf.InputNewsTopic,
	})
	return &KafkaConsumer{articlesReader: articlesReader, newsReader: newsReader, handler: sender}
}

func (k *KafkaConsumer) StartReadingArticles(ctx context.Context) error {
	for {
		msg, err := k.articlesReader.ReadMessage(ctx)
		if err != nil {
			return fmt.Errorf("error reading message: %w", err)
		}

		articleDto := domain.ArticleDto{}
		if err = json.Unmarshal(msg.Value, &articleDto); err != nil {
			return fmt.Errorf("error unmarshalling message: %w", err)
		}
		slog.Info("new articles dto", slog.Any("articles", articleDto))
		k.handler.HandleArticle(ctx, articleDto)
	}
}

func (k *KafkaConsumer) StartReadingNews(ctx context.Context) error {
	for {
		msg, err := k.newsReader.ReadMessage(ctx)
		if err != nil {
			return fmt.Errorf("error reading message: %w", err)
		}

		newsDto := domain.NewsDto{}
		if err = json.Unmarshal(msg.Value, &newsDto); err != nil {
			return fmt.Errorf("error unmarshalling message: %w", err)
		}
		slog.Info("new articles dto", slog.Any("news", newsDto))
		k.handler.HandleNews(ctx, newsDto)
	}
}

func (k *KafkaConsumer) Close() error {
	if err := k.articlesReader.Close(); err != nil {
		return fmt.Errorf("close kafka articles reader: %w", err)
	}
	if err := k.newsReader.Close(); err != nil {
		return fmt.Errorf("close kafka news reader: %w", err)
	}
	return nil
}

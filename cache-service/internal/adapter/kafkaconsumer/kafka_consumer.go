package kafkaconsumer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/segmentio/kafka-go"

	"data-cleaner/internal/adapter/config"
	"data-cleaner/internal/domain"
	"data-cleaner/internal/observability"
)

type ArticlesHandler interface {
	HandleArticle(ctx context.Context, article domain.ArticleDto)
	HandleNews(ctx context.Context, news domain.NewsDto)
	HandleLLMResponse(ctx context.Context, response domain.LLMResponse)
}

type KafkaConsumer struct {
	articlesReader    *kafka.Reader
	newsReader        *kafka.Reader
	llmResponseReader *kafka.Reader
	handler           ArticlesHandler
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
	llmResponseReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: conf.Brokers,
		GroupID: conf.GroupID,
		Topic:   conf.InputLLMResponseTopic,
	})

	return &KafkaConsumer{articlesReader: articlesReader,
		newsReader:        newsReader,
		llmResponseReader: llmResponseReader,
		handler:           sender,
	}
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
		msgCtx := observability.ContextWithKafkaTraceID(ctx, msg.Headers)
		if articleDto.TraceID == "" {
			articleDto.TraceID = observability.TraceIDFromContext(msgCtx)
		}
		k.handler.HandleArticle(msgCtx, articleDto)
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
		msgCtx := observability.ContextWithKafkaTraceID(ctx, msg.Headers)
		if newsDto.TraceID == "" {
			newsDto.TraceID = observability.TraceIDFromContext(msgCtx)
		}
		k.handler.HandleNews(msgCtx, newsDto)
	}
}

func (k *KafkaConsumer) StartLLMResponse(ctx context.Context) error {
	for {
		msg, err := k.llmResponseReader.ReadMessage(ctx)
		if err != nil {
			return fmt.Errorf("error reading message: %w", err)
		}

		llmDto := domain.LLMResponse{}
		if err = json.Unmarshal(msg.Value, &llmDto); err != nil {
			return fmt.Errorf("error unmarshalling message: %w", err)
		}
		msgCtx := observability.ContextWithKafkaTraceID(ctx, msg.Headers)
		if llmDto.TraceID == "" {
			llmDto.TraceID = observability.TraceIDFromContext(msgCtx)
		}
		k.handler.HandleLLMResponse(msgCtx, llmDto)
	}
}

func (k *KafkaConsumer) Close() error {
	if err := k.articlesReader.Close(); err != nil {
		return fmt.Errorf("close kafka articles reader: %w", err)
	}
	if err := k.newsReader.Close(); err != nil {
		return fmt.Errorf("close kafka news reader: %w", err)
	}
	if err := k.llmResponseReader.Close(); err != nil {
		return fmt.Errorf("close kafka llm response reader: %w", err)
	}
	return nil
}

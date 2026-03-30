package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"data-cleaner/internal/adapter/config"
	"data-cleaner/internal/adapter/kafkaconsumer"
	"data-cleaner/internal/adapter/kafkaproducer"
	"data-cleaner/internal/adapter/storage"
	"data-cleaner/internal/application"
)

func main() {
	kafkaCfg, err := config.NewKafkaConfig()
	if err != nil {
		slog.Error("Error loading kafka config", slog.String("error", err.Error()))
		os.Exit(1)
	}

	redisConf, err := config.NewRedisConfig()
	if err != nil {
		slog.Error("Error loading redis config", slog.String("error", err.Error()))
		os.Exit(1)
	}
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	articlesStorage := storage.NewDataStorage(redisConf)
	producer := kafkaproducer.NewKafkaProducer(kafkaCfg)
	dedupService := application.DedupService{Sender: producer, Storage: articlesStorage}

	consumer := kafkaconsumer.NewKafkaConsumer(kafkaCfg, dedupService)

	go func() {
		if err = consumer.StartReadingArticles(ctx); err != nil {
			cancel()
		}
	}()
	<-ctx.Done()
	
	if err = articlesStorage.CloseRedis(); err != nil {
		slog.Error("Error closing redis connection", slog.String("error", err.Error()))
	}

}

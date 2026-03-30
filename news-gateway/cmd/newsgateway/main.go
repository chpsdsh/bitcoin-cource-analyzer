package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"news-gateway/internal/adapter/config"
	"news-gateway/internal/adapter/kafkaconsumer"
	"news-gateway/internal/adapter/server"
	"news-gateway/internal/adapter/storage"
	"news-gateway/internal/application"
)

const (
	serverAddress    = ":8080"
	shutdownDuration = 5 * time.Second
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

	newsStorage := storage.NewNewsStorage(redisConf)
	newsService := application.NewsService{Storage: newsStorage}

	kafkaReader := kafkaconsumer.NewKafkaConsumer(kafkaCfg, newsStorage)

	newsServer := server.NewNewsServer(newsService)
	router := server.NewRouter(newsServer)
	newsHTTPServer := &http.Server{
		Addr:    serverAddress,
		Handler: router,
	}

	go func(ctx context.Context) {
		if err = kafkaReader.StartReadingNews(ctx); err != nil {
			slog.Error("Error starting kafka reader", slog.String("error", err.Error()))
			cancel()
		}
	}(ctx)

	go func() {
		if err = newsHTTPServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("listen err:", slog.String("error", err.Error()))
		}
	}()

	<-ctx.Done()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownDuration)
	defer shutdownCancel()

	if err = newsHTTPServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutdown err:", slog.String("error", err.Error()))
	}
	if err = kafkaReader.Close(); err != nil {
		slog.Error("Error closing kafka reader", slog.String("error", err.Error()))
	}
	if err = newsStorage.CloseRedis(); err != nil {
		slog.Error("close redis storage", slog.String("error", err.Error()))
	}
}

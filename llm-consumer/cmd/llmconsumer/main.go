package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"llm-consumer/internal/adapter/config"
	"llm-consumer/internal/adapter/server"
	"llm-consumer/internal/adapter/storage"
	"llm-consumer/internal/application"
)

const (
	serverAddress    = ":8080"
	shutdownDuration = 5 * time.Second
)

func main() {
	redisConf, err := config.NewRedisConfig()
	if err != nil {
		slog.Error("Error loading redis config", slog.String("error", err.Error()))
		os.Exit(1)
	}
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	llmStorage := storage.NewLLMStorage(redisConf)
	predictionService := application.NewPredictionService(llmStorage)

	predictionServer := server.Server{Predictor: predictionService}
	router := server.NewRouter(predictionServer)
	predictorHTTPServer := http.Server{
		Addr:    serverAddress,
		Handler: router,
	}

	go func() {
		if err = predictorHTTPServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("listen err:", slog.String("error", err.Error()))
			cancel()
		}
	}()

	<-ctx.Done()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownDuration)
	defer shutdownCancel()

	if err = predictorHTTPServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutdown http server:", slog.String("error", err.Error()))
	}

	if err = llmStorage.CloseRedis(); err != nil {
		slog.Error("shutdown redis:", slog.String("error", err.Error()))
	}
}

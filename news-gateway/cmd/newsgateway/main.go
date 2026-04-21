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
	"news-gateway/internal/adapter/server"
	"news-gateway/internal/adapter/storage"
	"news-gateway/internal/application"
	"news-gateway/internal/observability"
)

const (
	serverAddress    = ":8080"
	shutdownDuration = 5 * time.Second
)

func main() {
	observability.InitLogger("news-gateway")

	redisConf, err := config.NewRedisConfig()
	if err != nil {
		slog.Error("Error loading redis config", slog.String("error", err.Error()))
		os.Exit(1)
	}
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	newsStorage := storage.NewNewsStorage(redisConf)
	newsService := application.NewsService{Storage: newsStorage}

	newsServer := server.NewNewsServer(newsService)
	router := server.NewRouter(newsServer)
	newsHTTPServer := &http.Server{
		Addr:    serverAddress,
		Handler: router,
	}

	go func() {
		if err = newsHTTPServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("listen err:", slog.String("error", err.Error()))
			cancel()
		}
	}()

	<-ctx.Done()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownDuration)
	defer shutdownCancel()

	if err = newsHTTPServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutdown err:", slog.String("error", err.Error()))
	}

	if err = newsStorage.CloseRedis(); err != nil {
		slog.Error("close redis storage", slog.String("error", err.Error()))
	}
}

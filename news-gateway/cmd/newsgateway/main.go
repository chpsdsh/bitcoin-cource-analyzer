package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"news-gateway/internal/adapter/kafkaconsumer"
	"news-gateway/internal/adapter/server"
)

const (
	serverAddress    = ":8080"
	shutdownDuration = 5 * time.Second
)

func main() {
	kafkaReader := kafkaconsumer.NewKafkaConsumer()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	newsServer := server.NewNewsServer(kafkaReader)
	router := server.NewRouter(newsServer)

	newsHTTPServer := &http.Server{
		Addr:    serverAddress,
		Handler: router,
	}

	go func() {
		if err := newsHTTPServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("listen err:", slog.String("error", err.Error()))
		}
	}()

	<-ctx.Done()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownDuration)
	defer shutdownCancel()
	if err := newsHTTPServer.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("server shutdown: %s\n", err)
	}
	if err := kafkaReader.Close(); err != nil {
		log.Fatalf("kafka reader: %s\n", err)
	}
}

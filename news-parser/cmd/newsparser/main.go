package main

import (
	"context"
	"log/slog"
	"net/http"
	"news-parser/internal/application"
	"news-parser/internal/application/readers"
	"news-parser/internal/domain"
	"news-parser/internal/infrastructure"
	"news-parser/internal/infrastructure/kafkaproducer"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"time"
)

func main() {
	client := &http.Client{Timeout: 40 * time.Second}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	wg := &sync.WaitGroup{}
	newsChan := make(chan domain.GdeltApiDto, application.NewsRequestsCount)
	kafkaSendChan := make(chan domain.ResultDto, application.NewsRequestsCount)

	kafkaProducer := kafkaproducer.NewKafkaProducer()

	pool := readers.WorkerPool{Wg: wg}
	requester := infrastructure.NewsRequester{Client: client}
	pool.StartWorkers(ctx, newsChan, kafkaSendChan, requester)
	app := application.Application{Ticker: time.NewTicker(5 * time.Second), RequestHandler: requester, RequestChan: newsChan}
	wg.Go(func() { app.StartParsingNews(ctx) })

	kafkaproducer.NewSenderPool(wg, ctx, kafkaSendChan, kafkaProducer)

	<-ctx.Done()
	wg.Wait()
	close(kafkaSendChan)
	if err := kafkaProducer.Close(); err != nil {
		slog.Error("error closing kafka producer", "err", err)
	}
}

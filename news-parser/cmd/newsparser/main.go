package main

import (
	"context"
	"log/slog"
	"net/http"
	"news-parser/internal/adapter/kafkaproducer"
	"news-parser/internal/adapter/networkclient"
	"news-parser/internal/application"
	"news-parser/internal/application/readers"
	"news-parser/internal/domain"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"time"
)

const (
	clientTimeout = time.Second * 30
	tickerTimeout = time.Second * 10
)

func main() {
	llmAddr := os.Getenv("LLM_ADDRESS")
	if llmAddr == "" {
		slog.Error("LLM_ADDRESS environment variable not set")
		os.Exit(1)
	}

	client := &http.Client{Timeout: clientTimeout}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	wg := &sync.WaitGroup{}
	newsChan := make(chan domain.GdeltApiDto, application.NewsRequestsCount)
	kafkaArticlesSendChan := make(chan domain.ArticleDto, application.NewsRequestsCount)
	kafkaNewsSendChan := make(chan domain.NewsDto, application.NewsRequestsCount)

	kafkaProducer := kafkaproducer.NewKafkaProducer()

	pool := readers.WorkerPool{Wg: wg}
	newsRequester := networkclient.NewsRequester{Client: client}
	llmNotifier := networkclient.LLMClient{Client: client, LLMAddress: llmAddr}

	pool.StartWorkers(ctx, newsChan, kafkaArticlesSendChan, newsRequester)
	app := application.Application{Ticker: time.NewTicker(tickerTimeout),
		RequestHandler: newsRequester,
		LLMNotifier:    llmNotifier,
		RequestChan:    newsChan,
		NewsChan:       kafkaNewsSendChan}
	wg.Go(func() { app.StartParsingNews(ctx) })

	kafkaproducer.NewSenderPool(ctx, wg, kafkaArticlesSendChan, kafkaNewsSendChan, kafkaProducer)

	<-ctx.Done()
	wg.Wait()
	close(kafkaArticlesSendChan)
	if err := kafkaProducer.Close(); err != nil {
		slog.Error("error closing kafka producer", "err", err)
	}
}

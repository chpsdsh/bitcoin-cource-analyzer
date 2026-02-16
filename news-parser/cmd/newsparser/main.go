package main

import (
	"context"
	"net/http"
	"news-parser/internal/application"
	"news-parser/internal/application/threadpool"
	"news-parser/internal/domain"
	"news-parser/internal/infrastructure"
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

	pool := threadpool.WorkerPool{Wg: wg}
	requester := infrastructure.NewsRequester{Client: client}
	pool.StartWorkers(ctx, newsChan, requester)
	app := application.Application{Ticker: time.NewTicker(5 * time.Second), RequestHandler: requester, RequestChan: newsChan}
	app.StartParsingNews(ctx)

	<-ctx.Done()
	wg.Wait()
}

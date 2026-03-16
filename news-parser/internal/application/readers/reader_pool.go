package readers

import (
	"bytes"
	"context"
	"log/slog"
	"net/url"
	"news-parser/internal/application"
	"news-parser/internal/application/utils"
	"news-parser/internal/domain"
	"sync"

	"github.com/go-shiori/go-readability"
)

const NumWorkers = 10

type WorkerPool struct {
	Wg      *sync.WaitGroup
	workers []*worker
}

type worker struct {
	tasks   chan domain.GdeltApiDto
	results chan domain.ResultDto
}

func (w *worker) work(ctx context.Context, handler application.RequestHandler) {
	for {
		select {
		case task, ok := <-w.tasks:
			if !ok {
				return
			}
			requestUrl, err := url.Parse(task.Url)
			if err != nil {
				continue
			}
			data, err := handler.DoDataRequest(task.Url)
			if err != nil {
				slog.Error("error requesting data", "url:", task.Url, "err:", err)
				continue
			}
			article, err := readability.FromReader(bytes.NewReader(data), requestUrl)
			if err != nil {
				slog.Error("error getting data from html", "url:", task.Url, "err:", err)
				continue
			}
			w.results <- domain.ResultDto{Category: task.Category, Title: task.Title, Text: utils.NormalizeText(article.TextContent)}
		case <-ctx.Done():
			return
		}
	}
}

func (pool *WorkerPool) StartWorkers(ctx context.Context,
	tasks chan domain.GdeltApiDto,
	results chan domain.ResultDto,
	handler application.RequestHandler) {
	pool.workers = make([]*worker, NumWorkers)
	for i := 0; i < NumWorkers; i++ {
		pool.workers[i] = &worker{tasks: tasks, results: results}
		pool.Wg.Go(func() {
			pool.workers[i].work(ctx, handler)
		})
	}
}

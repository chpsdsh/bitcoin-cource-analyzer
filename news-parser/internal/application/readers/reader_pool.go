package readers

import (
	"bytes"
	"context"
	"log/slog"
	"net/url"
	"sync"

	"news-parser/internal/application"
	"news-parser/internal/application/utils"
	"news-parser/internal/domain"

	"github.com/go-shiori/go-readability"
)

const NumWorkers = 10

type WorkerPool struct {
	Wg      *sync.WaitGroup
	workers []*worker
}

type worker struct {
	tasks   chan domain.GdeltAPIDto
	results chan domain.ArticleDto
}

func (w *worker) work(ctx context.Context, handler application.RequestHandler) {
	for {
		select {
		case task, ok := <-w.tasks:
			if !ok {
				return
			}
			requestURL, err := url.Parse(task.URL)
			if err != nil {
				continue
			}
			data, err := handler.DoDataRequest(task.URL)
			if err != nil {
				slog.Error("error requesting data", "url:", task.URL, "err:", err)
				continue
			}
			article, err := readability.FromReader(bytes.NewReader(data), requestURL)
			if err != nil {
				slog.Error("error getting data from html", "url:", task.URL, "err:", err)
				continue
			}
			w.results <- domain.ArticleDto{Category: task.Category, Title: task.Title, Text: utils.NormalizeText(article.TextContent)}
		case <-ctx.Done():
			return
		}
	}
}

func (pool *WorkerPool) StartWorkers(ctx context.Context,
	tasks chan domain.GdeltAPIDto,
	results chan domain.ArticleDto,
	handler application.RequestHandler) {
	pool.workers = make([]*worker, NumWorkers)
	for i := range NumWorkers {
		pool.workers[i] = &worker{tasks: tasks, results: results}
		pool.Wg.Go(func() {
			pool.workers[i].work(ctx, handler)
		})
	}
}

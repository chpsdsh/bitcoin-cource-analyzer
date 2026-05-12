package observability

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"sync"
	"time"
)

const metricsShutdownTimeout = 5 * time.Second

var gdeltMetrics = struct {
	sync.Mutex
	attempts map[string]int64
}{
	attempts: make(map[string]int64),
}

func ObserveGDELTAttempt(category, result string, statusCode int) {
	status := "transport_error"
	if statusCode > 0 {
		status = strconv.Itoa(statusCode)
	}

	key := fmt.Sprintf(`category="%s",result="%s",status_code="%s"`, category, result, status)
	gdeltMetrics.Lock()
	defer gdeltMetrics.Unlock()
	gdeltMetrics.attempts[key]++
}

func StartMetricsServer(ctx context.Context, address string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		_, _ = fmt.Fprintln(w, "# HELP news_parser_gdelt_api_attempts_total Total number of public GDELT API attempts grouped by category, result, and HTTP status.")
		_, _ = fmt.Fprintln(w, "# TYPE news_parser_gdelt_api_attempts_total counter")

		gdeltMetrics.Lock()
		defer gdeltMetrics.Unlock()
		for key, value := range gdeltMetrics.attempts {
			_, _ = fmt.Fprintf(w, "news_parser_gdelt_api_attempts_total{%s} %d\n", key, value)
		}
	})

	server := &http.Server{
		Addr:              address,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), metricsShutdownTimeout)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			slog.Error("shutdown metrics server", slog.String("error", err.Error()))
		}
	}()

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("metrics server failed", slog.String("error", err.Error()))
		}
	}()
}

package observability

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"
)

var predictionMetrics = struct {
	sync.Mutex
	attempts map[string]int64
}{
	attempts: make(map[string]int64),
}

func MetricsHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		_, _ = fmt.Fprintln(w, "# HELP llm_consumer_prediction_attempts_total Total number of prediction endpoint attempts grouped by result and HTTP status.")
		_, _ = fmt.Fprintln(w, "# TYPE llm_consumer_prediction_attempts_total counter")

		predictionMetrics.Lock()
		defer predictionMetrics.Unlock()
		for key, value := range predictionMetrics.attempts {
			_, _ = fmt.Fprintf(w, "llm_consumer_prediction_attempts_total{%s} %d\n", key, value)
		}
	})
}

func ObservePredictionAttempt(statusCode int) {
	result := "failure"
	if statusCode >= http.StatusOK && statusCode < http.StatusBadRequest {
		result = "success"
	}

	key := fmt.Sprintf(`result="%s",status_code="%s"`, result, strconv.Itoa(statusCode))
	predictionMetrics.Lock()
	defer predictionMetrics.Unlock()
	predictionMetrics.attempts[key]++
}

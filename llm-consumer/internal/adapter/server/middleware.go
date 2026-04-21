package server

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"

	"llm-consumer/internal/observability"
)

func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := c.GetHeader(observability.TraceIDHeader)
		if traceID == "" {
			traceID = observability.NewTraceID()
		}

		ctx := observability.ContextWithTraceID(c.Request.Context(), traceID)
		c.Request = c.Request.WithContext(ctx)
		c.Header(observability.TraceIDHeader, traceID)

		startedAt := time.Now()
		c.Next()

		slog.Info("http request completed",
			slog.String("trace_id", traceID),
			slog.String("method", c.Request.Method),
			slog.String("path", c.FullPath()),
			slog.Int("status", c.Writer.Status()),
			slog.Int64("duration_ms", time.Since(startedAt).Milliseconds()),
		)
	}
}

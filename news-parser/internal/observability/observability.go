package observability

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"os"

	"github.com/segmentio/kafka-go"
)

const (
	TraceIDHeader      = "X-Request-ID"
	KafkaTraceIDHeader = "trace_id"
	traceIDKey         = contextKey("trace_id")
)

type contextKey string

func InitLogger(service string) {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	slog.SetDefault(slog.New(handler).With(slog.String("service", service)))
}

func NewTraceID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "trace-id-unavailable"
	}
	return hex.EncodeToString(b[:])
}

func ContextWithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDKey, traceID)
}

func TraceIDFromContext(ctx context.Context) string {
	traceID, _ := ctx.Value(traceIDKey).(string)
	return traceID
}

func KafkaHeadersFromContext(ctx context.Context) []kafka.Header {
	traceID := TraceIDFromContext(ctx)
	if traceID == "" {
		return nil
	}
	return []kafka.Header{{Key: KafkaTraceIDHeader, Value: []byte(traceID)}}
}

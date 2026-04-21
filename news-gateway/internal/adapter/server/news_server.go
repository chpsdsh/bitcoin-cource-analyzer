package server

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"news-gateway/internal/domain"
	"news-gateway/internal/observability"
)

type Reader interface {
	RequestNews(ctx context.Context, key string) ([]domain.NewsDto, error)
}

type NewsServer struct {
	reader Reader
}

func NewNewsServer(reader Reader) NewsServer {
	return NewsServer{reader: reader}
}

func (s NewsServer) GetNews(c *gin.Context) {
	traceID := observability.TraceIDFromContext(c.Request.Context())
	category := c.Param("category")
	news, err := s.reader.RequestNews(c.Request.Context(), category)

	if err != nil {
		slog.Error("failed to get news",
			slog.String("trace_id", traceID),
			slog.String("category", category),
			slog.String("error", err.Error()),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	slog.Info("news returned",
		slog.String("trace_id", traceID),
		slog.String("category", category),
		slog.Int("count", len(news)),
	)
	c.JSON(http.StatusOK, gin.H{"news": news})
}

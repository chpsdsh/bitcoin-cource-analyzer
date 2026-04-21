package server

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"llm-consumer/internal/domain"
	"llm-consumer/internal/observability"
)

type Predictor interface {
	ValidateCategories(categories domain.Categories) error
	DoPrediction(ctx context.Context, categories domain.Categories) (domain.Prediction, error)
}

type Server struct {
	Predictor Predictor
}

func (s *Server) HandlePrediction(c *gin.Context) {
	traceID := observability.TraceIDFromContext(c.Request.Context())
	var categories domain.Categories
	if err := c.ShouldBindJSON(&categories); err != nil {
		slog.Error("invalid prediction request",
			slog.String("trace_id", traceID),
			slog.String("error", err.Error()),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.Predictor.ValidateCategories(categories); err != nil {
		slog.Error("invalid prediction categories",
			slog.String("trace_id", traceID),
			slog.Any("categories", categories.CategoriesList),
			slog.String("error", err.Error()),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	prediction, err := s.Predictor.DoPrediction(c.Request.Context(), categories)
	if err != nil {
		slog.Error("prediction failed",
			slog.String("trace_id", traceID),
			slog.Any("categories", categories.CategoriesList),
			slog.String("error", err.Error()),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	slog.Info("prediction returned",
		slog.String("trace_id", traceID),
		slog.Any("categories", categories.CategoriesList),
	)
	c.JSON(http.StatusOK, prediction)
}

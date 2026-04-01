package server

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"llm-consumer/internal/domain"
)

type Predictor interface {
	ValidateCategories(categories domain.Categories) error
	DoPrediction(ctx context.Context, categories domain.Categories) (domain.Prediction, error)
}

type Server struct {
	Predictor Predictor
}

func (s *Server) HandlePrediction(c *gin.Context) {
	var categories domain.Categories
	if err := c.ShouldBindJSON(&categories); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.Predictor.ValidateCategories(categories); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	prediction, err := s.Predictor.DoPrediction(c.Request.Context(), categories)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, prediction)
}

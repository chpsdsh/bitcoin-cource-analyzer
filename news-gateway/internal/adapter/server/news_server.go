package server

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"news-gateway/internal/domain"
)

var ErrNoCategoryPathParameter = errors.New("query should be news/{category}")

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
	category := c.Param("category")
	if category == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": ErrNoCategoryPathParameter})
		return
	}

	news, err := s.reader.RequestNews(c.Request.Context(), category)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	c.JSON(http.StatusOK, gin.H{"news": news})
}

package server

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"news-gateway/internal/domain"
)

type Reader interface {
	RequestNews(ctx context.Context) ([]domain.NewsDto, error)
}
type NewsServer struct {
	reader Reader
}

func NewNewsServer(reader Reader) NewsServer {
	return NewsServer{reader: reader}
}

func (s NewsServer) GetNews(c *gin.Context) {
	news, err := s.reader.RequestNews(c.Request.Context())

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	c.JSON(http.StatusOK, gin.H{"news": news})
}

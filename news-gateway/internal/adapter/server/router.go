package server

import (
	"github.com/gin-gonic/gin"
)

const (
	newsEndpoint = "/news/:category"
)

func NewRouter(server NewsServer) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery(), RequestIDMiddleware())
	r.GET(newsEndpoint, server.GetNews)
	return r
}

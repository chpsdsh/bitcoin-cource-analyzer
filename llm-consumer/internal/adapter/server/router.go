package server

import (
	"github.com/gin-gonic/gin"

	"llm-consumer/internal/observability"
)

const predictionEndpoint = "/predict"

func NewRouter(server Server) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery(), RequestIDMiddleware(), PredictionMetricsMiddleware())
	router.POST(predictionEndpoint, server.HandlePrediction)
	router.GET("/metrics", gin.WrapH(observability.MetricsHandler()))
	return router
}

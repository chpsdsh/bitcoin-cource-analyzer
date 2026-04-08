package server

import "github.com/gin-gonic/gin"

const predictionEndpoint = "/predict"

func NewRouter(server Server) *gin.Engine {
	router := gin.New()
	router.POST(predictionEndpoint, server.HandlePrediction)
	return router
}

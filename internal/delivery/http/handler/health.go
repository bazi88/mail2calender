package handler

import (
	"net/http"

	"mono-golang/internal/infrastructure/logger"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Health godoc
// @Summary Health check
// @Description Check if the service is healthy
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /health [get]
func Health(c *gin.Context) {
	log := logger.GetLogger()
	log.WithFields(logrus.Fields{
		"path":   c.Request.URL.Path,
		"method": c.Request.Method,
		"ip":     c.ClientIP(),
	}).Info("Health check requested")

	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
	})
}

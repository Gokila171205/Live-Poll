package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"live-poll-backend/internal/services"
)

// HealthHandler handles incoming health check HTTP requests.
type HealthHandler struct {
	service services.HealthService
}

// NewHealthHandler returns a new instance of HealthHandler.
func NewHealthHandler(service services.HealthService) *HealthHandler {
	return &HealthHandler{service: service}
}

// GetHealth handles GET /api/health and returns the status of the backend, MongoDB, and Redis.
func (h *HealthHandler) GetHealth(c *gin.Context) {
	report := h.service.CheckHealth(c.Request.Context())

	statusCode := http.StatusOK
	if report.Status == "degraded" {
		// Return 200 so callers can inspect which dependency is offline,
		// while clearly indicating overall operational status in the body.
		statusCode = http.StatusOK
	}

	SendSuccess(c, statusCode, report)
}

package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"live-poll-backend/internal/handlers"
	"live-poll-backend/internal/services"
)

// mockHealthRepo implements repository.HealthRepository for unit testing.
type mockHealthRepo struct {
	mongoErr error
	redisErr error
}

func (m *mockHealthRepo) PingMongoDB(ctx context.Context) error {
	return m.mongoErr
}

func (m *mockHealthRepo) PingRedis(ctx context.Context) error {
	return m.redisErr
}

func TestHealthEndpoint_Healthy(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &mockHealthRepo{mongoErr: nil, redisErr: nil}
	healthService := services.NewHealthService(repo)
	healthHandler := handlers.NewHealthHandler(healthService)

	router := gin.New()
	router.GET("/api/health", healthHandler.GetHealth)

	req, _ := http.NewRequest(http.MethodGet, "/api/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp struct {
		Success bool                  `json:"success"`
		Data    services.HealthReport `json:"data"`
	}

	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response JSON: %v", err)
	}

	if !resp.Success {
		t.Errorf("expected success to be true")
	}

	if resp.Data.Status != "healthy" {
		t.Errorf("expected status 'healthy', got '%s'", resp.Data.Status)
	}

	if !resp.Data.Details["mongodb"].Available {
		t.Errorf("expected mongodb to be available")
	}

	if !resp.Data.Details["redis"].Available {
		t.Errorf("expected redis to be available")
	}
}

func TestHealthEndpoint_Degraded(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &mockHealthRepo{
		mongoErr: errors.New("connection refused"),
		redisErr: errors.New("dial tcp: lookup redis failed"),
	}
	healthService := services.NewHealthService(repo)
	healthHandler := handlers.NewHealthHandler(healthService)

	router := gin.New()
	router.GET("/api/health", healthHandler.GetHealth)

	req, _ := http.NewRequest(http.MethodGet, "/api/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp struct {
		Success bool                  `json:"success"`
		Data    services.HealthReport `json:"data"`
	}

	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response JSON: %v", err)
	}

	if resp.Data.Status != "degraded" {
		t.Errorf("expected status 'degraded', got '%s'", resp.Data.Status)
	}

	if resp.Data.Details["mongodb"].Available {
		t.Errorf("expected mongodb to be unavailable")
	}

	if resp.Data.Details["redis"].Available {
		t.Errorf("expected redis to be unavailable")
	}
}

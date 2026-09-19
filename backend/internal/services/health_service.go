package services

import (
	"context"
	"time"

	"live-poll-backend/internal/repository"
)

// ServiceStatus represents the connection status of an individual dependency.
type ServiceStatus struct {
	Available bool   `json:"available"`
	Status    string `json:"status"`
	Error     string `json:"error,omitempty"`
}

// HealthReport encapsulates the health status of the application and its dependencies.
type HealthReport struct {
	Status    string                   `json:"status"`
	Service   string                   `json:"service"`
	Timestamp string                   `json:"timestamp"`
	Uptime    string                   `json:"uptime"`
	Details   map[string]ServiceStatus `json:"dependencies"`
}

// HealthService defines the business logic contract for health inspections.
type HealthService interface {
	CheckHealth(ctx context.Context) HealthReport
}

type healthService struct {
	repo      repository.HealthRepository
	startTime time.Time
}

// NewHealthService instantiates a new HealthService.
func NewHealthService(repo repository.HealthRepository) HealthService {
	return &healthService{
		repo:      repo,
		startTime: time.Now(),
	}
}

// CheckHealth inspects the backend service, MongoDB, and Redis health.
func (s *healthService) CheckHealth(ctx context.Context) HealthReport {
	mongoStatus := ServiceStatus{Available: true, Status: "connected"}
	if err := s.repo.PingMongoDB(ctx); err != nil {
		mongoStatus.Available = false
		mongoStatus.Status = "unavailable"
		mongoStatus.Error = err.Error()
	}

	redisStatus := ServiceStatus{Available: true, Status: "connected"}
	if err := s.repo.PingRedis(ctx); err != nil {
		redisStatus.Available = false
		redisStatus.Status = "unavailable"
		redisStatus.Error = err.Error()
	}

	overallStatus := "healthy"
	if !mongoStatus.Available || !redisStatus.Available {
		overallStatus = "degraded"
	}

	return HealthReport{
		Status:    overallStatus,
		Service:   "live-poll-backend",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Uptime:    time.Since(s.startTime).Round(time.Second).String(),
		Details: map[string]ServiceStatus{
			"mongodb": mongoStatus,
			"redis":   redisStatus,
		},
	}
}

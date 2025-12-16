package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/cotai/tenant-manager/internal/delivery/http/dto"
)

// HealthChecker interface for checking service health
type HealthChecker interface {
	Ping(ctx context.Context) error
}

// HealthHandler handles health check endpoints
type HealthHandler struct {
	db     HealthChecker
	logger *zap.Logger
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(db HealthChecker, logger *zap.Logger) *HealthHandler {
	return &HealthHandler{
		db:     db,
		logger: logger,
	}
}

// Health returns basic service health status
// GET /health
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	response := dto.HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().UTC(),
		Service:   "tenant-manager",
		Version:   "1.0.0",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// Ready returns readiness status (checks dependencies)
// GET /ready
func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	checks := map[string]string{
		"database": "ok",
	}

	// Check database connectivity
	if err := h.db.Ping(ctx); err != nil {
		h.logger.Warn("Database health check failed",
			zap.Error(err),
		)
		checks["database"] = "unavailable"

		response := dto.HealthResponse{
			Status:    "not_ready",
			Timestamp: time.Now().UTC(),
			Service:   "tenant-manager",
			Version:   "1.0.0",
			Checks:    checks,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := dto.HealthResponse{
		Status:    "ready",
		Timestamp: time.Now().UTC(),
		Service:   "tenant-manager",
		Version:   "1.0.0",
		Checks:    checks,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

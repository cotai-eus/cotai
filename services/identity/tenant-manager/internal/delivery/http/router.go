package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	"github.com/cotai/tenant-manager/internal/delivery/http/handler"
	"github.com/cotai/tenant-manager/internal/delivery/http/middleware"
)

// RouterConfig holds router dependencies
type RouterConfig struct {
	TenantHandler *handler.TenantHandler
	HealthHandler *handler.HealthHandler
	AuthMiddleware *middleware.AuthMiddleware
	LoggingMiddleware *middleware.LoggingMiddleware
	RecoveryMiddleware *middleware.RecoveryMiddleware
	CORSMiddleware *middleware.CORSMiddleware
	MetricsMiddleware *middleware.MetricsMiddleware
	Logger *zap.Logger
}

// NewRouter creates and configures the Chi router
func NewRouter(cfg RouterConfig) http.Handler {
	r := chi.NewRouter()

	// ==========================
	// Global Middleware Chain
	// ==========================
	// Order matters: Recovery → Logging → Metrics → CORS → RequestID → RealIP
	r.Use(cfg.RecoveryMiddleware.Handler)
	r.Use(cfg.LoggingMiddleware.Handler)
	r.Use(cfg.MetricsMiddleware.Handler)
	r.Use(cfg.CORSMiddleware.Handler)
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Compress(5))

	// ==========================
	// Public Routes (No Auth)
	// ==========================
	r.Get("/health", cfg.HealthHandler.Health)
	r.Get("/ready", cfg.HealthHandler.Ready)
	r.Get("/metrics", promhttp.Handler().ServeHTTP)

	// ==========================
	// API v1 Routes (Authenticated)
	// ==========================
	r.Route("/api/v1", func(r chi.Router) {
		// Apply authentication to all API routes
		r.Use(cfg.AuthMiddleware.Handler)

		// Tenant Management Routes
		// All require cotai_admin role
		r.Route("/tenants", func(r chi.Router) {
			// Require admin role for all tenant operations
			r.Use(cfg.AuthMiddleware.RequireRole("cotai_admin"))

			r.Post("/", cfg.TenantHandler.CreateTenant)           // POST /api/v1/tenants
			r.Get("/", cfg.TenantHandler.ListTenants)             // GET /api/v1/tenants
			r.Get("/{id}", cfg.TenantHandler.GetTenant)           // GET /api/v1/tenants/{id}
			r.Patch("/{id}", cfg.TenantHandler.UpdateTenant)      // PATCH /api/v1/tenants/{id}
			r.Delete("/{id}", cfg.TenantHandler.DeleteTenant)     // DELETE /api/v1/tenants/{id}

			// Tenant lifecycle operations
			r.Post("/{id}/suspend", cfg.TenantHandler.SuspendTenant)   // POST /api/v1/tenants/{id}/suspend
			r.Post("/{id}/activate", cfg.TenantHandler.ActivateTenant) // POST /api/v1/tenants/{id}/activate
		})
	})

	// ==========================
	// 404 Handler
	// ==========================
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":{"code":"NOT_FOUND","message":"Endpoint not found"}}`))
	})

	// ==========================
	// 405 Method Not Allowed Handler
	// ==========================
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(`{"error":{"code":"METHOD_NOT_ALLOWED","message":"HTTP method not allowed"}}`))
	})

	return r
}

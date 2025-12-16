package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/cotai/tenant-manager/internal/infrastructure/observability"
	"github.com/go-chi/chi/v5"
)

// MetricsMiddleware records HTTP request metrics
type MetricsMiddleware struct {
	metrics *observability.Metrics
}

// NewMetricsMiddleware creates a new metrics middleware
func NewMetricsMiddleware(metrics *observability.Metrics) *MetricsMiddleware {
	return &MetricsMiddleware{
		metrics: metrics,
	}
}

// Handler returns the middleware handler
func (m *MetricsMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap response writer to capture status code
		ww := &metricsResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Process request
		next.ServeHTTP(ww, r)

		// Record metrics
		duration := time.Since(start).Seconds()
		path := chi.RouteContext(r.Context()).RoutePattern()
		if path == "" {
			path = r.URL.Path
		}

		m.metrics.HTTPRequestsTotal.WithLabelValues(
			r.Method,
			path,
			strconv.Itoa(ww.statusCode),
		).Inc()

		m.metrics.HTTPRequestDuration.WithLabelValues(
			r.Method,
			path,
		).Observe(duration)
	})
}

// metricsResponseWriter wraps http.ResponseWriter to capture status code
type metricsResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *metricsResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

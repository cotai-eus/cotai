package observability

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics holds all Prometheus metrics for the tenant manager service
type Metrics struct {
	// Tenant lifecycle metrics
	TenantCreatedTotal       *prometheus.CounterVec
	TenantProvisioningTotal  *prometheus.CounterVec
	TenantSuspendedTotal     *prometheus.CounterVec
	TenantActivatedTotal     *prometheus.CounterVec
	TenantDeletedTotal       *prometheus.CounterVec

	// Provisioning metrics
	ProvisioningDuration     *prometheus.HistogramVec
	ProvisioningErrorsTotal  prometheus.Counter

	// Active tenant gauge
	ActiveTenants            *prometheus.GaugeVec

	// API metrics
	HTTPRequestsTotal        *prometheus.CounterVec
	HTTPRequestDuration      *prometheus.HistogramVec
	GRPCRequestsTotal        *prometheus.CounterVec
	GRPCRequestDuration      *prometheus.HistogramVec

	// Database metrics
	DBConnectionsActive      prometheus.Gauge
	DBQueryDuration          *prometheus.HistogramVec
	DBErrorsTotal            prometheus.Counter

	// Event publishing metrics
	EventsPublishedTotal     *prometheus.CounterVec
	EventPublishingErrors    prometheus.Counter
}

// NewMetrics creates and registers all Prometheus metrics
func NewMetrics() *Metrics {
	return &Metrics{
		// Tenant lifecycle
		TenantCreatedTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "tenant_manager_tenant_created_total",
				Help: "Total number of tenants created",
			},
			[]string{"plan"},
		),
		TenantProvisioningTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "tenant_manager_tenant_provisioning_total",
				Help: "Total number of tenants in provisioning state",
			},
			[]string{"status"}, // success, failed
		),
		TenantSuspendedTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "tenant_manager_tenant_suspended_total",
				Help: "Total number of tenants suspended",
			},
			[]string{"reason"},
		),
		TenantActivatedTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "tenant_manager_tenant_activated_total",
				Help: "Total number of tenants activated or reactivated",
			},
			[]string{"plan"},
		),
		TenantDeletedTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "tenant_manager_tenant_deleted_total",
				Help: "Total number of tenants deleted (soft delete)",
			},
			[]string{"plan"},
		),

		// Provisioning
		ProvisioningDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "tenant_manager_provisioning_duration_seconds",
				Help:    "Duration of tenant schema provisioning",
				Buckets: prometheus.DefBuckets, // Default: [0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10]
			},
			[]string{"status"}, // success, failed
		),
		ProvisioningErrorsTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "tenant_manager_provisioning_errors_total",
				Help: "Total number of provisioning errors",
			},
		),

		// Active tenants
		ActiveTenants: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "tenant_manager_active_tenants",
				Help: "Number of currently active tenants",
			},
			[]string{"plan"},
		),

		// HTTP metrics
		HTTPRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "tenant_manager_http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "path", "status"},
		),
		HTTPRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "tenant_manager_http_request_duration_seconds",
				Help:    "Duration of HTTP requests",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "path"},
		),

		// gRPC metrics
		GRPCRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "tenant_manager_grpc_requests_total",
				Help: "Total number of gRPC requests",
			},
			[]string{"method", "status"},
		),
		GRPCRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "tenant_manager_grpc_request_duration_seconds",
				Help:    "Duration of gRPC requests",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method"},
		),

		// Database metrics
		DBConnectionsActive: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "tenant_manager_db_connections_active",
				Help: "Number of active database connections",
			},
		),
		DBQueryDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "tenant_manager_db_query_duration_seconds",
				Help:    "Duration of database queries",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"query_type"}, // select, insert, update, delete
		),
		DBErrorsTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "tenant_manager_db_errors_total",
				Help: "Total number of database errors",
			},
		),

		// Event metrics
		EventsPublishedTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "tenant_manager_events_published_total",
				Help: "Total number of events published to Kafka",
			},
			[]string{"event_type"},
		),
		EventPublishingErrors: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "tenant_manager_event_publishing_errors_total",
				Help: "Total number of event publishing errors",
			},
		),
	}
}

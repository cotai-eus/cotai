package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/cotai/tenant-manager/internal/app"
	"github.com/cotai/tenant-manager/internal/delivery/grpc"
	"github.com/cotai/tenant-manager/internal/delivery/http"
	"github.com/cotai/tenant-manager/internal/delivery/http/handler"
	"github.com/cotai/tenant-manager/internal/delivery/http/middleware"
	"github.com/cotai/tenant-manager/internal/domain"
	"github.com/cotai/tenant-manager/internal/infrastructure/database"
	"github.com/cotai/tenant-manager/internal/infrastructure/messaging"
	"github.com/cotai/tenant-manager/internal/infrastructure/observability"
	"github.com/cotai/tenant-manager/internal/infrastructure/provisioning"
	"github.com/cotai/tenant-manager/internal/pkg/jwt"
	"github.com/cotai/tenant-manager/internal/usecase"
	"go.uber.org/zap"
)

const version = "0.1.0"

func main() {
	// Load configuration
	cfg, err := app.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	logger, err := observability.NewLogger(cfg.Server.Env, cfg.Server.LogLevel)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("Starting Tenant Manager Service",
		zap.String("version", version),
		zap.String("env", cfg.Server.Env),
		zap.Int("port", cfg.Server.Port),
		zap.Int("grpc_port", cfg.Server.GRPCPort),
	)

	// ==========================
	// Initialize Observability
	// ==========================

	// Initialize metrics
	metrics := observability.NewMetrics()
	logger.Info("Prometheus metrics initialized")

	// Initialize tracer (optional, only if Jaeger is configured)
	if cfg.Observability.JaegerAgentHost != "" {
		tracer, closer, err := observability.InitTracer(
			cfg.Observability.JaegerServiceName,
			cfg.Observability.JaegerAgentHost,
			cfg.Observability.JaegerAgentPort,
			logger,
		)
		if err != nil {
			logger.Warn("Failed to initialize Jaeger tracer", zap.Error(err))
		} else {
			defer closer.Close()
			_ = tracer // tracer is set globally via opentracing.SetGlobalTracer
		}
	}

	// ==========================
	// Initialize Infrastructure
	// ==========================

	// Database connection
	dbConfig := database.Config{
		Host:            cfg.Database.Host,
		Port:            cfg.Database.Port,
		Database:        cfg.Database.Name,
		User:            cfg.Database.User,
		Password:        cfg.Database.Password,
		SSLMode:         cfg.Database.SSLMode,
		MaxConns:        cfg.Database.MaxConns,
		MaxIdleConns:    cfg.Database.MaxIdleConns,
		ConnMaxLifetime: cfg.Database.ConnMaxLifetime,
	}

	db, err := database.NewPostgresDB(dbConfig, logger)
	if err != nil {
		logger.Fatal("Failed to connect to database",
			zap.Error(err),
		)
	}
	defer db.Close()

	logger.Info("Connected to PostgreSQL",
		zap.String("host", cfg.Database.Host),
		zap.String("database", cfg.Database.Name),
	)

	// ==========================
	// Initialize Repositories
	// ==========================

	tenantRepo := database.NewTenantRepository(db.DB(), logger)

	// ==========================
	// Initialize Provisioners
	// ==========================

	schemaProvisioner := provisioning.NewSchemaProvisioner(db.DB(), cfg.Database.MigrationsPath, logger)
	// rls manager can be used later for manual RLS management
	// rlsManager := provisioning.NewRLSManager(db, logger)

	// ==========================
	// Initialize Event Publishers
	// ==========================

	var eventPublisher usecase.EventPublisher
	if len(cfg.Kafka.Brokers) > 0 && cfg.Kafka.Brokers[0] != "" {
		kafkaProducer, err := messaging.NewKafkaProducer(
			cfg.Kafka.Brokers,
			cfg.Kafka.TopicTenantLifecycle,
			logger,
		)
		if err != nil {
			logger.Warn("Failed to initialize Kafka producer, using no-op publisher",
				zap.Error(err),
			)
			eventPublisher = &noopEventPublisher{logger: logger}
		} else {
			eventPublisher = kafkaProducer
			defer kafkaProducer.Close()
			logger.Info("Kafka producer initialized",
				zap.Strings("brokers", cfg.Kafka.Brokers),
				zap.String("topic", cfg.Kafka.TopicTenantLifecycle),
			)
		}
	} else {
		logger.Warn("Kafka not configured, using no-op event publisher")
		eventPublisher = &noopEventPublisher{logger: logger}
	}

	// ==========================
	// Initialize Use Cases
	// ==========================

	createTenantUC := usecase.NewCreateTenantUseCase(tenantRepo, schemaProvisioner, eventPublisher, logger)
	getTenantUC := usecase.NewGetTenantUseCase(tenantRepo, logger)
	listTenantsUC := usecase.NewListTenantsUseCase(tenantRepo, logger)
	updateTenantUC := usecase.NewUpdateTenantUseCase(tenantRepo, eventPublisher, logger)
	suspendTenantUC := usecase.NewSuspendTenantUseCase(tenantRepo, eventPublisher, logger)
	activateTenantUC := usecase.NewActivateTenantUseCase(tenantRepo, eventPublisher, logger)
	deleteTenantUC := usecase.NewDeleteTenantUseCase(tenantRepo, eventPublisher, logger)

	// ==========================
	// Initialize HTTP Components
	// ==========================

	// JWT validator
	jwtValidator := jwt.NewValidator(cfg.JWT.PublicKeyURL, cfg.JWT.Issuer)

	// Middleware
	authMiddleware := middleware.NewAuthMiddleware(jwtValidator, logger)
	loggingMiddleware := middleware.NewLoggingMiddleware(logger)
	recoveryMiddleware := middleware.NewRecoveryMiddleware(logger)
	corsMiddleware := middleware.NewCORSMiddleware()
	metricsMiddleware := middleware.NewMetricsMiddleware(metrics)

	// Handlers
	tenantHandler := handler.NewTenantHandler(
		createTenantUC,
		getTenantUC,
		listTenantsUC,
		updateTenantUC,
		suspendTenantUC,
		activateTenantUC,
		deleteTenantUC,
		logger,
	)
	healthHandler := handler.NewHealthHandler(db, logger)

	// Router
	routerConfig := http.RouterConfig{
		TenantHandler:      tenantHandler,
		HealthHandler:      healthHandler,
		AuthMiddleware:     authMiddleware,
		LoggingMiddleware:  loggingMiddleware,
		RecoveryMiddleware: recoveryMiddleware,
		CORSMiddleware:     corsMiddleware,
		MetricsMiddleware:  metricsMiddleware,
		Logger:             logger,
	}
	router := http.NewRouter(routerConfig)

	// HTTP Server
	httpServer := http.NewServer(cfg.Server.Port, router, logger)

	// ==========================
	// Initialize gRPC Components
	// ==========================

	// gRPC service
	tenantGRPCService := grpc.NewTenantServiceServer(getTenantUC, listTenantsUC, logger)

	// gRPC Server
	grpcServer := grpc.NewServer(cfg.Server.GRPCPort, tenantGRPCService, logger)

	// ==========================
	// Start Both Servers
	// ==========================

	// Create error channel for server failures
	serverErrors := make(chan error, 2)
	var wg sync.WaitGroup

	// Start HTTP server
	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.Info("Starting HTTP server",
			zap.String("addr", fmt.Sprintf(":%d", cfg.Server.Port)),
		)
		if err := httpServer.Start(); err != nil {
			serverErrors <- fmt.Errorf("HTTP server error: %w", err)
		}
	}()

	// Start gRPC server
	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.Info("Starting gRPC server",
			zap.String("addr", fmt.Sprintf(":%d", cfg.Server.GRPCPort)),
		)
		if err := grpcServer.Start(); err != nil {
			serverErrors <- fmt.Errorf("gRPC server error: %w", err)
		}
	}()

	// Wait for shutdown signal or server error
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		logger.Error("Server error occurred", zap.Error(err))
	case sig := <-stop:
		logger.Info("Shutdown signal received", zap.String("signal", sig.String()))
	}

	// Shutdown both servers
	logger.Info("Shutting down servers...")
	httpServer.Shutdown()
	grpcServer.Shutdown()

	// Wait for both servers to stop
	wg.Wait()

	logger.Info("Tenant Manager Service stopped")
}

// noopEventPublisher is a temporary no-op implementation
// Will be replaced with Kafka implementation in Phase G
type noopEventPublisher struct {
	logger *zap.Logger
}

func (p *noopEventPublisher) PublishTenantCreated(ctx context.Context, tenant *domain.Tenant) error {
	p.logger.Debug("Event publishing not implemented yet (noop)",
		zap.String("tenant_id", tenant.TenantID.String()),
	)
	return nil
}

func (p *noopEventPublisher) PublishTenantUpdated(ctx context.Context, tenant *domain.Tenant) error {
	p.logger.Debug("Event publishing not implemented yet (noop)",
		zap.String("tenant_id", tenant.TenantID.String()),
	)
	return nil
}

func (p *noopEventPublisher) PublishTenantSuspended(ctx context.Context, tenant *domain.Tenant) error {
	p.logger.Debug("Event publishing not implemented yet (noop)",
		zap.String("tenant_id", tenant.TenantID.String()),
	)
	return nil
}

func (p *noopEventPublisher) PublishTenantActivated(ctx context.Context, tenant *domain.Tenant) error {
	p.logger.Debug("Event publishing not implemented yet (noop)",
		zap.String("tenant_id", tenant.TenantID.String()),
	)
	return nil
}

func (p *noopEventPublisher) PublishTenantDeleted(ctx context.Context, tenant *domain.Tenant) error {
	p.logger.Debug("Event publishing not implemented yet (noop)",
		zap.String("tenant_id", tenant.TenantID.String()),
	)
	return nil
}

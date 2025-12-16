package main

import (
	"context"
	"fmt"
	"os"

	"github.com/cotai/tenant-manager/internal/app"
	"github.com/cotai/tenant-manager/internal/delivery/http"
	"github.com/cotai/tenant-manager/internal/delivery/http/handler"
	"github.com/cotai/tenant-manager/internal/delivery/http/middleware"
	"github.com/cotai/tenant-manager/internal/domain"
	"github.com/cotai/tenant-manager/internal/infrastructure/database"
	"github.com/cotai/tenant-manager/internal/infrastructure/observability"
	"github.com/cotai/tenant-manager/internal/infrastructure/provisioning"
	"github.com/cotai/tenant-manager/internal/pkg/jwt"
	"github.com/cotai/tenant-manager/internal/usecase"
	"github.com/google/uuid"
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

	// TODO: Implement Kafka producer in Phase G
	// For now, use a no-op publisher
	eventPublisher := &noopEventPublisher{logger: logger}

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
		Logger:             logger,
	}
	router := http.NewRouter(routerConfig)

	// HTTP Server
	httpServer := http.NewServer(cfg.Server.Port, router, logger)

	// ==========================
	// Start HTTP Server
	// ==========================

	logger.Info("Starting HTTP server",
		zap.String("addr", fmt.Sprintf(":%d", cfg.Server.Port)),
	)

	if err := httpServer.Start(); err != nil {
		logger.Fatal("HTTP server failed",
			zap.Error(err),
		)
	}

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

func (p *noopEventPublisher) PublishTenantDeleted(ctx context.Context, tenantID uuid.UUID) error {
	p.logger.Debug("Event publishing not implemented yet (noop)",
		zap.String("tenant_id", tenantID.String()),
	)
	return nil
}

package grpc

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cotai/tenant-manager/internal/delivery/grpc/interceptor"
	tenantv1 "github.com/cotai/tenant-manager/proto/tenant/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

// Server represents the gRPC server
type Server struct {
	port          int
	grpcServer    *grpc.Server
	logger        *zap.Logger
	tenantService *TenantServiceServer
}

// NewServer creates a new gRPC server
func NewServer(port int, tenantService *TenantServiceServer, logger *zap.Logger) *Server {
	// Create gRPC server with interceptors
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptor.LoggingInterceptor(logger),
			interceptor.AuthInterceptor(logger),
		),
	)

	return &Server{
		port:          port,
		grpcServer:    grpcServer,
		logger:        logger,
		tenantService: tenantService,
	}
}

// Start starts the gRPC server with graceful shutdown
func (s *Server) Start() error {
	// Create TCP listener
	addr := fmt.Sprintf(":%d", s.port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	// Register services
	tenantv1.RegisterTenantServiceServer(s.grpcServer, s.tenantService)

	// Register health check service
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(s.grpcServer, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	// Register reflection service (for grpcurl, etc.)
	reflection.Register(s.grpcServer)

	// Channel to listen for interrupt signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Channel to listen for server errors
	serverErrors := make(chan error, 1)

	// Start server in goroutine
	go func() {
		s.logger.Info("gRPC server starting",
			zap.String("addr", addr),
		)

		if err := s.grpcServer.Serve(lis); err != nil {
			serverErrors <- err
		}
	}()

	// Block until signal received or error occurs
	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)
	case sig := <-stop:
		s.logger.Info("Shutdown signal received",
			zap.String("signal", sig.String()),
		)
	}

	// Graceful shutdown
	return s.Shutdown()
}

// Shutdown gracefully shuts down the gRPC server
func (s *Server) Shutdown() error {
	s.logger.Info("gRPC server shutting down...")

	// Create a channel to signal when graceful stop is done
	stopped := make(chan struct{})

	// Graceful stop in goroutine
	go func() {
		s.grpcServer.GracefulStop()
		close(stopped)
	}()

	// Wait for graceful stop or timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	select {
	case <-stopped:
		s.logger.Info("gRPC server stopped gracefully")
		return nil
	case <-ctx.Done():
		s.logger.Warn("gRPC server graceful shutdown timeout, forcing stop")
		s.grpcServer.Stop()
		return fmt.Errorf("graceful shutdown timeout")
	}
}

// Stop immediately stops the gRPC server
func (s *Server) Stop() {
	s.grpcServer.Stop()
}

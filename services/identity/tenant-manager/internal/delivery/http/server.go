package http

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
)

// Server represents the HTTP server
type Server struct {
	httpServer *http.Server
	logger     *zap.Logger
}

// NewServer creates a new HTTP server
func NewServer(port int, handler http.Handler, logger *zap.Logger) *Server {
	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &Server{
		httpServer: httpServer,
		logger:     logger,
	}
}

// Start starts the HTTP server with graceful shutdown
func (s *Server) Start() error {
	// Channel to listen for interrupt signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Channel to listen for server errors
	serverErrors := make(chan error, 1)

	// Start server in goroutine
	go func() {
		s.logger.Info("HTTP server starting",
			zap.String("addr", s.httpServer.Addr),
		)

		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
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

// Shutdown gracefully shuts down the HTTP server
func (s *Server) Shutdown() error {
	s.logger.Info("HTTP server shutting down...")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.logger.Error("HTTP server shutdown error",
			zap.Error(err),
		)
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	s.logger.Info("HTTP server stopped gracefully")
	return nil
}

// Close immediately closes the HTTP server
func (s *Server) Close() error {
	return s.httpServer.Close()
}

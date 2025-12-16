package interceptor

import (
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// AuthInterceptor creates a unary server interceptor for authentication
// For Phase F, this is a basic implementation
// TODO: Implement full JWT validation with Keycloak in future phases
func AuthInterceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// For internal gRPC communication, we can skip strict auth for now
		// In production, you would:
		// 1. Extract metadata from context
		// 2. Validate JWT token
		// 3. Check required scopes/roles

		md, ok := metadata.FromIncomingContext(ctx)
		if ok {
			// Log metadata for debugging
			logger.Debug("gRPC request metadata",
				zap.String("method", info.FullMethod),
				zap.Any("metadata", md),
			)
		}

		// For now, allow all requests (internal service-to-service)
		// TODO: Implement proper authentication
		return handler(ctx, req)
	}
}

// RequireAuth is a stricter version that requires authorization header
// Use this for sensitive operations
func RequireAuth(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		// Check for authorization header
		authHeaders := md.Get("authorization")
		if len(authHeaders) == 0 {
			return nil, status.Error(codes.Unauthenticated, "missing authorization header")
		}

		// TODO: Validate JWT token
		// For now, just log and allow
		logger.Debug("gRPC auth check",
			zap.String("method", info.FullMethod),
			zap.Bool("has_auth", len(authHeaders) > 0),
		)

		return handler(ctx, req)
	}
}

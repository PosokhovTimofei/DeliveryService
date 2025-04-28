package middleware

import (
	"context"
	"strings"

	"github.com/maksroxx/DeliveryService/auth/metrics"
	"github.com/maksroxx/DeliveryService/auth/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type contextKey string

const userIDContextKey = contextKey("userID")

type AuthInterceptor struct {
	authService   *service.AuthService
	publicMethods map[string]bool
}

func NewAuthInterceptor(authService *service.AuthService) *AuthInterceptor {
	return &AuthInterceptor{
		authService: authService,
		publicMethods: map[string]bool{
			"/auth.AuthService/Register": true,
			"/auth.AuthService/Login":    true,
		},
	}
}

func (a *AuthInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {

		if a.publicMethods[info.FullMethod] {
			return handler(ctx, req)
		}

		newCtx, err := a.authorize(ctx, info.FullMethod)
		if err != nil {
			return nil, err
		}

		return handler(newCtx, req)
	}
}

func (a *AuthInterceptor) authorize(ctx context.Context, method string) (context.Context, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		metrics.ValidateFailureTotal.WithLabelValues(method, "no_metadata").Inc()
		return nil, status.Error(codes.Unauthenticated, "metadata is not provided")
	}

	authHeader := md.Get("authorization")
	if len(authHeader) == 0 {
		metrics.ValidateFailureTotal.WithLabelValues(method, "no_token").Inc()
		return nil, status.Error(codes.Unauthenticated, "authorization token is not provided")
	}

	token := strings.TrimPrefix(authHeader[0], "Bearer ")
	claims, err := a.authService.ValidateToken(token)
	if err != nil {
		metrics.ValidateFailureTotal.WithLabelValues(method, "invalid_token").Inc()
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}

	metrics.ValidateSuccessTotal.WithLabelValues(method).Inc()

	newCtx := context.WithValue(ctx, userIDContextKey, claims.UserID)
	return newCtx, nil
}

func UserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userIDContextKey).(string)
	return userID, ok
}

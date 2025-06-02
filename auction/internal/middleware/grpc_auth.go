package middleware

import (
	"context"
	"errors"

	"github.com/maksroxx/DeliveryService/auction/configs"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type contextKey string

const userIDKey contextKey = "user_id"

func GRPCUserIDKey() contextKey {
	return userIDKey
}

func AuthInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		logrus.Infof("Intercepting unary: %s", info.FullMethod)

		if configs.IsExcludedMethod(info.FullMethod) {
			return handler(ctx, req)
		}

		userID, err := extractUserIDFromMetadata(ctx)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}
		ctx = context.WithValue(ctx, userIDKey, userID)
		return handler(ctx, req)
	}
}

func StreamAuthInterceptor() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		userID, err := extractUserIDFromMetadata(ss.Context())
		if err != nil {
			return status.Error(codes.Unauthenticated, err.Error())
		}

		wrapped := &wrappedStream{
			ServerStream: ss,
			ctx:          context.WithValue(ss.Context(), userIDKey, userID),
		}
		return handler(srv, wrapped)
	}
}

func extractUserIDFromMetadata(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", errors.New("missing metadata")
	}
	authHeaders := md.Get("authorization")
	if len(authHeaders) == 0 || authHeaders[0] == "" {
		return "", errors.New("missing authorization header")
	}
	return authHeaders[0], nil
}

type wrappedStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedStream) Context() context.Context {
	return w.ctx
}

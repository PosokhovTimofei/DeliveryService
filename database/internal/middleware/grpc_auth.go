package middleware

import (
	"context"
	"errors"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type contextKey string

const userIDKey contextKey = "user_id"

var excludedMethods = map[string]bool{
	"/delivery.PackageService/TransferExpiredPackages": true,
}

func GRPCUserIDKey() contextKey {
	return userIDKey
}

func GRPCAuthInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		logrus.Printf("Called RPC method: %s", info.FullMethod)
		if excludedMethods[info.FullMethod] {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, errors.New("missing metadata")
		}

		ids := md.Get("authorization")
		if len(ids) == 0 || ids[0] == "" {
			return nil, errors.New("unauthorized: authorization required")
		}

		userID := ids[0]
		newCtx := context.WithValue(ctx, userIDKey, userID)
		return handler(newCtx, req)
	}
}

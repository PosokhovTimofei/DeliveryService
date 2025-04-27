package middleware

import (
	"context"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

func NewLoggingInterceptor(logger *logrus.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp any, err error) {
		resp, err = handler(ctx, req)
		st, _ := status.FromError(err)
		entry := logger.WithFields(logrus.Fields{
			"method": info.FullMethod,
			"error":  st.Message(),
			"code":   st.Code(),
		})
		if err != nil {
			entry.Error("gRPC call failed")
		} else {
			entry.Info("gRPC call success")
		}
		return resp, err
	}
}

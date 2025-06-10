package handler

import (
	"context"

	"github.com/maksroxx/DeliveryService/auth/middleware"
	"github.com/maksroxx/DeliveryService/auth/models"
	"github.com/maksroxx/DeliveryService/auth/service"
	authpb "github.com/maksroxx/DeliveryService/proto/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthGRPCServer struct {
	authpb.UnimplementedAuthServiceServer
	service *service.AuthService
}

func NewAuthGRPCServer(service *service.AuthService) *AuthGRPCServer {
	return &AuthGRPCServer{service: service}
}

func (s *AuthGRPCServer) Register(ctx context.Context, req *authpb.RegisterRequest) (*authpb.AuthResponse, error) {
	user, token, err := s.service.Register(ctx, req.Email, req.Password)
	if err != nil {
		return nil, grpcError(err)
	}
	return &authpb.AuthResponse{
		UserId: user.ID,
		Token:  token,
		Role:   user.Role,
	}, nil
}

func (s *AuthGRPCServer) RegisterModerator(ctx context.Context, req *authpb.RegisterRequest) (*authpb.AuthResponse, error) {
	user, token, err := s.service.RegisterModerator(ctx, req.Email, req.Password)
	if err != nil {
		return nil, grpcError(err)
	}
	return &authpb.AuthResponse{
		UserId: user.ID,
		Token:  token,
		Role:   user.Role,
	}, nil
}

func (s *AuthGRPCServer) Login(ctx context.Context, req *authpb.LoginRequest) (*authpb.AuthResponse, error) {
	user, token, err := s.service.Login(ctx, req.Email, req.Password)
	if err != nil {
		return nil, grpcError(err)
	}
	return &authpb.AuthResponse{
		UserId: user.ID,
		Token:  token,
		Role:   user.Role,
	}, nil
}

func (s *AuthGRPCServer) Validate(ctx context.Context, req *authpb.ValidateRequest) (*authpb.ValidateResponse, error) {
	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok || userID == "" {
		return &authpb.ValidateResponse{
			Valid: "not",
		}, nil
	}

	return &authpb.ValidateResponse{
		Valid:  "ok",
		UserId: userID,
	}, nil
}

func (s *AuthGRPCServer) GetUserByTelegramCode(ctx context.Context, req *authpb.TelegramCodeLookupRequest) (*authpb.TelegramCodeLookupResponse, error) {
	userId, err := s.service.GetUserIDByTelegramCode(req.Code)
	if err != nil {
		return nil, grpcError(err)
	}
	return &authpb.TelegramCodeLookupResponse{UserId: userId}, nil
}

func (s *AuthGRPCServer) GenerateTelegramCode(ctx context.Context, req *authpb.TelegramCodeRequest) (*authpb.TelegramCodeResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}
	code, err := s.service.GenerateTelegramCode(req.UserId)
	if err != nil {
		return nil, grpcError(err)
	}
	return &authpb.TelegramCodeResponse{Code: code}, nil
}

func grpcError(err error) error {
	switch err {
	case models.ErrEmailAlreadyExists:
		return status.Error(codes.AlreadyExists, err.Error())
	case models.ErrInvalidCredentials:
		return status.Error(codes.Unauthenticated, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}

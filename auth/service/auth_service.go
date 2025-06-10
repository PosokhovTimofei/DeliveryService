package service

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/maksroxx/DeliveryService/auth/models"
	"github.com/maksroxx/DeliveryService/auth/repository"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	repo      repository.UserRepository
	tgRepo    repository.Telegramer
	jwtSecret string
}

func NewAuthService(userRepo repository.UserRepository, tgRepo repository.Telegramer, jwtSecret string) *AuthService {
	return &AuthService{
		repo:      userRepo,
		tgRepo:    tgRepo,
		jwtSecret: jwtSecret,
	}
}

func (s *AuthService) Register(ctx context.Context, email, password string) (*models.User, string, error) {
	existing, _ := s.repo.GetByEmail(ctx, email)
	if existing != nil {
		return nil, "", models.ErrEmailAlreadyExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", err
	}

	user := &models.User{
		Email:             email,
		EncryptedPassword: string(hashedPassword),
	}
	user.GenerateUserID()
	user.Role = "user"
	if err = s.repo.CreateUser(ctx, user); err != nil {
		return nil, "", err
	}

	token, err := s.GenerateToken(user)
	return user, token, err
}

func (s *AuthService) RegisterModerator(ctx context.Context, email, password string) (*models.User, string, error) {
	existing, _ := s.repo.GetByEmail(ctx, email)
	if existing != nil {
		return nil, "", models.ErrEmailAlreadyExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", err
	}

	user := &models.User{
		Email:             email,
		EncryptedPassword: string(hashedPassword),
	}
	user.GenerateUserID()
	user.Role = "moderator"
	if err = s.repo.CreateUser(ctx, user); err != nil {
		return nil, "", err
	}

	token, err := s.GenerateToken(user)
	return user, token, err
}

func (s *AuthService) GenerateToken(user *models.User) (string, error) {
	claims := &models.JWTClaims{
		UserID: user.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*models.User, string, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, "", models.ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(user.EncryptedPassword),
		[]byte(password),
	); err != nil {
		return nil, "", models.ErrInvalidCredentials
	}
	token, err := s.GenerateToken(user)
	return user, token, err
}

func (s *AuthService) ValidateToken(tokenString string) (*models.JWTClaims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&models.JWTClaims{},
		func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, models.ErrInvalidToken
			}
			return []byte(s.jwtSecret), nil
		},
	)

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*models.JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, models.ErrInvalidToken
}

func (s *AuthService) GetUserIDByTelegramCode(code string) (string, error) {
	return s.tgRepo.FindUserIDByCode(code)
}

func (s *AuthService) GenerateTelegramCode(userID string) (string, error) {
	code := "auth_" + uuid.NewString()[:8]
	err := s.tgRepo.Save(code, userID, 10*time.Minute)
	if err != nil {
		return "", err
	}
	return code, nil
}

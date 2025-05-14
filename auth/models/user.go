package models

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type User struct {
	ID                string `bson:"user_id" json:"user_id"`
	Email             string `bson:"email" json:"email"`
	EncryptedPassword string `bson:"encrypted_password" json:"-"`
	Role              string `bson:"role" json:"role"`
}

type JWTClaims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

func (u *User) GenerateUserID() {
	u.ID = uuid.New().String()
}

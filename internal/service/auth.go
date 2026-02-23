package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"cprt-lis/internal/domain"
	"cprt-lis/internal/repository"

	"github.com/golang-jwt/jwt/v5"
)

type AuthService struct {
	users     repository.UserRepository
	jwtSecret []byte
	jwtIssuer string
}

func NewAuthService(users repository.UserRepository, secret, issuer string) *AuthService {
	return &AuthService{
		users:     users,
		jwtSecret: []byte(secret),
		jwtIssuer: issuer,
	}
}

func (s *AuthService) Login(ctx context.Context, username, password string) (string, domain.User, error) {
	user, err := s.users.GetByUsername(ctx, username)
	if err != nil {
		return "", domain.User{}, errors.New("invalid credentials")
	}

	if user.Status != "ACTIVE" {
		return "", domain.User{}, errors.New("user inactive")
	}

	hash := sha256.Sum256([]byte(password))
	passwordHash := hex.EncodeToString(hash[:])

	if user.PasswordHash != passwordHash {
		return "", domain.User{}, errors.New("invalid credentials")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss":  s.jwtIssuer,
		"sub":  user.UserUUID,
		"role": string(user.Role),
		"iat":  time.Now().Unix(),
		"exp":  time.Now().Add(48 * time.Hour).Unix(),
	})

	signed, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", domain.User{}, errors.New("token generation failed")
	}

	return signed, user, nil
}

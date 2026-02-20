package service

import (
	"context"
	"errors"
	"time"

	"cprt-lis/internal/domain"
	"cprt-lis/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	users     repository.UserRepository
	jwtSecret []byte
	jwtIssuer string
	jwtTTL    time.Duration
}

func NewAuthService(users repository.UserRepository, secret, issuer string, ttlMinutes int) *AuthService {
	return &AuthService{
		users:     users,
		jwtSecret: []byte(secret),
		jwtIssuer: issuer,
		jwtTTL:    time.Duration(ttlMinutes) * time.Minute,
	}
}

func (s *AuthService) Login(ctx context.Context, username, password string) (string, domain.User, error) {
	user, err := s.users.GetByUsername(ctx, username)
	if err != nil {
		return "", domain.User{}, errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", domain.User{}, errors.New("invalid credentials")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    s.jwtIssuer,
		Subject:   user.UserUUID,
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.jwtTTL)),
	})

	signed, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", domain.User{}, errors.New("token generation failed")
	}

	return signed, user, nil
}

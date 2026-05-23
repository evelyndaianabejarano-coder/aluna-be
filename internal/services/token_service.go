package services

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/evelyndaianabejarano-coder/aluna-be/config"
	"github.com/evelyndaianabejarano-coder/aluna-be/internal/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type Claims struct {
	UserID uuid.UUID   `json:"user_id"`
	Role   models.Role `json:"role"`
	JTI    string      `json:"jti,omitempty"`
	jwt.RegisteredClaims
}

type TokenService interface {
	GenerateAccessToken(userID uuid.UUID, role models.Role) (string, error)
	GenerateRefreshToken(userID uuid.UUID) (string, error)
	ValidateToken(tokenString string) (*Claims, error)
	ValidateRefreshToken(tokenString string) (*Claims, error)
	RevokeRefreshToken(tokenString string) error
}

type tokenService struct {
	secret          []byte
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
	rdb             *redis.Client
}

func NewTokenService(cfg *config.Config, rdb *redis.Client) (TokenService, error) {
	accessMinutes, err := strconv.Atoi(cfg.JWTExpirationMinutes)
	if err != nil {
		return nil, fmt.Errorf("JWT_EXPIRATION_MINUTES inválido: %w", err)
	}
	refreshDays, err := strconv.Atoi(cfg.RefreshTokenExpirationDays)
	if err != nil {
		return nil, fmt.Errorf("REFRESH_TOKEN_EXPIRATION_DAYS inválido: %w", err)
	}

	return &tokenService{
		secret:          []byte(cfg.JWTSecret),
		accessTokenTTL:  time.Duration(accessMinutes) * time.Minute,
		refreshTokenTTL: time.Duration(refreshDays) * 24 * time.Hour,
		rdb:             rdb,
	}, nil
}

func (s *tokenService) GenerateAccessToken(userID uuid.UUID, role models.Role) (string, error) {
	claims := &Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.accessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

func (s *tokenService) GenerateRefreshToken(userID uuid.UUID) (string, error) {
	jti := uuid.New().String()
	claims := &Claims{
		UserID: userID,
		JTI:    jti,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.refreshTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.secret)
	if err != nil {
		return "", err
	}

	key := fmt.Sprintf("refresh:%s", jti)
	if err := s.rdb.Set(context.Background(), key, userID.String(), s.refreshTokenTTL).Err(); err != nil {
		return "", err
	}
	return signed, nil
}

func (s *tokenService) ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("método de firma inesperado: %v", t.Header["alg"])
		}
		return s.secret, nil
	})
	if err != nil || !token.Valid {
		return nil, err
	}
	return claims, nil
}

func (s *tokenService) ValidateRefreshToken(tokenString string) (*Claims, error) {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}
	key := fmt.Sprintf("refresh:%s", claims.JTI)
	if err := s.rdb.Get(context.Background(), key).Err(); err != nil {
		return nil, fmt.Errorf("refresh token revocado o inexistente")
	}
	return claims, nil
}

func (s *tokenService) RevokeRefreshToken(tokenString string) error {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return err
	}
	key := fmt.Sprintf("refresh:%s", claims.JTI)
	return s.rdb.Del(context.Background(), key).Err()
}

package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenManager handles JWT issuing and refresh token hashing.
type TokenManager struct {
	secret          []byte
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

var errTokenManagerNil = errors.New("token manager is nil")

// Claims describes the custom payload embedded in access tokens.
type Claims struct {
	UserID   int64  `json:"userId"`
	UserUUID string `json:"userUuid"`
	jwt.RegisteredClaims
}

// NewTokenManager creates a TokenManager with validated configuration.
func NewTokenManager(secret string, accessTTL, refreshTTL time.Duration) (*TokenManager, error) {
	if secret == "" {
		return nil, fmt.Errorf("token secret cannot be empty")
	}
	if accessTTL <= 0 {
		return nil, fmt.Errorf("access token ttl must be positive")
	}
	if refreshTTL <= 0 {
		return nil, fmt.Errorf("refresh token ttl must be positive")
	}

	return &TokenManager{
		secret:          []byte(secret),
		accessTokenTTL:  accessTTL,
		refreshTokenTTL: refreshTTL,
	}, nil
}

// IssueAccessToken builds a signed JWT for the provided principal.
func (m *TokenManager) IssueAccessToken(userID int64, userUUID string) (string, time.Time, error) {
	if m == nil {
		return "", time.Time{}, errTokenManagerNil
	}

	expiresAt := time.Now().Add(m.accessTokenTTL)
	claims := Claims{
		UserID:   userID,
		UserUUID: userUUID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			Subject:   fmt.Sprintf("user:%d", userID),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign jwt: %w", err)
	}

	return signed, expiresAt, nil
}

// ParseAccessToken validates the JWT signature and extracts claims.
func (m *TokenManager) ParseAccessToken(token string) (*Claims, error) {
	if m == nil {
		return nil, errTokenManagerNil
	}

	parsed, err := jwt.ParseWithClaims(token, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %T", t.Method)
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("parse jwt: %w", err)
	}

	claims, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid {
		return nil, fmt.Errorf("invalid jwt claims")
	}

	return claims, nil
}

// GenerateRefreshToken returns a random base64 token and its expiry timestamp.
func (m *TokenManager) GenerateRefreshToken() (string, time.Time, error) {
	if m == nil {
		return "", time.Time{}, errTokenManagerNil
	}

	expiresAt := time.Now().Add(m.refreshTokenTTL)
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", time.Time{}, fmt.Errorf("generate refresh token: %w", err)
	}

	token := base64.RawURLEncoding.EncodeToString(buf)
	return token, expiresAt, nil
}

// HashRefreshToken hashes the provided refresh token for persistence.
func HashRefreshToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

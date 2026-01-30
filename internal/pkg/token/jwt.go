package token

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalidToken     = errors.New("invalid token")
	ErrExpiredToken     = errors.New("token has expired")
	ErrInvalidSignature = errors.New("invalid signature")
)

type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	Role   string    `json:"role"`
	jwt.RegisteredClaims
}

type JWTManager interface {
	GenerateAccessToken(userID uuid.UUID, email, role string) (string, error)
	GenerateRefreshToken() (string, string, error)
	ValidateAccessToken(tokenString string) (*Claims, error)
	HashRefreshToken(token string) string
}

type jwtManager struct {
	secretKey          []byte
	accessTokenExpiry  time.Duration
	refreshTokenExpiry time.Duration
	issuer             string
}

func NewJWTManager(secretKey string, accessExpiry, refreshExpiry time.Duration, issuer string) JWTManager {
	return &jwtManager{
		secretKey:          []byte(secretKey),
		accessTokenExpiry:  accessExpiry,
		refreshTokenExpiry: refreshExpiry,
		issuer:             issuer,
	}
}

func (m *jwtManager) GenerateAccessToken(userID uuid.UUID, email, role string) (string, error) {
	now := time.Now()
	claims := &Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(m.accessTokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    m.issuer,
			Subject:   userID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secretKey)
}

func (m *jwtManager) GenerateRefreshToken() (string, string, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", "", err
	}

	token := base64.URLEncoding.EncodeToString(tokenBytes)
	hash := m.HashRefreshToken(token)

	return token, hash, nil
}

func (m *jwtManager) ValidateAccessToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidSignature
		}
		return m.secretKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

func (m *jwtManager) HashRefreshToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func (m *jwtManager) GetAccessTokenExpiry() time.Duration {
	return m.accessTokenExpiry
}

func (m *jwtManager) GetRefreshTokenExpiry() time.Duration {
	return m.refreshTokenExpiry
}

package internal

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// TokenManager defines the interface for token operations.
type TokenManager interface {
	CreateTokenPair(userID int, username string, isAdmin bool) (*TokenPair, error)
	VerifyToken(tokenString string) (*TokenClaims, error)
	RefreshAccessToken(refreshToken string) (string, error)
}

// TokenHandler struct holds the JWT operations
type TokenHandler struct {
	secretKey     []byte
	accessExpiry  time.Duration // Short-lived
	refreshExpiry time.Duration // Long-lived
}

// TokenPair represents a pair of access and refresh tokens
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// TokenClaims represents the claims in a JWT token
type TokenClaims struct {
	jwt.RegisteredClaims
	Username string `json:"username"`
	UserID   int    `json:"user_id"`
	Admin    bool   `json:"admin"`
}

func NewTokenHandler(secretKey string, accessExpiry, refreshExpiry time.Duration) *TokenHandler {
	return &TokenHandler{
		secretKey:     []byte(secretKey),
		accessExpiry:  accessExpiry,
		refreshExpiry: refreshExpiry,
	}
}

// CreateTokenPair generates a new access and refresh token pair
func (h *TokenHandler) CreateTokenPair(userID int, username string, isAdmin bool) (*TokenPair, error) {
	accessToken, err := h.createAccessToken(userID, username, isAdmin)
	if err != nil {
		return nil, err
	}
	refreshToken, err := h.createRefreshToken(userID, username, isAdmin)
	if err != nil {
		return nil, err
	}
	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// VerifyToken verifies the token and returns the claims
func (h *TokenHandler) VerifyToken(tokenString string) (*TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return h.secretKey, nil
	})
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(*TokenClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

// createAccessToken generates a new access token
func (h *TokenHandler) createAccessToken(userID int, username string, isAdmin bool) (string, error) {
	claims := TokenClaims{
		Username: username,
		UserID:   userID,
		Admin:    isAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(h.accessExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(h.secretKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// createRefreshToken generates a new refresh token
func (h *TokenHandler) createRefreshToken(userID int, username string, isAdmin bool) (string, error) {
	claims := TokenClaims{
		Username: username,
		UserID:   userID,
		Admin:    isAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(h.accessExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(h.secretKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// RefreshAccessToken refreshes the access token using a refresh token
func (h *TokenHandler) RefreshAccessToken(refreshToken string) (string, error) {
	claims, err := h.VerifyToken(refreshToken)
	if err != nil {
		return "", err
	}

	accessToken, err := h.createAccessToken(claims.UserID, claims.Username, claims.Admin)
	if err != nil {
		return "", err
	}

	return accessToken, nil
}

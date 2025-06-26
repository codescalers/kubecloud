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
	AccessTokenFromRefresh(refreshToken string) (string, error)
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
	accessToken, err := h.createToken(userID, username, isAdmin, h.accessExpiry)
	if err != nil {
		return nil, err
	}
	refreshToken, err := h.createToken(userID, username, isAdmin, h.refreshExpiry)
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

	if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
		return nil, fmt.Errorf("token has expired")
	}

	return claims, nil
}

// createToken creates token with given expiry time
func (h *TokenHandler) createToken(userID int, username string, isAdmin bool, expiry time.Duration) (string, error) {
	claims := TokenClaims{
		Username: username,
		UserID:   userID,
		Admin:    isAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(h.secretKey)
}

// RefreshAccessToken refreshes the access token using a refresh token
func (h *TokenHandler) AccessTokenFromRefresh(refreshToken string) (string, error) {
	claims, err := h.VerifyToken(refreshToken)
	if err != nil {
		return "", err
	}

	accessToken, err := h.createToken(claims.UserID, claims.Username, claims.Admin, h.refreshExpiry)
	if err != nil {
		return "", err
	}

	return accessToken, nil
}

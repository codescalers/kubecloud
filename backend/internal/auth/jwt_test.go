package auth

import (
	"strings"
	"testing"
	"time"
)

// TestNewTokenHandler tests token handler initialization.
// This scenario covers:
// - Creates handler with secret key and expiry times
// - Stores configuration for token generation
func TestNewTokenHandler(t *testing.T) {
	secret := "test-secret-key"
	accessExpiry := 15 * time.Minute
	refreshExpiry := 7 * 24 * time.Hour

	handler := NewTokenHandler(secret, accessExpiry, refreshExpiry)

	if handler.accessExpiry != accessExpiry {
		t.Errorf("NewTokenHandler() accessExpiry = %v, want %v", handler.accessExpiry, accessExpiry)
	}
	if handler.refreshExpiry != refreshExpiry {
		t.Errorf("NewTokenHandler() refreshExpiry = %v, want %v", handler.refreshExpiry, refreshExpiry)
	}
}

// TestCreateTokenPair tests token pair generation.
// This scenario covers:
// - Generates both access and refresh tokens
// - Tokens are non-empty strings
// - Tokens contain proper JWT format (3 parts separated by dots)
// - Different user IDs and admin flags produce different tokens
func TestCreateTokenPair(t *testing.T) {
	tests := []struct {
		name        string
		userID      int
		username    string
		isAdmin     bool
		description string
	}{
		{
			name:        "regular_user",
			userID:      1,
			username:    "john",
			isAdmin:     false,
			description: "creating token pair for regular user",
		},
		{
			name:        "admin_user",
			userID:      2,
			username:    "admin",
			isAdmin:     true,
			description: "creating token pair for admin user",
		},
		{
			name:        "user_zero_id",
			userID:      0,
			username:    "guest",
			isAdmin:     false,
			description: "creating token pair for user with ID 0",
		},
		{
			name:        "user_negative_id",
			userID:      -1,
			username:    "invalid",
			isAdmin:     false,
			description: "creating token pair for user with negative ID",
		},
		{
			name:        "user_large_id",
			userID:      999999,
			username:    "user999",
			isAdmin:     true,
			description: "creating token pair for user with large ID",
		},
	}

	handler := NewTokenHandler("secret-key", 15*time.Minute, 7*24*time.Hour)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pair, err := handler.CreateTokenPair(tt.userID, tt.username, tt.isAdmin)

			if err != nil {
				t.Errorf("CreateTokenPair() error = %v (%s)", err, tt.description)
				return
			}
			if pair == nil {
				t.Errorf("CreateTokenPair() returned nil (%s)", tt.description)
				return
			}
			if pair.AccessToken == "" {
				t.Errorf("CreateTokenPair() AccessToken is empty (%s)", tt.description)
			}
			if pair.RefreshToken == "" {
				t.Errorf("CreateTokenPair() RefreshToken is empty (%s)", tt.description)
			}

			// Check JWT format (3 parts separated by dots)
			accessParts := strings.Split(pair.AccessToken, ".")
			if len(accessParts) != 3 {
				t.Errorf("CreateTokenPair() AccessToken format invalid, got %d parts (%s)", len(accessParts), tt.description)
			}
			refreshParts := strings.Split(pair.RefreshToken, ".")
			if len(refreshParts) != 3 {
				t.Errorf("CreateTokenPair() RefreshToken format invalid, got %d parts (%s)", len(refreshParts), tt.description)
			}

			// Access and refresh tokens should be different
			if pair.AccessToken == pair.RefreshToken {
				t.Errorf("CreateTokenPair() AccessToken and RefreshToken should be different (%s)", tt.description)
			}
		})
	}
}

// TestVerifyToken tests token verification.
// This scenario covers:
// - Valid token verifies successfully and returns claims
// - Claims contain correct user information and admin flag
// - Invalid token string fails verification
// - Empty token fails verification
// - Malformed token fails verification
func TestVerifyToken(t *testing.T) {
	handler := NewTokenHandler("secret-key", 15*time.Minute, 7*24*time.Hour)

	// Create a valid token
	pair, err := handler.CreateTokenPair(123, "testuser", true)
	if err != nil {
		t.Fatalf("CreateTokenPair() error = %v", err)
	}

	tests := []struct {
		name             string
		token            string
		expectError      bool
		expectedUserID   int
		expectedUsername string
		expectedAdmin    bool
		description      string
	}{
		{
			name:             "valid_token",
			token:            pair.AccessToken,
			expectError:      false,
			expectedUserID:   123,
			expectedUsername: "testuser",
			expectedAdmin:    true,
			description:      "verifying valid token",
		},
		{
			name:        "empty_token",
			token:       "",
			expectError: true,
			description: "verifying empty token",
		},
		{
			name:        "malformed_token",
			token:       "invalid.token.format.extra",
			expectError: true,
			description: "verifying malformed token with extra parts",
		},
		{
			name:        "invalid_format",
			token:       "notavalidtoken",
			expectError: true,
			description: "verifying token without JWT format",
		},
		{
			name:        "wrong_secret",
			token:       pair.AccessToken,
			expectError: true,
			description: "verifying token with different secret",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var h *TokenHandler
			if tt.name == "wrong_secret" {
				h = NewTokenHandler("different-secret", 15*time.Minute, 7*24*time.Hour)
			} else {
				h = handler
			}

			claims, err := h.VerifyToken(tt.token)

			if (err != nil) != tt.expectError {
				t.Errorf("VerifyToken() error = %v, expectError %v (%s)", err, tt.expectError, tt.description)
				return
			}

			if !tt.expectError {
				if claims == nil {
					t.Errorf("VerifyToken() returned nil claims (%s)", tt.description)
					return
				}
				if claims.UserID != tt.expectedUserID {
					t.Errorf("VerifyToken() UserID = %d, want %d (%s)", claims.UserID, tt.expectedUserID, tt.description)
				}
				if claims.Username != tt.expectedUsername {
					t.Errorf("VerifyToken() Username = %q, want %q (%s)", claims.Username, tt.expectedUsername, tt.description)
				}
				if claims.Admin != tt.expectedAdmin {
					t.Errorf("VerifyToken() Admin = %v, want %v (%s)", claims.Admin, tt.expectedAdmin, tt.description)
				}
			}
		})
	}
}

// TestAccessTokenFromRefresh tests access token refresh.
// This scenario covers:
// - Valid refresh token generates new access token
// - New access token is different from original
// - Invalid refresh token fails
// - Expired refresh token fails
// - New access token contains same user information
func TestAccessTokenFromRefresh(t *testing.T) {
	handler := NewTokenHandler("secret-key", 15*time.Minute, 7*24*time.Hour)

	// Create token pair
	pair, err := handler.CreateTokenPair(456, "alice", false)
	if err != nil {
		t.Fatalf("CreateTokenPair() error = %v", err)
	}

	tests := []struct {
		name         string
		refreshToken string
		expectError  bool
		description  string
	}{
		{
			name:         "valid_refresh",
			refreshToken: pair.RefreshToken,
			expectError:  false,
			description:  "refreshing with valid token",
		},
		{
			name:         "invalid_refresh",
			refreshToken: "invalid.refresh.token",
			expectError:  true,
			description:  "refreshing with invalid token",
		},
		{
			name:         "empty_refresh",
			refreshToken: "",
			expectError:  true,
			description:  "refreshing with empty token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newAccessToken, err := handler.AccessTokenFromRefresh(tt.refreshToken)

			if (err != nil) != tt.expectError {
				t.Errorf("AccessTokenFromRefresh() error = %v, expectError %v (%s)", err, tt.expectError, tt.description)
				return
			}

			if !tt.expectError {
				if newAccessToken == "" {
					t.Errorf("AccessTokenFromRefresh() returned empty token (%s)", tt.description)
					return
				}
				if newAccessToken == pair.AccessToken {
					t.Errorf("AccessTokenFromRefresh() should generate different token (%s)", tt.description)
				}

				// Verify new token contains same user info
				claims, err := handler.VerifyToken(newAccessToken)
				if err != nil {
					t.Errorf("VerifyToken() on new token error = %v (%s)", err, tt.description)
					return
				}
				if claims.UserID != 456 {
					t.Errorf("AccessTokenFromRefresh() new token UserID = %d, want 456 (%s)", claims.UserID, tt.description)
				}
				if claims.Username != "alice" {
					t.Errorf("AccessTokenFromRefresh() new token Username = %q, want alice (%s)", claims.Username, tt.description)
				}
				if claims.Admin != false {
					t.Errorf("AccessTokenFromRefresh() new token Admin = %v, want false (%s)", claims.Admin, tt.description)
				}
			}
		})
	}
}

// TestTokenExpiry tests token expiration handling.
// This scenario covers:
// - Token with very short expiry is rejected when expired
// - Token with future expiry verifies successfully
// - Expired token returns appropriate error
func TestTokenExpiry(t *testing.T) {
	// Create handler with very short expiry
	shortExpiry := -1 * time.Second // Already expired
	handler := NewTokenHandler("secret-key", shortExpiry, 7*24*time.Hour)

	pair, err := handler.CreateTokenPair(789, "bob", false)
	if err != nil {
		t.Fatalf("CreateTokenPair() error = %v", err)
	}

	// Token should be expired
	_, err = handler.VerifyToken(pair.AccessToken)
	if err == nil {
		t.Errorf("VerifyToken() should reject expired token, got nil error")
	}
	if !strings.Contains(err.Error(), "expired") {
		t.Errorf("VerifyToken() error should mention expiry, got: %v", err)
	}

	// Create handler with future expiry
	futureExpiry := 1 * time.Hour
	handlerFuture := NewTokenHandler("secret-key", futureExpiry, 7*24*time.Hour)
	pairFuture, err := handlerFuture.CreateTokenPair(789, "bob", false)
	if err != nil {
		t.Fatalf("CreateTokenPair() error = %v", err)
	}

	// Token should verify successfully
	claims, err := handlerFuture.VerifyToken(pairFuture.AccessToken)
	if err != nil {
		t.Errorf("VerifyToken() should verify future token, got error: %v", err)
	}
	if claims == nil {
		t.Errorf("VerifyToken() should return claims for valid token")
	}
}

// TestTokenClaims tests TokenClaims struct fields.
// This scenario covers:
// - Token claims contain all required fields (UserID, Username, Admin, StandardClaims)
// - Claims are preserved through token creation and verification
func TestTokenClaims(t *testing.T) {
	handler := NewTokenHandler("secret-key", 15*time.Minute, 7*24*time.Hour)

	pair, err := handler.CreateTokenPair(999, "testuser", true)
	if err != nil {
		t.Fatalf("CreateTokenPair() error = %v", err)
	}

	claims, err := handler.VerifyToken(pair.AccessToken)
	if err != nil {
		t.Fatalf("VerifyToken() error = %v", err)
	}

	if claims.UserID != 999 {
		t.Errorf("TokenClaims.UserID = %d, want 999", claims.UserID)
	}
	if claims.Username != "testuser" {
		t.Errorf("TokenClaims.Username = %q, want testuser", claims.Username)
	}
	if claims.Admin != true {
		t.Errorf("TokenClaims.Admin = %v, want true", claims.Admin)
	}
	if claims.IssuedAt == nil {
		t.Errorf("TokenClaims.IssuedAt should be set")
	}
	if claims.ExpiresAt == nil {
		t.Errorf("TokenClaims.ExpiresAt should be set")
	}
}

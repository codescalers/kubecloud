package middlewares

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"kubecloud/internal/auth"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

// MockTokenManager implements auth.TokenManager for testing
type MockTokenManager struct {
	validToken   string
	validClaims  *auth.TokenClaims
	shouldFail   bool
	errorMessage string
}

func (m *MockTokenManager) CreateTokenPair(userID int, username string, isAdmin bool) (*auth.TokenPair, error) {
	return nil, nil
}

func (m *MockTokenManager) VerifyToken(tokenString string) (*auth.TokenClaims, error) {
	if m.shouldFail {
		return nil, errors.New("invalid token")
	}
	if tokenString == m.validToken {
		return m.validClaims, nil
	}
	return nil, fmt.Errorf("invalid token")
}

func (m *MockTokenManager) AccessTokenFromRefresh(refreshToken string) (string, error) {
	return "", nil
}

// TestUserMiddlewareValidToken tests middleware with valid Authorization header
func TestUserMiddlewareValidToken(t *testing.T) {
	tokenManager := &MockTokenManager{
		validToken: "valid-token",
		validClaims: &auth.TokenClaims{
			Username: "testuser",
			UserID:   123,
			Admin:    false,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			},
		},
		shouldFail: false,
	}

	// Create test context
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/events", nil)
	c.Request.Header.Set("Authorization", "Bearer valid-token")

	// Create middleware
	middleware := UserMiddleware(tokenManager)
	middleware(c)

	// Check if user_id was set in context
	userID, exists := c.Get("user_id")
	if !exists {
		t.Errorf("UserMiddleware should set user_id in context")
	}
	if userID != 123 {
		t.Errorf("UserMiddleware user_id = %v, want 123", userID)
	}

	// Check if admin was set in context
	admin, exists := c.Get("admin")
	if !exists {
		t.Errorf("UserMiddleware should set admin in context")
	}
	if admin != false {
		t.Errorf("UserMiddleware admin = %v, want false", admin)
	}
}

// TestUserMiddlewareMissingAuthHeader tests middleware with missing Authorization header
func TestUserMiddlewareMissingAuthHeader(t *testing.T) {
	tokenManager := &MockTokenManager{}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/events", nil)
	// No Authorization header

	middleware := UserMiddleware(tokenManager)
	middleware(c)

	// Should abort with 401
	if w.Code != http.StatusUnauthorized {
		t.Errorf("UserMiddleware status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

// TestUserMiddlewareInvalidBearerFormat tests middleware with invalid Bearer format
func TestUserMiddlewareInvalidBearerFormat(t *testing.T) {
	tokenManager := &MockTokenManager{}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/events", nil)
	c.Request.Header.Set("Authorization", "InvalidFormat token")

	middleware := UserMiddleware(tokenManager)
	middleware(c)

	// Should abort with 401
	if w.Code != http.StatusUnauthorized {
		t.Errorf("UserMiddleware status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

// TestUserMiddlewareInvalidToken tests middleware with invalid token
func TestUserMiddlewareInvalidToken(t *testing.T) {
	tokenManager := &MockTokenManager{
		validToken: "valid-token",
		validClaims: &auth.TokenClaims{
			UserID: 123,
		},
		shouldFail: true,
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/events", nil)
	c.Request.Header.Set("Authorization", "Bearer invalid-token")

	middleware := UserMiddleware(tokenManager)
	middleware(c)

	// Should abort with 401
	if w.Code != http.StatusUnauthorized {
		t.Errorf("UserMiddleware status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

// TestUserMiddlewareNoQueryParameter tests that query parameter tokens are rejected
func TestUserMiddlewareNoQueryParameter(t *testing.T) {
	tokenManager := &MockTokenManager{
		validToken: "valid-token",
		validClaims: &auth.TokenClaims{
			UserID: 123,
		},
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/events?token=valid-token", nil)
	// No Authorization header - only query parameter

	middleware := UserMiddleware(tokenManager)
	middleware(c)

	// Should abort with 401 (query parameter should be ignored)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("UserMiddleware should reject query parameter token, status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

// TestUserMiddlewareAdminUser tests middleware with admin user
func TestUserMiddlewareAdminUser(t *testing.T) {
	tokenManager := &MockTokenManager{
		validToken: "admin-token",
		validClaims: &auth.TokenClaims{
			Username: "admin",
			UserID:   999,
			Admin:    true,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			},
		},
		shouldFail: false,
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/events", nil)
	c.Request.Header.Set("Authorization", "Bearer admin-token")

	middleware := UserMiddleware(tokenManager)
	middleware(c)

	// Check if admin flag is correctly set
	admin, exists := c.Get("admin")
	if !exists {
		t.Errorf("UserMiddleware should set admin in context")
	}
	if admin != true {
		t.Errorf("UserMiddleware admin = %v, want true", admin)
	}
}

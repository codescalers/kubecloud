package kyc

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/vedhavyas/go-subkey/sr25519"
)

// mockHTTPClient is a mock implementation for testing
type mockHTTPClient struct {
	doFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.doFunc(req)
}

// TestIsUserVerified tests the user verification status checking.
// This scenario covers:
// - Successful verification status retrieval for verified users
// - Handling of non-verified users
// - HTTP error responses (500, etc.)
// - Malformed JSON responses
// - Missing status field in response
func TestIsUserVerified(t *testing.T) {
	tests := []struct {
		name          string
		responseBody  string
		statusCode    int
		expectErr     bool
		expectVerified bool
		description   string
	}{
		{
			name:          "verified_user",
			responseBody:  `{"result":{"status":"VERIFIED"}}`,
			statusCode:    http.StatusOK,
			expectErr:     false,
			expectVerified: true,
			description:   "user is verified",
		},
		{
			name:          "unverified_user",
			responseBody:  `{"result":{"status":"NOT_VERIFIED"}}`,
			statusCode:    http.StatusOK,
			expectErr:     false,
			expectVerified: false,
			description:   "user is not verified",
		},
		{
			name:          "pending_status",
			responseBody:  `{"result":{"status":"PENDING"}}`,
			statusCode:    http.StatusOK,
			expectErr:     false,
			expectVerified: false,
			description:   "user verification is pending",
		},
		{
			name:          "server_error",
			responseBody:  `{"error":"server error"}`,
			statusCode:    http.StatusInternalServerError,
			expectErr:     true,
			expectVerified: false,
			description:   "server returns 500 error",
		},
		{
			name:          "malformed_json",
			responseBody:  `not a json`,
			statusCode:    http.StatusOK,
			expectErr:     true,
			expectVerified: false,
			description:   "response body is not valid JSON",
		},
		{
			name:          "missing_status_field",
			responseBody:  `{"result":{}}`,
			statusCode:    http.StatusOK,
			expectErr:     false,
			expectVerified: false,
			description:   "status field is missing from result",
		},
		{
			name:          "empty_status",
			responseBody:  `{"result":{"status":""}}`,
			statusCode:    http.StatusOK,
			expectErr:     false,
			expectVerified: false,
			description:   "status field is empty string",
		},
		{
			name:          "unauthorized",
			responseBody:  `{"error":"unauthorized"}`,
			statusCode:    http.StatusUnauthorized,
			expectErr:     true,
			expectVerified: false,
			description:   "request is unauthorized (401)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.responseBody))
			}))
			defer ts.Close()

			client := NewKYCClient(ts.URL, "testdomain", nil)
			verified, err := client.IsUserVerified(context.Background(), "testaddress")

			if (err != nil) != tt.expectErr {
				t.Errorf("IsUserVerified() error = %v, wantErr %v (%s)", err, tt.expectErr, tt.description)
			}
			if verified != tt.expectVerified {
				t.Errorf("IsUserVerified() verified = %v, want %v (%s)", verified, tt.expectVerified, tt.description)
			}
		})
	}
}

// TestIsUserVerified_EmptyAddress tests validation of input parameters.
// This scenario covers:
// - Rejection of empty sponsor address
// - Rejection of whitespace-only address
func TestIsUserVerified_EmptyAddress(t *testing.T) {
	tests := []struct {
		name    string
		address string
	}{
		{"empty_string", ""},
		{"whitespace_only", "   "},
		{"newlines", "\n\t"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewKYCClient("http://localhost", "testdomain", nil)
			verified, err := client.IsUserVerified(context.Background(), tt.address)

			if err == nil {
				t.Errorf("IsUserVerified() expected error for address %q, got nil", tt.address)
			}
			if verified {
				t.Errorf("IsUserVerified() expected verified=false for invalid address, got true")
			}
		})
	}
}

// TestIsUserVerified_ContextCancelled tests behavior with cancelled context.
// This scenario covers:
// - Proper error handling when request context is cancelled
// - Early termination without making HTTP requests
func TestIsUserVerified_ContextCancelled(t *testing.T) {
	client := NewKYCClient("http://localhost", "testdomain", nil)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	verified, err := client.IsUserVerified(ctx, "testaddress")
	if err == nil {
		t.Errorf("IsUserVerified() with cancelled context expected error, got nil")
	}
	if verified {
		t.Errorf("IsUserVerified() with cancelled context expected verified=false, got true")
	}
}

// TestIsValidSponsor tests sponsor validation checking.
// This scenario covers:
// - Valid sponsor with VERIFIED status and idenfyRef
// - Verified user but missing idenfyRef
// - Non-verified users cannot be sponsors
// - HTTP error responses
// - Missing response fields
func TestIsValidSponsor(t *testing.T) {
	tests := []struct {
		name          string
		responseBody  string
		statusCode    int
		expectErr     bool
		expectValid   bool
		description   string
	}{
		{
			name:        "valid_sponsor",
			responseBody: `{"result":{"status":"VERIFIED","idenfyRef":"ref123"}}`,
			statusCode:  http.StatusOK,
			expectErr:   false,
			expectValid: true,
			description: "user is verified with idenfyRef",
		},
		{
			name:        "verified_no_idenfy_ref",
			responseBody: `{"result":{"status":"VERIFIED","idenfyRef":""}}`,
			statusCode:  http.StatusOK,
			expectErr:   false,
			expectValid: false,
			description: "verified user but missing idenfyRef",
		},
		{
			name:        "unverified_user",
			responseBody: `{"result":{"status":"NOT_VERIFIED","idenfyRef":"ref123"}}`,
			statusCode:  http.StatusOK,
			expectErr:   false,
			expectValid: false,
			description: "unverified user cannot be sponsor",
		},
		{
			name:        "server_error",
			responseBody: `{"error":"server error"}`,
			statusCode:  http.StatusInternalServerError,
			expectErr:   true,
			expectValid: false,
			description: "server returns 500 error",
		},
		{
			name:        "malformed_json",
			responseBody: `invalid json`,
			statusCode:  http.StatusOK,
			expectErr:   true,
			expectValid: false,
			description: "response body is not valid JSON",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.responseBody))
			}))
			defer ts.Close()

			client := NewKYCClient(ts.URL, "testdomain", nil)
			valid, err := client.IsValidSponsor(context.Background(), "testaddress")

			if (err != nil) != tt.expectErr {
				t.Errorf("IsValidSponsor() error = %v, wantErr %v (%s)", err, tt.expectErr, tt.description)
			}
			if valid != tt.expectValid {
				t.Errorf("IsValidSponsor() valid = %v, want %v (%s)", valid, tt.expectValid, tt.description)
			}
		})
	}
}

// TestIsValidSponsor_EmptyAddress tests validation of input parameters.
// This scenario covers:
// - Rejection of empty sponsor address
// - Rejection of whitespace-only address
func TestIsValidSponsor_EmptyAddress(t *testing.T) {
	tests := []struct {
		name    string
		address string
	}{
		{"empty_string", ""},
		{"whitespace_only", "   "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewKYCClient("http://localhost", "testdomain", nil)
			valid, err := client.IsValidSponsor(context.Background(), tt.address)

			if err == nil {
				t.Errorf("IsValidSponsor() expected error for address %q, got nil", tt.address)
			}
			if valid {
				t.Errorf("IsValidSponsor() expected valid=false for invalid address, got true")
			}
		})
	}
}

// TestCreateSponsorship tests sponsorship creation with various scenarios.
// This scenario covers:
// - Successful sponsorship creation (201 response)
// - HTTP error responses (400, 401, 500)
// - Reading error response bodies for debugging
// - Proper header construction with signatures
func TestCreateSponsorship(t *testing.T) {
	sponsorKp, _ := sr25519.Scheme{}.FromPhrase("bottom drive obey lake curtain smoke basket hold race lonely fit walk", "")
	sponseeKp, _ := sr25519.Scheme{}.FromPhrase("bottom drive obey lake curtain smoke basket hold race lonely fit walk", "")

	tests := []struct {
		name           string
		statusCode     int
		responseBody   string
		expectErr      bool
		description    string
	}{
		{
			name:        "success",
			statusCode:  http.StatusCreated,
			responseBody: `{"result":"ok"}`,
			expectErr:   false,
			description: "sponsorship created successfully",
		},
		{
			name:        "bad_request",
			statusCode:  http.StatusBadRequest,
			responseBody: `{"error":"invalid signature"}`,
			expectErr:   true,
			description: "server rejects request with 400",
		},
		{
			name:        "unauthorized",
			statusCode:  http.StatusUnauthorized,
			responseBody: `{"error":"unauthorized"}`,
			expectErr:   true,
			description: "authentication fails (401)",
		},
		{
			name:        "conflict",
			statusCode:  http.StatusConflict,
			responseBody: `{"error":"sponsorship already exists"}`,
			expectErr:   true,
			description: "sponsorship already exists (409)",
		},
		{
			name:        "server_error",
			statusCode:  http.StatusInternalServerError,
			responseBody: `{"error":"database error"}`,
			expectErr:   true,
			description: "server error (500)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request method and path
				if r.Method != http.MethodPost || r.URL.Path != "/api/v1/sponsorships" {
					t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
				}
				// Verify required headers are present
				if r.Header.Get("X-Client-ID") == "" {
					t.Errorf("missing X-Client-ID header")
				}
				if r.Header.Get("X-Signature") == "" {
					t.Errorf("missing X-Signature header")
				}
				
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.responseBody))
			}))
			defer ts.Close()

			client := NewKYCClient(ts.URL, "testdomain", nil)
			err := client.CreateSponsorship(context.Background(), "sponsor", sponsorKp, "sponsee", sponseeKp)

			if (err != nil) != tt.expectErr {
				t.Errorf("CreateSponsorship() error = %v, wantErr %v (%s)", err, tt.expectErr, tt.description)
			}
			if err != nil && !strings.Contains(err.Error(), tt.responseBody) {
				t.Logf("CreateSponsorship() error message: %v", err)
			}
		})
	}
}

// TestCreateSponsorship_InvalidInput tests validation of sponsorship parameters.
// This scenario covers:
// - Rejection of empty sponsor address
// - Rejection of empty sponsee address
// - Rejection of whitespace-only addresses
func TestCreateSponsorship_InvalidInput(t *testing.T) {
	validKp, _ := sr25519.Scheme{}.FromPhrase("bottom drive obey lake curtain smoke basket hold race lonely fit walk", "")

	tests := []struct {
		name        string
		sponsorAddr string
		sponseeAddr string
		description string
	}{
		{"empty_sponsor_addr", "", "sponsee", "empty sponsor address"},
		{"empty_sponsee_addr", "sponsor", "", "empty sponsee address"},
		{"whitespace_sponsor", "   ", "sponsee", "whitespace sponsor address"},
		{"whitespace_sponsee", "sponsor", "   ", "whitespace sponsee address"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewKYCClient("http://localhost", "testdomain", nil)
			err := client.CreateSponsorship(context.Background(), tt.sponsorAddr, validKp, tt.sponseeAddr, validKp)

			if err == nil {
				t.Errorf("CreateSponsorship() expected error for %s, got nil", tt.description)
			}
		})
	}
}

// TestCreateChallenge tests internal challenge message creation.
// This scenario covers:
// - Challenge format correctness (domain:timestamp)
// - Challenge includes the configured domain
// - Challenge timestamp portion is numeric (Unix timestamp)
func TestCreateChallenge(t *testing.T) {
	client := NewKYCClient("http://localhost", "mydomain.test", nil)
	
	challenge := client.createChallengeMessage()

	// Challenge should have the format "domain:timestamp"
	parts := strings.Split(challenge, ":")
	if len(parts) != 2 {
		t.Errorf("challenge format incorrect, expected 'domain:timestamp', got %q", challenge)
	}
	if parts[0] != "mydomain.test" {
		t.Errorf("challenge domain mismatch, expected 'mydomain.test', got %q", parts[0])
	}

	// Timestamp portion should be numeric (Unix timestamp)
	if _, err := strconv.ParseInt(parts[1], 10, 64); err != nil {
		t.Errorf("challenge timestamp should be numeric Unix timestamp, got %q: %v", parts[1], err)
	}
}

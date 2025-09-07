package internal

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	retryablehttp "github.com/hashicorp/go-retryablehttp"
	"github.com/vedhavyas/go-subkey"
	"kubecloud/internal/logger"
)

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// KYCClient holds configuration for tf-kyc-verifier API client
type KYCClient struct {
	APIURL          string
	ChallengeDomain string
	httpClient      httpClient
}

// NewKYCClient creates a new KYCClient instance. If client is nil, creates a retryable client with default settings.
func NewKYCClient(apiURL, challengeDomain string, client httpClient) *KYCClient {
	if client == nil {
		retryClient := retryablehttp.NewClient()
		retryClient.RetryMax = 3
		retryClient.RetryWaitMin = 500 * time.Millisecond
		retryClient.RetryWaitMax = 2 * time.Second
		retryClient.Logger = nil              // Disable default logger
		client = retryClient.StandardClient() // Get *http.Client that does retries
	}

	return &KYCClient{
		APIURL:          apiURL,
		ChallengeDomain: challengeDomain,
		httpClient:      client,
	}
}

// createChallengeMessage creates the challenge message string
func (c *KYCClient) createChallengeMessage() string {
	timestamp := time.Now().Unix()
	return fmt.Sprintf("%s:%d", c.ChallengeDomain, timestamp)
}

// signMessage signs the message with given keypair and returns hex encoded signature
func signMessage(kr subkey.KeyPair, message string) (string, error) {
	sig, err := kr.Sign([]byte(message)) // Sign raw bytes, not hash
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(sig), nil
}

// CreateSponsorship creates a sponsorship between sponsor and sponsee addresses
func (c *KYCClient) CreateSponsorship(ctx context.Context, sponsorAddress string, sponsorKeyPair subkey.KeyPair, sponseeAddress string, sponseeKeyPair subkey.KeyPair) error {
	if strings.TrimSpace(sponsorAddress) == "" {
		return fmt.Errorf("sponsor address is empty")
	}
	if sponsorKeyPair == nil {
		return fmt.Errorf("sponsor keypair is nil")
	}
	if strings.TrimSpace(sponseeAddress) == "" {
		return fmt.Errorf("sponsee address is empty")
	}
	if sponseeKeyPair == nil {
		return fmt.Errorf("sponsee keypair is nil")
	}

	// Create challenge messages
	sponsorChallenge := c.createChallengeMessage()
	sponseeChallenge := c.createChallengeMessage()

	// Sign challenges
	sponsorSignature, err := signMessage(sponsorKeyPair, sponsorChallenge)
	if err != nil {
		return fmt.Errorf("failed to sign sponsor challenge: %w", err)
	}
	sponseeSignature, err := signMessage(sponseeKeyPair, sponseeChallenge)
	if err != nil {
		return fmt.Errorf("failed to sign sponsee challenge: %w", err)
	}

	// Debug logs for troubleshooting
	logger.GetLogger().Debug().Msgf("KYC Sponsorship Debug: sponsorAddress=%s, sponseeAddress=%s", sponsorAddress, sponseeAddress)
	logger.GetLogger().Debug().Msgf("KYC Sponsorship Debug: sponsorChallenge=%s, sponseeChallenge=%s", sponsorChallenge, sponseeChallenge)
	logger.GetLogger().Debug().Msgf("KYC Sponsorship Debug: sponsorSignature=%s, sponseeSignature=%s", sponsorSignature, sponseeSignature)

	// Prepare HTTP request
	url := fmt.Sprintf("%s/api/v1/sponsorships", c.APIURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer([]byte{}))
	if err != nil {
		return fmt.Errorf("creating HTTP request to %s: %w", url, err)
	}

	// Set required headers
	req.Header.Set("X-Client-ID", sponsorAddress)
	req.Header.Set("X-Sponsee-ID", sponseeAddress)
	req.Header.Set("X-Challenge", hex.EncodeToString([]byte(sponsorChallenge)))
	req.Header.Set("X-Sponsee-Challenge", hex.EncodeToString([]byte(sponseeChallenge)))
	req.Header.Set("X-Signature", sponsorSignature)
	req.Header.Set("X-Sponsee-Signature", sponseeSignature)

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send sponsorship request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		// Read response body for error details
		bodyBytes := new(bytes.Buffer)
		_, err := bodyBytes.ReadFrom(resp.Body)
		if err != nil {
			return fmt.Errorf("sponsorship creation failed with status: %s and failed to read response body: %w", resp.Status, err)
		}
		return fmt.Errorf("sponsorship creation failed with status: %s and response: %s", resp.Status, bodyBytes.String())
	}

	logger.GetLogger().Info().Msgf("Sponsorship created successfully between sponsor %s and sponsee %s", sponsorAddress, sponseeAddress)
	return nil
}

// IsUserVerified checks if a user is verified (directly or via sponsorship) by calling the tf-kyc-verifier API
func (c *KYCClient) IsUserVerified(ctx context.Context, sponseeAddress string) (bool, error) {
	if strings.TrimSpace(sponseeAddress) == "" {
		return false, fmt.Errorf("sponsee address is empty")
	}
	url := fmt.Sprintf("%s/api/v1/status?client_id=%s", c.APIURL, sponseeAddress)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return false, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("KYC verifier returned status: %s", resp.Status)
	}

	var result struct {
		Result struct {
			Status string `json:"status"`
		} `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, err
	}

	return result.Result.Status == "VERIFIED", nil
}

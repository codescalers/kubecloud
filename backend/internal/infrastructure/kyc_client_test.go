package infrastructure

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/vedhavyas/go-subkey/sr25519"
)

func TestIsUserVerified(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/status" && r.URL.Query().Get("client_id") == "dummyaddress" {
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte(`{"result":{"status":"VERIFIED"}}`)); err != nil {
				t.Errorf("failed to write response: %v", err)
			}
			return
		}
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{"result":{"status":"NOT_VERIFIED"}}`)); err != nil {
			t.Errorf("failed to write response: %v", err)
		}
	}))
	defer ts.Close()

	client := NewKYCClient(ts.URL, "testdomain", nil)

	verified, err := client.IsUserVerified(context.Background(), "dummyaddress")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !verified {
		t.Errorf("expected verified to be true for status VERIFIED")
	}

	verified, err = client.IsUserVerified(context.Background(), "otheraddress")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if verified {
		t.Errorf("expected verified to be false for status NOT_VERIFIED")
	}
}

func TestIsUserVerified_Non200Status(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := w.Write([]byte(`{"error":"server error"}`)); err != nil {
			t.Errorf("failed to write response: %v", err)
		}
	}))
	defer ts.Close()

	client := NewKYCClient(ts.URL, "testdomain", nil)
	verified, err := client.IsUserVerified(context.Background(), "dummyaddress")
	if err == nil {
		t.Errorf("expected error for non-200 status")
	}
	if verified {
		t.Errorf("expected verified to be false on error")
	}
}

func TestIsUserVerified_MalformedJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`not a json`)); err != nil {
			t.Errorf("failed to write response: %v", err)
		}
	}))
	defer ts.Close()

	client := NewKYCClient(ts.URL, "testdomain", nil)
	verified, err := client.IsUserVerified(context.Background(), "dummyaddress")
	if err == nil {
		t.Errorf("expected error for malformed JSON")
	}
	if verified {
		t.Errorf("expected verified to be false on error")
	}
}

func TestIsUserVerified_MissingStatusField(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{"result":{}}`)); err != nil {
			t.Errorf("failed to write response: %v", err)
		}
	}))
	defer ts.Close()

	client := NewKYCClient(ts.URL, "testdomain", nil)
	verified, err := client.IsUserVerified(context.Background(), "dummyaddress")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if verified {
		t.Errorf("expected verified to be false when status field is missing")
	}
}

func TestCreateSponsorship(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/sponsorships" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusCreated)
		if _, err := w.Write([]byte(`{"result":"ok"}`)); err != nil {
			t.Errorf("failed to write response: %v", err)
		}
	}))
	defer ts.Close()

	client := NewKYCClient(ts.URL, "testdomain", nil)
	TEST_MNEMONIC := os.Getenv("TEST_MNEMONIC")
	if TEST_MNEMONIC == "" {
		t.Fatal("TEST_MNEMONIC environment variable is not set")
	}
	kp, _ := sr25519.Scheme{}.FromPhrase(TEST_MNEMONIC, "")
	err := client.CreateSponsorship(context.Background(), "sponsor", kp, "sponsee", kp)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestCreateSponsorship_Non201(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		if _, err := w.Write([]byte(`{"error":"bad request"}`)); err != nil {
			t.Errorf("failed to write response: %v", err)
		}
	}))
	defer ts.Close()

	client := NewKYCClient(ts.URL, "testdomain", nil)
	TEST_MNEMONIC := os.Getenv("TEST_MNEMONIC")
	if TEST_MNEMONIC == "" {
		t.Fatal("TEST_MNEMONIC environment variable is not set")
	}
	kp, _ := sr25519.Scheme{}.FromPhrase(TEST_MNEMONIC, "")
	err := client.CreateSponsorship(context.Background(), "sponsor", kp, "sponsee", kp)
	if err == nil {
		t.Errorf("expected error for non-201 response")
	}
	if !strings.Contains(err.Error(), "bad request") {
		t.Errorf("expected error message to contain 'bad request', got %v", err)
	}
}

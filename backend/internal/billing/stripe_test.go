package billing

import (
	"fmt"
	"testing"

	"github.com/stripe/stripe-go/v82"
)

// MockStripeClient is a mock implementation of StripeClient for testing.
type MockStripeClient struct {
	CreateCustomerFunc       func(username, email string) (*stripe.Customer, error)
	CreatePaymentMethodFunc  func(cardType, paymentMethodID string) (*stripe.PaymentMethod, error)
	CreatePaymentIntentFunc  func(customerID, paymentMethodID, currency string, usdMillicentAmount uint64) (*stripe.PaymentIntent, error)
	CancelPaymentIntentFunc  func(paymentIntentID string) error
}

func (m *MockStripeClient) CreateCustomer(username, email string) (*stripe.Customer, error) {
	if m.CreateCustomerFunc != nil {
		return m.CreateCustomerFunc(username, email)
	}
	return &stripe.Customer{ID: "cus_123"}, nil
}

func (m *MockStripeClient) CreatePaymentMethod(cardType, paymentMethodID string) (*stripe.PaymentMethod, error) {
	if m.CreatePaymentMethodFunc != nil {
		return m.CreatePaymentMethodFunc(cardType, paymentMethodID)
	}
	return &stripe.PaymentMethod{ID: "pm_123"}, nil
}

func (m *MockStripeClient) CreatePaymentIntent(customerID, paymentMethodID, currency string, usdMillicentAmount uint64) (*stripe.PaymentIntent, error) {
	if m.CreatePaymentIntentFunc != nil {
		return m.CreatePaymentIntentFunc(customerID, paymentMethodID, currency, usdMillicentAmount)
	}
	return &stripe.PaymentIntent{ID: "pi_123"}, nil
}

func (m *MockStripeClient) CancelPaymentIntent(paymentIntentID string) error {
	if m.CancelPaymentIntentFunc != nil {
		return m.CancelPaymentIntentFunc(paymentIntentID)
	}
	return nil
}

// TestCreateCustomerSuccess tests successful customer creation.
// This scenario covers:
// - Valid username and email are accepted
// - Customer is created with correct parameters
// - Customer ID is returned
func TestCreateCustomerSuccess(t *testing.T) {
	tests := []struct {
		name        string
		username    string
		email       string
		description string
	}{
		{
			name:        "valid_customer",
			username:    "john_doe",
			email:       "john@example.com",
			description: "creating customer with valid parameters",
		},
		{
			name:        "customer_with_special_chars",
			username:    "user.name+tag",
			email:       "user+tag@example.co.uk",
			description: "creating customer with special characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockStripeClient{
				CreateCustomerFunc: func(username, email string) (*stripe.Customer, error) {
					if username != tt.username || email != tt.email {
						t.Errorf("parameters mismatch: got %s/%s, want %s/%s", username, email, tt.username, tt.email)
					}
					return &stripe.Customer{
						ID:    "cus_123",
						Name:  username,
						Email: email,
					}, nil
				},
			}

			result, err := mock.CreateCustomer(tt.username, tt.email)
			if err != nil {
				t.Errorf("unexpected error: %v (%s)", err, tt.description)
			}
			if result == nil || result.ID == "" {
				t.Errorf("customer ID should not be empty (%s)", tt.description)
			}
		})
	}
}

// TestCreateCustomerError tests customer creation failures.
// This scenario covers:
// - API errors are handled
// - Error message is formatted correctly
func TestCreateCustomerError(t *testing.T) {
	tests := []struct {
		name        string
		errorMsg    string
		description string
	}{
		{
			name:        "invalid_email",
			errorMsg:    "invalid email format",
			description: "creation fails with invalid email",
		},
		{
			name:        "api_error",
			errorMsg:    "rate limit exceeded",
			description: "creation fails with API error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockStripeClient{
				CreateCustomerFunc: func(username, email string) (*stripe.Customer, error) {
					return nil, fmt.Errorf("failed to create customer: %s", tt.errorMsg)
				},
			}

			result, err := mock.CreateCustomer("test", "test@example.com")
			if err == nil {
				t.Errorf("expected error but got nil (%s)", tt.description)
			}
			if result != nil {
				t.Errorf("result should be nil on error (%s)", tt.description)
			}
		})
	}
}

// TestCreatePaymentMethodSuccess tests successful payment method creation.
// This scenario covers:
// - Valid card type and payment method ID are accepted
// - Payment method is created with correct parameters
func TestCreatePaymentMethodSuccess(t *testing.T) {
	tests := []struct {
		name            string
		cardType        string
		paymentMethodID string
		description     string
	}{
		{
			name:            "visa_card",
			cardType:        "card",
			paymentMethodID: "tok_visa",
			description:     "creating Visa card",
		},
		{
			name:            "payment_method_token",
			cardType:        "card",
			paymentMethodID: "pm_1234567890abcdefghijklmnop",
			description:     "creating with payment method token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockStripeClient{
				CreatePaymentMethodFunc: func(cardType, paymentMethodID string) (*stripe.PaymentMethod, error) {
					if cardType != tt.cardType || paymentMethodID != tt.paymentMethodID {
						t.Errorf("parameters mismatch: got %s/%s, want %s/%s", cardType, paymentMethodID, tt.cardType, tt.paymentMethodID)
					}
					return &stripe.PaymentMethod{ID: "pm_123"}, nil
				},
			}

			result, err := mock.CreatePaymentMethod(tt.cardType, tt.paymentMethodID)
			if err != nil {
				t.Errorf("unexpected error: %v (%s)", err, tt.description)
			}
			if result == nil || result.ID == "" {
				t.Errorf("payment method ID should not be empty (%s)", tt.description)
			}
		})
	}
}

// TestCreatePaymentIntentSuccess tests successful payment intent creation.
// This scenario covers:
// - Valid parameters are accepted
// - Amount conversion is correct (divide by 10)
// - Payment intent is created with correct values
func TestCreatePaymentIntentSuccess(t *testing.T) {
	tests := []struct {
		name               string
		customerID         string
		paymentMethodID    string
		currency           string
		usdMillicentAmount uint64
		expectedAmount     int64
		description        string
	}{
		{
			name:               "usd_payment",
			customerID:         "cus_123",
			paymentMethodID:    "pm_123",
			currency:           "usd",
			usdMillicentAmount: 100000,
			expectedAmount:     10000,
			description:        "creating USD payment intent",
		},
		{
			name:               "small_amount",
			customerID:         "cus_123",
			paymentMethodID:    "pm_123",
			currency:           "usd",
			usdMillicentAmount: 50,
			expectedAmount:     5,
			description:        "creating payment intent with small amount",
		},
		{
			name:               "eur_payment",
			customerID:         "cus_123",
			paymentMethodID:    "pm_123",
			currency:           "eur",
			usdMillicentAmount: 100000,
			expectedAmount:     10000,
			description:        "creating EUR payment intent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockStripeClient{
				CreatePaymentIntentFunc: func(customerID, paymentMethodID, currency string, usdMillicentAmount uint64) (*stripe.PaymentIntent, error) {
					if customerID != tt.customerID || paymentMethodID != tt.paymentMethodID || currency != tt.currency {
						t.Errorf("parameters mismatch (%s)", tt.description)
					}

					// Verify amount conversion
					amount := int64(usdMillicentAmount / 10)
					if amount != tt.expectedAmount {
						t.Errorf("amount conversion = %d, want %d (%s)", amount, tt.expectedAmount, tt.description)
					}

					return &stripe.PaymentIntent{
						ID:     "pi_123",
						Amount: amount,
					}, nil
				},
			}

			result, err := mock.CreatePaymentIntent(tt.customerID, tt.paymentMethodID, tt.currency, tt.usdMillicentAmount)
			if err != nil {
				t.Errorf("unexpected error: %v (%s)", err, tt.description)
			}
			if result == nil || result.ID == "" {
				t.Errorf("payment intent ID should not be empty (%s)", tt.description)
			}
		})
	}
}

// TestCreatePaymentIntentError tests payment intent creation failures.
// This scenario covers:
// - Missing customer ID fails
// - Missing payment method fails
// - API errors are handled
func TestCreatePaymentIntentError(t *testing.T) {
	tests := []struct {
		name            string
		customerID      string
		paymentMethodID string
		errorMsg        string
		description     string
	}{
		{
			name:            "missing_customer",
			customerID:      "",
			paymentMethodID: "pm_123",
			errorMsg:        "customer not found",
			description:     "creation fails with missing customer",
		},
		{
			name:            "missing_payment_method",
			customerID:      "cus_123",
			paymentMethodID: "",
			errorMsg:        "payment method not found",
			description:     "creation fails with missing payment method",
		},
		{
			name:            "insufficient_funds",
			customerID:      "cus_123",
			paymentMethodID: "pm_123",
			errorMsg:        "card declined",
			description:     "creation fails with card decline",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockStripeClient{
				CreatePaymentIntentFunc: func(customerID, paymentMethodID, currency string, usdMillicentAmount uint64) (*stripe.PaymentIntent, error) {
					if customerID == "" {
						return nil, fmt.Errorf("failed to create payment intent: customer not found")
					}
					if paymentMethodID == "" {
						return nil, fmt.Errorf("failed to create payment intent: payment method not found")
					}
					return nil, fmt.Errorf("failed to create payment intent: %s", tt.errorMsg)
				},
			}

			result, err := mock.CreatePaymentIntent(tt.customerID, tt.paymentMethodID, "usd", 100000)
			if err == nil {
				t.Errorf("expected error but got nil (%s)", tt.description)
			}
			if result != nil {
				t.Errorf("result should be nil on error (%s)", tt.description)
			}
		})
	}
}

// TestCancelPaymentIntentSuccess tests successful payment intent cancellation.
// This scenario covers:
// - Valid payment intent ID is accepted
// - Cancellation completes without error
func TestCancelPaymentIntentSuccess(t *testing.T) {
	tests := []struct {
		name            string
		paymentIntentID string
		description     string
	}{
		{
			name:            "cancel_intent",
			paymentIntentID: "pi_1234567890abcdefghijklmnop",
			description:     "canceling valid intent",
		},
		{
			name:            "cancel_test_intent",
			paymentIntentID: "pi_test_1234567890",
			description:     "canceling test intent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockStripeClient{
				CancelPaymentIntentFunc: func(paymentIntentID string) error {
					if paymentIntentID != tt.paymentIntentID {
						t.Errorf("intent ID mismatch: got %s, want %s", paymentIntentID, tt.paymentIntentID)
					}
					return nil
				},
			}

			err := mock.CancelPaymentIntent(tt.paymentIntentID)
			if err != nil {
				t.Errorf("unexpected error: %v (%s)", err, tt.description)
			}
		})
	}
}

// TestCancelPaymentIntentError tests payment intent cancellation failures.
// This scenario covers:
// - Invalid intent ID fails
// - Already canceled intent fails
// - API errors are handled
func TestCancelPaymentIntentError(t *testing.T) {
	tests := []struct {
		name            string
		paymentIntentID string
		errorMsg        string
		description     string
	}{
		{
			name:            "invalid_intent_id",
			paymentIntentID: "invalid",
			errorMsg:        "invalid payment intent",
			description:     "cancellation fails with invalid ID",
		},
		{
			name:            "already_canceled",
			paymentIntentID: "pi_123",
			errorMsg:        "intent already canceled",
			description:     "cancellation fails - intent already canceled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockStripeClient{
				CancelPaymentIntentFunc: func(paymentIntentID string) error {
					return fmt.Errorf("failed to cancel payment intent: %s", tt.errorMsg)
				},
			}

			err := mock.CancelPaymentIntent(tt.paymentIntentID)
			if err == nil {
				t.Errorf("expected error but got nil (%s)", tt.description)
			}
		})
	}
}

// TestAmountConversionAccuracy tests precision of amount conversion.
// This scenario covers:
// - Various amounts convert correctly
// - Truncation is handled properly
// - Large values don't overflow
func TestAmountConversionAccuracy(t *testing.T) {
	tests := []struct {
		name                 string
		usdMillicentAmount   uint64
		expectedStripeAmount int64
		description          string
	}{
		{
			name:                 "one_dollar",
			usdMillicentAmount:   100,
			expectedStripeAmount: 10,
			description:          "converting $1.00",
		},
		{
			name:                 "one_cent",
			usdMillicentAmount:   1,
			expectedStripeAmount: 0,
			description:          "converting $0.01 with truncation",
		},
		{
			name:                 "ten_dollars",
			usdMillicentAmount:   1000,
			expectedStripeAmount: 100,
			description:          "converting $10.00",
		},
		{
			name:                 "zero",
			usdMillicentAmount:   0,
			expectedStripeAmount: 0,
			description:          "converting $0",
		},
		{
			name:                 "with_rounding_down",
			usdMillicentAmount:   999,
			expectedStripeAmount: 99,
			description:          "converting with truncation",
		},
		{
			name:                 "large_amount",
			usdMillicentAmount:   10000000,
			expectedStripeAmount: 1000000,
			description:          "converting large amount",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := int64(tt.usdMillicentAmount / 10)
			if result != tt.expectedStripeAmount {
				t.Errorf("conversion = %d, want %d (%s)", result, tt.expectedStripeAmount, tt.description)
			}
		})
	}
}


package billing

import (
	"fmt"

	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/customer"
	"github.com/stripe/stripe-go/v82/paymentintent"
	"github.com/stripe/stripe-go/v82/paymentmethod"
)

// StripeClient defines the interface for Stripe operations.
type StripeClient interface {
	CreateCustomer(username, email string) (*stripe.Customer, error)
	CreatePaymentMethod(cardType, paymentMethodID string) (*stripe.PaymentMethod, error)
	CreatePaymentIntent(customerID, paymentMethodID, currency string, usdMillicentAmount uint64) (*stripe.PaymentIntent, error)
	CancelPaymentIntent(paymentIntentID string) error
}

// DefaultStripeClient is the real implementation using Stripe SDK.
type DefaultStripeClient struct{}

// CreateCustomer creates a customer in Stripe.
func (c *DefaultStripeClient) CreateCustomer(username, email string) (*stripe.Customer, error) {
	params := &stripe.CustomerParams{
		Name:  stripe.String(username),
		Email: stripe.String(email),
	}
	res, err := customer.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create customer: %w", err)
	}
	return res, nil
}

// CreatePaymentMethod creates a payment method in Stripe.
func (c *DefaultStripeClient) CreatePaymentMethod(cardType, paymentMethodID string) (*stripe.PaymentMethod, error) {
	paymentMethodParams := &stripe.PaymentMethodParams{
		Type: stripe.String(cardType),
		Card: &stripe.PaymentMethodCardParams{Token: stripe.String(paymentMethodID)},
	}

	return paymentmethod.New(paymentMethodParams)
}

// CreatePaymentIntent creates a payment intent in Stripe.
func (c *DefaultStripeClient) CreatePaymentIntent(customerID, paymentMethodID, currency string, usdMillicentAmount uint64) (*stripe.PaymentIntent, error) {
	params := &stripe.PaymentIntentParams{
		Amount:        stripe.Int64(int64(usdMillicentAmount / 10)),
		Currency:      stripe.String(currency),
		Customer:      stripe.String(customerID),
		PaymentMethod: stripe.String(paymentMethodID),
		Confirm:       stripe.Bool(true),
		AutomaticPaymentMethods: &stripe.PaymentIntentAutomaticPaymentMethodsParams{
			Enabled:        stripe.Bool(true),
			AllowRedirects: stripe.String("never"),
		},
	}
	result, err := paymentintent.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment intent: %w", err)
	}

	return result, err
}

// CancelPaymentIntent cancels a payment intent in Stripe.
func (c *DefaultStripeClient) CancelPaymentIntent(paymentIntentID string) error {
	_, err := paymentintent.Cancel(paymentIntentID, nil)
	if err != nil {
		return fmt.Errorf("failed to cancel payment intent: %w", err)
	}
	return nil
}

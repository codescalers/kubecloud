package internal

import (
	"fmt"

	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/customer"
	"github.com/stripe/stripe-go/v82/paymentintent"
	"github.com/stripe/stripe-go/v82/paymentmethod"
)

// CreateStripeCustomer create a customer in stripe
func CreateStripeCustomer(username, email string) (*stripe.Customer, error) {
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

func CreatePaymentMethod(cardType, paymentMethodID string) (*stripe.PaymentMethod, error) {
	paymentMethodParams := &stripe.PaymentMethodParams{
		Type: stripe.String(cardType),
		Card: &stripe.PaymentMethodCardParams{Token: stripe.String(paymentMethodID)},
	}

	return paymentmethod.New(paymentMethodParams)
}

func CreatePaymentIntent(customerID, paymentMethodID, currency string, amount float64) (*stripe.PaymentIntent, error) {
	params := &stripe.PaymentIntentParams{
		Amount:        stripe.Int64(int64(amount * 100)),
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

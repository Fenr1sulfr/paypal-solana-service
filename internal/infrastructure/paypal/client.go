package paypal

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"payment.Diploma.service/internal/domain/payment"

	"github.com/plutov/paypal/v4"
)

// PayPalClient wraps the plutov/paypal client to implement the payment.PayPalClient interface.
type PayPalClient struct {
	client *paypal.Client
}

// NewPayPalClient initializes a new PayPal client using the plutov/paypal package.
func NewPayPalClient(clientID, clientSecret string, sandbox bool) payment.PayPalClient {
	var apiBase string
	if sandbox {
		apiBase = paypal.APIBaseSandBox
	} else {
		apiBase = paypal.APIBaseLive
	}

	client, err := paypal.NewClient(clientID, clientSecret, apiBase)
	if err != nil {
		// In a production system, you might want to handle this error differently
		// (e.g., panic or return an error), but here we assume the client is always valid.
		panic(fmt.Sprintf("failed to create PayPal client: %v", err))
	}
	client.SetLog(os.Stdout)
	return &PayPalClient{
		client: client,
	}
}

// CreatePayment creates a PayPal payment and returns the payment ID and approval URL.
func (c *PayPalClient) CreatePayment(ctx context.Context, amount float64, currency, description, returnURL, cancelURL string) (string, string, error) {
	amountStr := fmt.Sprintf("%.2f", amount)
	rand.Seed(time.Now().UnixNano())
	randomNumber := rand.Intn(1000000000)
	randomNumberString := strconv.Itoa(randomNumber)

	unit := paypal.PurchaseUnitRequest{
		ReferenceID: randomNumberString,
		Amount: &paypal.PurchaseUnitAmount{
			Currency: currency,
			Value:    amountStr,
		},
		Payee: &paypal.PayeeForOrders{
			EmailAddress: "bluetemimorphine@gmail.com",
		},
		Description: description, // Add description
	}

	appContext := paypal.ApplicationContext{
		BrandName:          "Agrochain",
		Locale:             "ru",
		ShippingPreference: paypal.ShippingPreferenceNoShipping,
		UserAction:         paypal.UserActionPayNow,
		PaymentMethod: paypal.PaymentMethod{
			PayeePreferred:         paypal.PayeePreferredUnrestricted,
			StandardEntryClassCode: paypal.StandardEntryClassCodeTel,
		},	
		LandingPage: "NO_PREFERENCE",
		ReturnURL:   "http://localhost:3000/payment-success", // Use parameter instead of hardcoded value
		CancelURL:   "http://localhost:3000",                 // Use parameter instead of hardcoded value
	}

	order, err := c.client.CreateOrder(ctx, paypal.OrderIntentCapture, []paypal.PurchaseUnitRequest{unit}, nil, &appContext)
	if err != nil {
		return "", "", fmt.Errorf("failed to create payment: %w", err)
	}

	if order.ID == "" {	
		return "", "", fmt.Errorf("payment creation succeeded but no payment ID returned")
	}

	var approvalURL string
	for _, link := range order.Links {
		if link.Rel == "approve" { // Changed from "approval_url" to "approve"
			approvalURL = link.Href
			break
		}
	}

	if approvalURL == "" {
		return "", "", fmt.Errorf("no approval URL found in payment response")
	}

	return order.ID, approvalURL, nil
}

// ExecutePayment executes a PayPal payment with the given payment ID and payer ID.
func (c *PayPalClient) ExecutePayment(ctx context.Context, paymentID, payerID string) (*payment.PaymentDetails, error) {
	// Execute PayPal payment
	captureOrder, err := c.client.CaptureOrder(ctx, paymentID, paypal.CaptureOrderRequest{})
	if err != nil {
		return nil, fmt.Errorf("failed to execute payment: %w", err)
	}

	// Extract payment amount and currency
	var amount float64
	var currency string
	if len(captureOrder.PurchaseUnits) > 0 && captureOrder.PurchaseUnits[0].Payments != nil &&
		len(captureOrder.PurchaseUnits[0].Payments.Captures) > 0 {
		capture := captureOrder.PurchaseUnits[0].Payments.Captures[0]
		if capture.Amount != nil {
			amountStr := capture.Amount.Value
			parsedAmount, err := strconv.ParseFloat(amountStr, 64)
			if err == nil {
				amount = parsedAmount
				currency = capture.Amount.Currency
			}
		}
	}

	// Check payment status
	if captureOrder.Status == "COMPLETED" {
		// Payment successfully completed
		return &payment.PaymentDetails{
			PaymentID:   paymentID,
			Status:      "completed",
			Amount:      amount,
			Currency:    currency,
			PayerEmail:  captureOrder.Payer.EmailAddress,
			PayerID:     payerID,
			MerchantID:  captureOrder.Payer.PayerID,
			Description: "", // Fill with transaction description if available
		}, nil
	}

	return &payment.PaymentDetails{
		PaymentID:   paymentID,
		Status:      strings.ToLower(captureOrder.Status),
		Amount:      amount,
		Currency:    currency,
		PayerID:     payerID,
		MerchantID:  captureOrder.Payer.PayerID,
		Description: "",
	}, nil
}

// GetPaymentDetails retrieves details for a specific PayPal payment.
func (c *PayPalClient) GetPaymentDetails(ctx context.Context, paymentID string) (*payment.PaymentDetails, error) {
	paymentResp, err := c.client.GetOrder(ctx, paymentID)

	if err != nil {
		return nil, fmt.Errorf("failed to get payment details: %w", err)
	}
	if len(paymentResp.PurchaseUnits) == 0 {
		return nil, fmt.Errorf("no transactions found in payment response")
	}

	transaction := paymentResp.PurchaseUnits[0]
	amount := transaction.Amount

	var payerID, payerEmail string
	if paymentResp.Payer.Name != nil {
		payerID = paymentResp.Payer.PayerID
		payerEmail = paymentResp.Payer.EmailAddress
	}

	return &payment.PaymentDetails{
		PaymentID:   paymentResp.ID,
		Status:      paymentResp.Status,
		Amount:      parseAmount(amount.Value),
		Currency:    amount.Currency,
		PayerEmail:  payerEmail,
		PayerID:     payerID,
		Description: transaction.Description,
	}, nil
}

// CancelPayment simulates canceling a payment by retrieving its details.
// PayPal doesn't support direct cancellation of created payments before execution.
func (c *PayPalClient) CancelPayment(ctx context.Context, paymentID, reason string) error {
	_, err := c.GetPaymentDetails(ctx, paymentID)
	if err != nil {
		return fmt.Errorf("failed to verify payment for cancellation: %w", err)
	}

	// Note: PayPal doesn't provide a direct API to cancel a payment after creation but before execution.
	// You would typically mark it as canceled in your local system.
	return nil
}

// parseAmount converts a string amount to a float64.
func parseAmount(amount string) float64 {
	var result float64
	fmt.Sscanf(amount, "%f", &result)
	return result
}

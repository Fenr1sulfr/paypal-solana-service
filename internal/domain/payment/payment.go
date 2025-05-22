package payment

import (
	"context"
	"time"
)

type Payment struct {
	ID              string
	PaymentID       string // PayPal payment ID
	Amount          float64
	Currency        string
	Status          string
	Description     string
	PayerEmail      string
	PayerID         string
	MerchantID      string
	SolanaSignature string
	SolanaAddress   string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	CompletedAt     time.Time
	Metadata        map[string]string	
}

// Repository interface for database operations
type Repository interface {
	Save(ctx context.Context, payment *Payment) error
	FindByID(ctx context.Context, id string) (*Payment, error)
	FindByPaymentID(ctx context.Context, paymentID string) (*Payment, error)
	UpdateStatus(ctx context.Context, id, status string) error
	ListPayments(ctx context.Context, status string, startDate, endDate time.Time, page, pageSize int) ([]*Payment, int, error)
}

// SolanaClient interface for blockchain operations
type SolanaClient interface {
	StorePayment(ctx context.Context, payment *Payment) (string, error)
	VerifyTransaction(ctx context.Context, signature string) (bool, time.Time, error)
}

// PayPalClient interface for payment gateway operations
type PayPalClient interface {
	CreatePayment(ctx context.Context, amount float64, currency, description, returnURL, cancelURL string) (string, string, error)
	ExecutePayment(ctx context.Context, paymentID, payerID string) (*PaymentDetails, error)
	GetPaymentDetails(ctx context.Context, paymentID string) (*PaymentDetails, error)
	CancelPayment(ctx context.Context, paymentID, reason string) error
}

// PaymentDetails returned from PayPal
type PaymentDetails struct {
	PaymentID   string
	Status      string
	Amount      float64
	Currency    string
	PayerEmail  string
	PayerID     string
	MerchantID  string
	Description string
}

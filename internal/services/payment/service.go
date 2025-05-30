package payment

import (
	"context"
	"errors"
	"time"

	"payment.Diploma.service/internal/domain/payment"
)

type Service struct {
	repo         payment.Repository
	solanaClient payment.SolanaClient
	paypalClient payment.PayPalClient
	baseURL      string
}

func NewService(
	repo payment.Repository,
	solanaClient payment.SolanaClient,
	paypalClient payment.PayPalClient,
	baseURL string,
) *Service {
	return &Service{
		repo:         repo,
		solanaClient: solanaClient,
		paypalClient: paypalClient,
		baseURL:      baseURL,
	}
}

func (s *Service) CreatePayment(
	ctx context.Context,
	amount float64,
	currency,
	description string,
	metadata map[string]string,
	autoStoreInSolana bool,
) (*payment.Payment, string, error) {
	returnURL := s.baseURL + "/payments/execute"
	cancelURL := s.baseURL + "/payments/cancel"

	// Create payment in PayPal
	paymentID, approvalURL, err := s.paypalClient.CreatePayment(
		ctx,
		amount,
		currency,
		description,
		returnURL,
		cancelURL,
	)
	if err != nil {
		return nil, "", err
	}

	// Create payment record
	p := &payment.Payment{
		PaymentID:   paymentID,
		Amount:      amount,
		Currency:    currency,
		Status:      "created",
		Description: description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Metadata:    metadata,
	}

	// Save to database
	if err := s.repo.Save(ctx, p); err != nil {
		return nil, "", err
	}

	return p, approvalURL, nil
}

func (s *Service) ExecutePayment(ctx context.Context, paymentID, payerID string) (*payment.Payment, error) {
	// Find payment in database
	p, err := s.repo.FindByPaymentID(ctx, paymentID)
	if err != nil {
		return nil, err
	}

	// Execute payment in PayPals
	details, err := s.paypalClient.ExecutePayment(ctx, paymentID, payerID)
	if err != nil {
		p.Status = "failed"
		s.repo.Save(ctx, p)
		return nil, err
	}

	// Update payment details
	p.Status = details.Status
	p.PayerEmail = details.PayerEmail
	p.PayerID = details.PayerID
	p.UpdatedAt = time.Now()

	if p.Status == "approved" {
		p.Status = "completed"
		p.CompletedAt = time.Now()
	}

	// Save updated payment
	if err := s.repo.Save(ctx, p); err != nil {
		return nil, err
	}

	return p, nil
}

func (s *Service) GetPayment(ctx context.Context, id string) (*payment.Payment, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *Service) GetPaymentByPaymentID(ctx context.Context, paymentID string) (*payment.Payment, error) {
	return s.repo.FindByPaymentID(ctx, paymentID)
}

func (s *Service) ListPayments(
	ctx context.Context,
	status string,
	startDate, endDate time.Time,
	page, pageSize int,
) ([]*payment.Payment, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	return s.repo.ListPayments(ctx, status, startDate, endDate, page, pageSize)
}

func (s *Service) CancelPayment(ctx context.Context, id, reason string) (*payment.Payment, error) {
	p, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if p.Status != "created" && p.Status != "pending" {
		return nil, errors.New("cannot cancel payment in current status")
	}

	// Cancel in PayPal if needed
	if err := s.paypalClient.CancelPayment(ctx, p.PaymentID, reason); err != nil {
		return nil, err
	}

	// Update status
	p.Status = "cancelled"
	p.UpdatedAt = time.Now()

	if err := s.repo.Save(ctx, p); err != nil {
		return nil, err
	}

	return p, nil
}

func (s *Service) StoreInSolana(ctx context.Context, id string, solanaAddress string) (*payment.Payment, error) {
	p, err := s.repo.FindByPaymentID(ctx, id)
	if err != nil {
		return nil, err
	}

	// if p.Status != "completed" {
	// 	return nil, errors.New("can only store completed payments in Solana")
	// }

	if p.SolanaAddress == "" {
		p.SolanaAddress = solanaAddress
	}
	tx, err := s.solanaClient.StorePayment(ctx, p)
	// Set Solana address
	p.Metadata["encodedTX"] = tx
	// Store in Solana
	// signature, err := s.solanaClient.StorePayment(ctx, p)
	if err != nil {
		return nil, err
	}

	// Update payment with Solana signature
	// p.SolanaSignature = signature
	p.UpdatedAt = time.Now()

	if err := s.repo.Save(ctx, p); err != nil {
		return nil, err
	}

	return p, nil
}
func (s *Service) InitSolanaStorage(ctx context.Context, publicKey, description string) (string, error) {
	// Initialize Solana storage
	// return s.solanaClient.InitSolanaStorage(ctx, publicKey, description)
	return s.solanaClient.InitSolanaStorage(ctx, publicKey, description)

	// return "", errors.New("not implemented")
}

func (s *Service) VerifySolanaTransaction(ctx context.Context, signature string) (bool, *payment.Payment, time.Time, error) {
	// Verify transaction on Solana
	// verified, timestamp, err := s.solanaClient.VerifyTransaction(ctx, signature)
	// if err != nil {
	// 	return false, nil, time.Time{}, err
	// }

	// Find payment with this signature
	payments, _, err := s.repo.ListPayments(ctx, "", time.Time{}, time.Time{}, 1, 10)
	if err != nil {
		return false, nil, time.Time{}, err
	}

	var p *payment.Payment
	for _, payment := range payments {
		if payment.SolanaSignature == signature {
			p = payment
			break
		}
	}

	// if p == nil {
	// 	return verified, nil, timestamp, nil
	// }

	// return verified, p, timestamp, nil

	return false, p, time.Time{}, errors.New("not implemented")
}

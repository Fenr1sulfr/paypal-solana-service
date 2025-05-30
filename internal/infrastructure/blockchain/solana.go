package solanats

import (
	"context"
	"fmt"
	"time"

	"payment.Diploma.service/internal/config"
	"payment.Diploma.service/internal/domain/payment"
	"payment.Diploma.service/internal/infrastructure/blockchain/http_client"
)

type SolanaService struct {
	Client *http_client.Client
}

func New(cfg config.Config) *SolanaService {
	return &SolanaService{http_client.NewClient(cfg)}
}

func (s *SolanaService) StorePayment(ctx context.Context, payment *payment.Payment) (string, error) {

	req := &http_client.AddPaymentRequest{
		PublicKey:        payment.SolanaAddress,
		PaymentId:        payment.PaymentID,
		PaymentOwner:     payment.PayerEmail,
		PaymentSender:    payment.PayerID,
		PaymentTimestamp: payment.CreatedAt.Unix(),
		PaymentAmount:    int64(payment.Amount * 100), // Assuming amount is in dollars, convert to cents
		PaymentCurrency:  []string{payment.Currency},
		PaymentUrl:       payment.Description, // Assuming description contains the URL or relevant info
	}
	resp, err := s.Client.ProxyAddPayment(req)
	if err != nil {
		return "", err
	}

	return resp.Transaction, nil
}

func (s *SolanaService) VerifyTransaction(context.Context, string) (bool, time.Time, error) {
	// This method should implement the logic to verify a transaction on Solana.
	// For now, we return false and zero time as a placeholder.
	return false, time.Time{}, nil
}

func (s *SolanaService) InitSolanaStorage(ctx context.Context, publicKey, description string) (string, error) {
	req := &http_client.InitializeStorageRequest{
		PublicKey: publicKey,
	}
	resp, err := s.Client.ProxyInitStorage(req)

	if err != nil {
		return "", err
	}

	if !resp.Success {
		return "", fmt.Errorf("failed to initialize Solana storage: %s", resp.Transaction)
	}

	return resp.Transaction, nil
}

package handlers

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"payment.Diploma.service/internal/domain/payment"
	pb "payment.Diploma.service/internal/ports/grpc/pb/proto"

	paymentService "payment.Diploma.service/internal/services/payment"
)

type PaymentHandler struct {
	pb.UnimplementedPaymentServiceServer
	service *paymentService.Service
}

func NewPaymentHandler(service *paymentService.Service) *PaymentHandler {
	return &PaymentHandler{
		service: service,
	}
}

func (h *PaymentHandler) CreatePayment(ctx context.Context, req *pb.CreatePaymentRequest) (*pb.PaymentResponse, error) {
	// Convert metadata
	metadata := make(map[string]string)
	for k, v := range req.Metadata {
		metadata[k] = v
	}

	// Create payment
	p, approvalURL, err := h.service.CreatePayment(
		ctx,
		req.Amount,
		req.Currency,
		req.Description,
		metadata,
		req.AutoStoreInSolana,
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create payment: %v", err)
	}

	return toPaymentResponse(p, approvalURL), nil
}

func (h *PaymentHandler) ExecutePayment(ctx context.Context, req *pb.ExecutePaymentRequest) (*pb.PaymentResponse, error) {
	p, err := h.service.ExecutePayment(ctx, req.PaymentId, req.PayerId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to execute payment: %v", err)
	}

	return toPaymentResponse(p, ""), nil
}

func (h *PaymentHandler) GetPayment(ctx context.Context, req *pb.GetPaymentRequest) (*pb.PaymentResponse, error) {
	var p *payment.Payment
	var err error

	p, err = h.service.GetPayment(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "payment not found: %v", err)
	}

	return toPaymentResponse(p, ""), nil
}

func (h *PaymentHandler) ListPayments(ctx context.Context, req *pb.ListPaymentsRequest) (*pb.ListPaymentsResponse, error) {
	var startDate, endDate time.Time
	if req.StartDate != nil {
		startDate = req.StartDate.AsTime()
	}
	if req.EndDate != nil {
		endDate = req.EndDate.AsTime()
	}

	page := int(req.Page)
	if page < 1 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	payments, total, err := h.service.ListPayments(ctx, req.Status, startDate, endDate, page, pageSize)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list payments: %v", err)
	}

	resp := &pb.ListPaymentsResponse{
		TotalCount: int32(total),
		Page:       int32(page),
		PageSize:   int32(pageSize),
	}

	for _, p := range payments {
		resp.Payments = append(resp.Payments, toPaymentResponse(p, ""))
	}

	return resp, nil
}

func (h *PaymentHandler) CancelPayment(ctx context.Context, req *pb.CancelPaymentRequest) (*pb.PaymentResponse, error) {
	p, err := h.service.CancelPayment(ctx, req.PaymentId, req.Reason)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to cancel payment: %v", err)
	}

	return toPaymentResponse(p, ""), nil
}

func (h *PaymentHandler) StoreInSolana(ctx context.Context, req *pb.StoreInSolanaRequest) (*pb.StoreInSolanaResponse, error) {
	log.Printf("StoreInSolana called: payment_id=%s, solana_address=%s", req.PaymentId, req.SolanaAddress)

	p, err := h.service.StoreInSolana(ctx, req.PaymentId, req.SolanaAddress)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to store payment in Solana: %v", err)
	}

	// Explorer URL
	explorerURL := "https://explorer.solana.com/tx/" + p.SolanaSignature
	if p.SolanaSignature == "" || !isValidSignature(p.SolanaSignature) {
		explorerURL = ""	
	}

	return &pb.StoreInSolanaResponse{
		PaymentId:       p.PaymentID,
		SolanaSignature: p.SolanaSignature,
		SolanaAddress:   p.SolanaAddress,
		Success:         p.SolanaSignature != "",
		TransactionUrl:  explorerURL,
	}, nil
}

func (h *PaymentHandler) InitSolanaStorage(ctx context.Context, req *pb.InitSolanaStorageRequest) (*pb.InitSolanaStorageResponse, error) {
	// Initialize Solana storage
	// In a real implementation, this would set up the necessary accounts or state on Solana
	// Here we just return a success response
	p, err := h.service.InitSolanaStorage(ctx, req.SolanaAddress, "")
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to store payment in Solana: %v", err)
	}
	return &pb.InitSolanaStorageResponse{
		Success:       true,
		SolanaAddress: p,
		Message:       "Solana storage initialized successfully",
	}, nil
}

func (h *PaymentHandler) VerifySolanaTransaction(ctx context.Context, req *pb.VerifySolanaRequest) (*pb.VerifySolanaResponse, error) {
	verified, p, timestamp, err := h.service.VerifySolanaTransaction(ctx, req.SolanaSignature)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to verify Solana transaction: %v", err)
	}

	resp := &pb.VerifySolanaResponse{
		Verified:        verified,
		TransactionTime: timestamppb.New(timestamp),
	}

	if p != nil {
		resp.PaymentId = p.PaymentID
	}

	return resp, nil
}

func (h *PaymentHandler) HandlePayPalWebhook(ctx context.Context, req *pb.PayPalWebhookRequest) (*emptypb.Empty, error) {
	// Process webhook based on event type
	// In a real implementation, you'd validate the webhook signature,
	// process different event types, and update payment records

	return &emptypb.Empty{}, nil
}

// Helper function to convert domain payment to gRPC response
func toPaymentResponse(p *payment.Payment, approvalURL string) *pb.PaymentResponse {
	if p == nil {
		return nil
	}

	resp := &pb.PaymentResponse{
		Id:              p.ID,
		PaymentId:       p.PaymentID,
		Amount:          p.Amount,
		Currency:        p.Currency,
		Status:          p.Status,
		Description:     p.Description,
		PaypalUrl:       approvalURL,
		PayerEmail:      p.PayerEmail,
		PayerId:         p.PayerID,
		MerchantId:      p.MerchantID,
		SolanaSignature: p.SolanaSignature,
		SolanaAddress:   p.SolanaAddress,
		StoredInSolana:  p.SolanaSignature != "",
		CreatedAt:       timestamppb.New(p.CreatedAt),
		UpdatedAt:       timestamppb.New(p.UpdatedAt),
	}

	if !p.CompletedAt.IsZero() {
		resp.CompletedAt = timestamppb.New(p.CompletedAt)
	}

	// Add Solana explorer URL if we have a signature
	if p.SolanaSignature != "" && isValidSignature(p.SolanaSignature) {
		resp.SolanaTransactionUrl = "https://explorer.solana.com/tx/" + p.SolanaSignature
	}

	// Convert metadata
	resp.Metadata = make(map[string]string)
	for k, v := range p.Metadata {
		resp.Metadata[k] = v
	}

	return resp
}

// Helper to validate a Solana signature format (basic check)
func isValidSignature(signature string) bool {
	// In a real implementation, validate the signature format
	return len(signature) > 10 && !isSimulated(signature)
}

// Check if this is a simulated signature
func isSimulated(signature string) bool {
	return len(signature) > 10 && signature[:10] == "simulated_"
}

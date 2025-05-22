package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pb "payment.Diploma.service/internal/ports/grpc/pb/proto"
)

func main() {
	// Connect to the server
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// Create a client
	client := pb.NewPaymentServiceClient(conn)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// Create a payment
	createResp, err := client.CreatePayment(ctx, &pb.CreatePaymentRequest{
		Amount:      99.99,
		Currency:    "USD",
		Description: "Test payment",
		ReturnUrl:   "http://localhost:8080/success",
		CancelUrl:   "http://localhost:8080/cancel",
		Metadata: map[string]string{
			"customer_id": "123456",
			"order_id":    "ORD-123",
		},
	})
	if err != nil {
		log.Fatalf("Failed to create payment: %v", err)
	}

	fmt.Printf("Payment created successfully\n")
	fmt.Printf("ID: %s\n", createResp.Id)
	fmt.Printf("Payment ID: %s\n", createResp.PaymentId)
	fmt.Printf("Approval URL: %s\n", createResp.PaypalUrl)
	fmt.Println("-----------------------------------")

	// Get the payment
	getResp, err := client.GetPayment(ctx, &pb.GetPaymentRequest{
		Id: createResp.Id,
	})
	if err != nil {
		log.Fatalf("Failed to get payment: %v", err)
	}

	fmt.Printf("Payment details retrieved\n")
	fmt.Printf("Status: %s\n", getResp.Status)
	fmt.Printf("Amount: %.2f %s\n", getResp.Amount, getResp.Currency)
	fmt.Printf("Created at: %s\n", getResp.CreatedAt.AsTime().Format(time.RFC3339))
	fmt.Println("-----------------------------------")

	// List payments
	listResp, err := client.ListPayments(ctx, &pb.ListPaymentsRequest{
		Page:     1,
		PageSize: 10,
	})
	if err != nil {
		log.Fatalf("Failed to list payments: %v", err)
	}

	fmt.Printf("Found %d payments (total: %d)\n", len(listResp.Payments), listResp.TotalCount)
	for i, p := range listResp.Payments {
		fmt.Printf("%d. %s - %s - %.2f %s\n", i+1, p.PaymentId, p.Status, p.Amount, p.Currency)
	}
}

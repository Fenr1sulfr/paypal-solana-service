package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"payment.Diploma.service/internal/config"
	solana_payment "payment.Diploma.service/internal/infrastructure/blockchain"
	"payment.Diploma.service/internal/infrastructure/database"
	"payment.Diploma.service/internal/infrastructure/paypal"

	grpcServer "payment.Diploma.service/internal/ports/grpc"
	paymentService "payment.Diploma.service/internal/services/payment"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Set up context with cancellation for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up MongoDB
	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoDB.URI))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer func() {
		if err := mongoClient.Disconnect(ctx); err != nil {
			log.Printf("Failed to disconnect from MongoDB: %v", err)
		}
	}()

	// Ping MongoDB to verify connection
	if err := mongoClient.Ping(ctx, nil); err != nil {
		log.Fatalf("Failed to ping MongoDB: %v", err)
	}
	log.Println("Connected to MongoDB")

	// Initialize database
	db := mongoClient.Database(cfg.MongoDB.Database)
	paymentRepo := database.NewPaymentRepository(db)

	// // Initialize Solana client
	// solanaClient := blockchain.NewSolanaClient(
	// 	cfg.Solana.Endpoint,
	// 	cfg.Solana.PrivateKey,
	// 	// cfg.Solana.ProgramID,
	// )
	// if err != nil {
	// 	log.Fatalf("Failed to initialize Solana client: %v", err)
	// }
	solanaClient := solana_payment.New(*cfg)

	// Initialize PayPal client
	paypalClient := paypal.NewPayPalClient(
		cfg.PayPal.ClientID,
		cfg.PayPal.ClientSecret,
		cfg.PayPal.Sandbox,
	)

	// Initialize payment service
	service := paymentService.NewService(
		paymentRepo,
		solanaClient,
		paypalClient,
		cfg.Server.BaseURL,
	)

	// Set up gRPC server
	server, err := grpcServer.NewServer(cfg.Server.Port, service)
	if err != nil {
		log.Fatalf("Failed to create gRPC server: %v", err)
	}

	// Start gRPC server
	log.Printf("Starting gRPC server on port %d", cfg.Server.Port)
	go func() {
		if err := server.Start(); err != nil {
			log.Fatalf("Failed to start gRPC server: %v", err)
		}
	}()

	// Handle graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	server.Stop()
	log.Println("Server stopped")
}

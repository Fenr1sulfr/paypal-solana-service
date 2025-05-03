package grpc

import (
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"payment.Diploma.service/internal/ports/grpc/handlers"
	pb "payment.Diploma.service/internal/ports/grpc/pb/proto"
	paymentService "payment.Diploma.service/internal/services/payment"
)

type Server struct {
	server   *grpc.Server
	listener net.Listener
	port     int
}

func NewServer(port int, paymentService *paymentService.Service) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, fmt.Errorf("failed to listen: %v", err)
	}

	server := grpc.NewServer()

	// Register payment service handler
	paymentHandler := handlers.NewPaymentHandler(paymentService)
	pb.RegisterPaymentServiceServer(server, paymentHandler)

	// Register reflection service for tools like grpcurl
	reflection.Register(server)

	return &Server{
		server:   server,
		listener: listener,
		port:     port,
	}, nil
}

func (s *Server) Start() error {
	return s.server.Serve(s.listener)
}

func (s *Server) Stop() {
	s.server.GracefulStop()
}

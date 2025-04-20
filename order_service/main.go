package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	pb "github.com/DinukaKaveen/Golang-gRPC-Microservices/proto/order/generated"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// JWT secret (in production, use environment variables or a secret manager)
const jwtSecret = "my-secret-key"

// orderServer implements the OrderService gRPC server
type orderServer struct {
	pb.UnimplementedOrderServiceServer
}

// JWTInterceptor validates JWT tokens in gRPC requests
func JWTInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "Missing metadata")
	}

	authHeader := md.Get("authorization")
	if len(authHeader) == 0 {
		return nil, status.Errorf(codes.Unauthenticated, "Missing authorization header")
	}
	
	parts := strings.Split(authHeader[0], " ")
	tokenStr := parts[1]

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Method)
		}
		return []byte(jwtSecret), nil
	})

	if err != nil || !token.Valid {
		return nil, status.Errorf(codes.Unauthenticated, "Invalid JWT token: %v", err)
	}

	return handler(ctx, req)
}

// CreateOrder handles the CreateOrder gRPC request
func (s *orderServer) CreateOrder(ctx context.Context, req *pb.OrderRequest) (*pb.OrderResponse, error) {
	orderId := uuid.New().String()
	// Simulate creating order to the database
	return &pb.OrderResponse{
		OrderId: orderId,
		Status: "CREATED",
	}, nil
}

// GetOrder handles the GetOrder gRPC request
func (s *orderServer) GetOrder(ctx context.Context, req *pb.GetOrderRequest) (*pb.GetOrderResponse, error) {
	// Simulate fetching order from the database
	return &pb.GetOrderResponse{
		OrderId: req.OrderId,
		UserId:  "user123",
		Status:  "pending",
	}, nil
}

func main() {
	// Load server certificate and key
	cert, err := tls.LoadX509KeyPair("certs/server.crt", "certs/server.key")
	if err != nil {
		log.Fatalf("Failed to load server certificate: %v", err)
	}
	// Load CA certificate
	caCert, err := os.ReadFile("certs/ca.crt")
	if err != nil {
		log.Fatalf("Failed to read CA cert: %v", err)
	}
	caPool := x509.NewCertPool()
	caPool.AppendCertsFromPEM(caCert)

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    caPool,
	}
	creds := credentials.NewTLS(tlsConfig)

	// Start gRPC server
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// Apply TLS and JWT to gRPC server
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(JWTInterceptor),
		grpc.Creds(creds),
	)
	pb.RegisterOrderServiceServer(grpcServer, &orderServer{})

	go func() {
		log.Printf("gRPC server listening on :50051")
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()

	// Start Fiber app
	app := fiber.New()

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("Order service is healthy")
	})

	log.Fatal(app.Listen(":3001"))
}
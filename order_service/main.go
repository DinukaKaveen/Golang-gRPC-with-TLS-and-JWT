package main

import (
	"context"
	"log"
	"net"

	"github.com/gofiber/fiber/v2"
	pb "github.com/DinukaKaveen/Golang-gRPC-Microservices/proto/order/generated"
	"github.com/google/uuid"
	"google.golang.org/grpc"
)

// orderServer implements the OrderService gRPC server
type orderServer struct {
	pb.UnimplementedOrderServiceServer
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
		UserId:  "user123", // Hardcoded for demo
		Status:  "pending", // Hardcoded for demo
	}, nil
}

func main() {
	// Start gRPC server
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
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
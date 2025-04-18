package main

import (
	"crypto/tls"
	"crypto/x509"
	"log"
	"os"
	"time"

	pb "github.com/DinukaKaveen/Golang-gRPC-Microservices/proto/order/generated"
	"github.com/golang-jwt/jwt/v5"

	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Generates JWT token for authentication (in production, use environment variables or a secret manager)
const jwtSecret = "my-secret-key"
func GenerateJWT() (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"service": "user-service",
		"exp":     time.Now().Add(time.Hour * 1).Unix(),
	})
	return token.SignedString([]byte(jwtSecret))
}

func main() {
	// Load client TLS credentials
	caCert, err := os.ReadFile("certs/ca.crt")
	if err != nil {
		log.Fatalf("Failed to read CA cert: %v", err)
	}
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(caCert) {
		log.Fatalf("Failed to append CA cert")
	}
	// Create TLS credentials
	tlsConfig := &tls.Config {
		RootCAs: certPool,
	}
	creds := credentials.NewTLS(tlsConfig)

	// Create new gRPC client to connect to Order Service gRPC
	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(creds))
	if err != nil {
		log.Fatalf("Failed to connect to order service: %v", err)
	}

	// Dont forget to close it
	defer conn.Close()

	// Create a new order service client from generated code and pass in the connection created above
	orderClient := pb.NewOrderServiceClient(conn)

	// Initialize Fiber app
	app := fiber.New()

	// Create Order REST endpoint
	app.Post("/users/:id/order", func(c *fiber.Ctx) error {
		userId := c.Params("id")

		// Generate JWT token and Add to gRPC metadata
		token, err := GenerateJWT()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate JWT"})
		}
		ctx := metadata.AppendToOutgoingContext(c.Context(), "authorization", "Bearer "+token)

		// Call gRPC CreateOrder method
		resp, err := orderClient.CreateOrder(ctx, &pb.OrderRequest{
			UserId: userId,
			Amount: 100.00,
		})

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(fiber.Map{
			"order_id": resp.OrderId,
			"status":   resp.Status,
		})
	})

	// Get Order REST endpoint
	app.Get("/user/:userId/orders/:orderId", func (c *fiber.Ctx) error {
		userId := c.Params("userId")
		orderId := c.Params("orderId")

		// Generate JWT token and Add to gRPC metadata
		token, err := GenerateJWT()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate JWT"})
		}
		ctx := metadata.AppendToOutgoingContext(c.Context(), "authorization", "Bearer "+token)

		// Call gRPC GetOrder method
		resp, err := orderClient.GetOrder(ctx, &pb.GetOrderRequest{
			OrderId: orderId,
		})

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		// Verify user ID matches
		if resp.UserId != userId {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Order does not belong to user"})
		}

		return c.JSON(fiber.Map{
			"order_id": resp.OrderId,
			"user_id":  resp.UserId,
			"status":   resp.Status,
		})
	})

	log.Fatal(app.Listen(":3000"))
}
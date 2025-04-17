package main

import (
	"log"

	pb "github.com/DinukaKaveen/Golang-gRPC-Microservices/proto/order/generated"

	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func main() {
	// Create new gRPC client to connect to Order Service gRPC
	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
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

		// Call gRPC CreateOrder method
		resp, err := orderClient.CreateOrder(c.Context(), &pb.OrderRequest{
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

		// Call gRPC GetOrder method
		resp, err := orderClient.GetOrder(c.Context(), &pb.GetOrderRequest{
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
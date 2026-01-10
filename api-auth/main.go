package main

import (
	"fmt"
	"net"
	//"time"
	"log"
	"github.com/joho/godotenv"
	"github.com/gin-gonic/gin"
	"github.com/burkeclove/auth-api/services"
	"github.com/burkeclove/auth-api/middleware"
	//"github.com/jackc/pgx/v5/pgxpool"
	//"context"
	"github.com/burkeclove/shared/config"
	pb "github.com/burkeclove/shared/gen/go/protos"
	//"github.com/burkeclove/shared/db/sqlc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	httpPort = 8080
	grpcPort = 50051
)

func main() {
	// get env to start
	if err := godotenv.Load(); err != nil {
        log.Println("An error occurred while loading env file:", err)
        panic(err)
    }		

	config := config.Load()
	q := config.CreatePool()

	// Create a listener on the specified port
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// Create a new gRPC server
	s := grpc.NewServer()

	// Register our service
	auth_service := services.NewAuthService(q)
	pb.RegisterAuthServiceServer(s, auth_service)

	// Register reflection service for easier debugging
	reflection.Register(s)

	// Start gRPC server in a goroutine
	go func() {
		log.Printf("Starting gRPC server on :%d", grpcPort)
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()

	// create auth middleware
	r := gin.Default()
	auth := r.Group("/api/auth")
	auth.POST("/login", auth_service.Login)
	auth.Use(middleware.PortalMiddleware(q, auth_service.JwtService.Validate))
	{
		auth.GET("/", auth_service.GetApiKeys)
		auth.POST("/", auth_service.CreateApiKey)
	}

	log.Printf("Starting HTTP server on :%d", httpPort)
    r.Run(":8080")
}

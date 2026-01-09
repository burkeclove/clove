package main

import (
	"os"
	"log"
	"github.com/joho/godotenv"
	"github.com/gin-gonic/gin"
	"github.com/burkeclove/shared/config"
	pb "github.com/burkeclove/shared/gen/go/protos"
	"github.com/burkeclove/users-api/services"
	"github.com/burkeclove/shared/middleware"
	"google.golang.org/grpc"
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
	log.Println("Setting up auth connection")
    auth_conn, err := grpc.Dial(os.Getenv("AUTH_CONNECTION"), grpc.WithInsecure(), grpc.WithBlock())
    if err != nil {
        log.Println("An error occurred while setting up auth grpc connection", err)
        panic(err) 
    }
    defer auth_conn.Close()
	auth_grpc_client := pb.NewAuthServiceClient(auth_conn) 	
    users_service := services.NewUserService(q, auth_grpc_client) 	

	// create auth middleware
	r := gin.Default()
	auth := r.Group("/api/users")	
	
	// no auth on create user
	auth.POST("/", users_service.CreateUser)

	auth.Use(middleware.PortalMiddleware(q, auth_grpc_client))
	{
		auth.GET("/", users_service.GetUser)
	}

	log.Println("About to serve on :8080")
    r.Run(":8080")
}

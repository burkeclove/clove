package main

import (
	"os"
	"log"
	"github.com/joho/godotenv"
	"github.com/gin-gonic/gin"
	"github.com/burkeclove/shared/config"
	pb "github.com/burkeclove/shared/gen/go/protos"
	"github.com/burkeclove/organizations-api/services"
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
    organization_service := services.NewOrganizationService(q, auth_grpc_client) 	

	// create auth middleware
	r := gin.Default()
	auth := r.Group("/api/organizations")	
	auth.Use(middleware.PortalMiddleware(q, auth_grpc_client))
	{
		auth.GET("/:orgId", organization_service.GetOrganizationById)
		auth.POST("/", organization_service.CreateOrganization)
	}

	log.Println("About to serve on :8080")
    r.Run(":8080")
}

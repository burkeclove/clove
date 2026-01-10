package main

import (
	"log"
	"os"

	"github.com/burkeclove/minio-api/helpers"
	localMiddleware "github.com/burkeclove/minio-api/middleware"
	"github.com/burkeclove/minio-api/services"
	"github.com/burkeclove/shared/config"
	pb "github.com/burkeclove/shared/gen/go/protos"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

var (
	httpPort = 8080
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("An error occurred while loading env file:", err)
		panic(err)
	}

	// Database and config setup
	config := config.Load()
	q := config.CreatePool()

	// Setup gRPC connection to auth service
	log.Println("Setting up auth connection")
	auth_conn, err := grpc.Dial(os.Getenv("AUTH_CONNECTION"), grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Println("An error occurred while setting up auth grpc connection", err)
		panic(err)
	}
	defer auth_conn.Close()
	auth_grpc_client := pb.NewAuthServiceClient(auth_conn)

	// Create MinIO client
	minio_client := services.NewMinioClient(auth_grpc_client)

	r := gin.Default()
	api := r.Group("/api")
	api.Use(localMiddleware.SigV4Middleware(q, auth_grpc_client))
	{
		// List all buckets - requires s3:ListAllMyBuckets on *
		api.GET("/buckets",
			localMiddleware.RequireAction("s3:ListAllMyBuckets", func(c *gin.Context) string {
				return helpers.AllBucketsARN()
			}),
			minio_client.ListBuckets,
		)

		// Create bucket - requires s3:CreateBucket on arn:clove:s3:::bucket-name
		api.PUT("/bucket/:bucket",
			localMiddleware.RequireAction("s3:CreateBucket", func(c *gin.Context) string {
				return helpers.BucketARN(c.Param("bucket"))
			}),
			minio_client.CreateBucketHandler,
		)

		// Delete bucket - requires s3:DeleteBucket on arn:clove:s3:::bucket-name
		api.DELETE("/bucket/:bucket",
			localMiddleware.RequireAction("s3:DeleteBucket", func(c *gin.Context) string {
				return helpers.BucketARN(c.Param("bucket"))
			}),
			minio_client.DeleteBucket,
		)

		// List objects in bucket - requires s3:ListBucket on arn:clove:s3:::bucket-name
		api.GET("/bucket/:bucket",
			localMiddleware.RequireAction("s3:ListBucket", func(c *gin.Context) string {
				return helpers.BucketARN(c.Param("bucket"))
			}),
			minio_client.ListObjects,
		)
	}

	// Object operations
	{
		// Get object - requires s3:GetObject on arn:clove:s3:::bucket-name/object-key
		api.GET("/bucket/:bucket/object/*object",
			localMiddleware.RequireAction("s3:GetObject", func(c *gin.Context) string {
				objectKey := c.Param("object")[1:] // Remove leading slash
				return helpers.ObjectARN(c.Param("bucket"), objectKey)
			}),
			minio_client.GetObject,
		)

		// Put object - requires s3:PutObject on arn:clove:s3:::bucket-name/object-key
		api.PUT("/bucket/:bucket/object/*object",
			localMiddleware.RequireAction("s3:PutObject", func(c *gin.Context) string {
				objectKey := c.Param("object")[1:] // Remove leading slash
				return helpers.ObjectARN(c.Param("bucket"), objectKey)
			}),
			minio_client.PutObject,
		)

		// Delete object - requires s3:DeleteObject on arn:clove:s3:::bucket-name/object-key
		api.DELETE("/bucket/:bucket/object/*object",
			localMiddleware.RequireAction("s3:DeleteObject", func(c *gin.Context) string {
				objectKey := c.Param("object")[1:] // Remove leading slash
				return helpers.ObjectARN(c.Param("bucket"), objectKey)
			}),
			minio_client.DeleteObject,
		)
	}

	log.Printf("Starting MinIO API server on :%d", httpPort)
	r.Run(":8080")
}

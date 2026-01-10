package services

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/burkeclove/shared/db/helpers"
	"github.com/burkeclove/shared/db/sqlc"
	pb "github.com/burkeclove/shared/gen/go/protos"
	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioClient struct {
	Client *minio.Client	
	AuthClient pb.AuthServiceClient
	Q *sqlc.Queries
}

func NewMinioClient(auth_conn pb.AuthServiceClient, q *sqlc.Queries) *MinioClient {
	endpoint := getEnv("MINIO_ENDPOINT", "localhost:9000")
	accessKeyID := getEnv("MINIO_ACCESS_KEY", "minioadmin")
	secretAccessKey := getEnv("MINIO_SECRET_KEY", "minioadmin")
	useSSL := getEnv("MINIO_USE_SSL", "false") == "true"

	// Initialize minio client object.
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		log.Fatalln(err)
		return nil
	}
	return &MinioClient{Client: minioClient, AuthClient: auth_conn, Q: q}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func (m *MinioClient) CreateBucketWithCheck(bucketName string) {
	location := "us-east-1"
	err := m.Client.MakeBucket(context.Background(), bucketName, minio.MakeBucketOptions{Region: location})
	if err != nil {
		exists, errBucketExists := m.Client.BucketExists(context.Background(), bucketName)
		if errBucketExists == nil && exists {
			fmt.Println("Bucket already exists")
		} else {
			log.Fatal(err)
		}
	}
	fmt.Println("Bucket created successfully!")
}

func (m *MinioClient) CreateBucket() {
	err := m.Client.MakeBucket(context.Background(), "shards", minio.MakeBucketOptions{Region: "us-east-1", ObjectLocking: true})
	if err != nil {
		fmt.Println("error while creating a bucket: ", err)
		return
	}
	fmt.Println("Successfully created mybucket.")
}

func (m *MinioClient) PutBytes(data []byte, name string) error {
	reader := bytes.NewReader(data)
	_, err := m.Client.PutObject(
		context.Background(),
		"my-bucket",
		name,
		reader,
		int64(len(data)),
		minio.PutObjectOptions{
			ContentType: "application/octet-stream",
		},
	)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

// ListBuckets lists all buckets (requires s3:ListAllMyBuckets permission)
func (m *MinioClient) ListBuckets(c *gin.Context) {
	orgId, _ := c.Get("org_id")
	log.Printf("Listing buckets for org: %s", orgId)

	buckets, err := m.Client.ListBuckets(c.Request.Context())
	if err != nil {
		log.Printf("Error listing buckets: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list buckets"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"buckets": buckets})
}

// ListObjects lists objects in a bucket (requires s3:ListBucket permission)
func (m *MinioClient) ListObjects(c *gin.Context) {
	bucketName := c.Param("bucket")
	prefix := c.Query("prefix")
	orgId, _ := c.Get("org_id")
	log.Printf("Listing objects in bucket %s for org: %s", bucketName, orgId)

	objectCh := m.Client.ListObjects(c.Request.Context(), bucketName, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	})

	objects := []minio.ObjectInfo{}
	for object := range objectCh {
		if object.Err != nil {
			log.Printf("Error listing objects: %v", object.Err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list objects"})
			return
		}
		objects = append(objects, object)
	}

	c.JSON(http.StatusOK, gin.H{"objects": objects})
}

// GetObject downloads an object (requires s3:GetObject permission)
func (m *MinioClient) GetObject(c *gin.Context) {
	bucketName := c.Param("bucket")
	objectName := c.Param("object")
	orgId, _ := c.Get("org_id")
	log.Printf("Getting object %s/%s for org: %s", bucketName, objectName, orgId)

	object, err := m.Client.GetObject(c.Request.Context(), bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		log.Printf("Error getting object: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get object"})
		return
	}
	defer object.Close()

	stat, err := object.Stat()
	if err != nil {
		log.Printf("Error getting object stat: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get object"})
		return
	}

	c.Header("Content-Type", stat.ContentType)
	c.Header("Content-Length", fmt.Sprintf("%d", stat.Size))
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", objectName))

	if _, err := io.Copy(c.Writer, object); err != nil {
		log.Printf("Error streaming object: %v", err)
		return
	}
}

// PutObject uploads an object (requires s3:PutObject permission)
func (m *MinioClient) PutObject(c *gin.Context) {
	bucketName := c.Param("bucket")
	objectName := c.Param("object")
	orgId, _ := c.Get("org_id")
	log.Printf("Putting object %s/%s for org: %s", bucketName, objectName, orgId)

	contentType := c.GetHeader("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	info, err := m.Client.PutObject(
		c.Request.Context(),
		bucketName,
		objectName,
		c.Request.Body,
		c.Request.ContentLength,
		minio.PutObjectOptions{
			ContentType: contentType,
		},
	)
	if err != nil {
		log.Printf("Error putting object: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload object"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"bucket": info.Bucket,
		"key":    info.Key,
		"etag":   info.ETag,
		"size":   info.Size,
	})
}

// DeleteObject deletes an object (requires s3:DeleteObject permission)
func (m *MinioClient) DeleteObject(c *gin.Context) {
	bucketName := c.Param("bucket")
	objectName := c.Param("object")
	orgId, _ := c.Get("org_id")
	log.Printf("Deleting object %s/%s for org: %s", bucketName, objectName, orgId)

	err := m.Client.RemoveObject(c.Request.Context(), bucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		log.Printf("Error deleting object: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete object"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Object deleted successfully"})
}

// CreateBucket creates a new bucket (requires s3:CreateBucket permission)
func (m *MinioClient) CreateBucketHandler(c *gin.Context) {
	reqBucketName := c.Param("bucket")
	orgId, _ := c.Get("org_id")
	log.Printf("Creating bucket %s for org: %s", reqBucketName, orgId)

	// create bucket in db first
	orgUUID, err := helpers.UUIDFromString(orgId.(string))
	if err != nil {
		log.Println("an error occured while getting uuid from string for org id: ", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "an error occured while getting uuid from string for org id"})
	}
	res, err := m.Q.CreateS3Bucket(c.Request.Context(), sqlc.CreateS3BucketParams{
		Name: reqBucketName,
		OrganizationID: orgUUID,
	})
	bucketName := m.GetBucketName(orgId.(string), res.ID.String(), reqBucketName)

	err = m.Client.MakeBucket(c.Request.Context(), reqBucketName, minio.MakeBucketOptions{
		Region: "us-east-1",
	})
	if err != nil {
		log.Printf("Error creating bucket: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create bucket"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"bucket":  bucketName,
		"message": "Bucket created successfully",
	})
}

// DeleteBucket deletes a bucket (requires s3:DeleteBucket permission)
func (m *MinioClient) DeleteBucket(c *gin.Context) {
	reqBucketName := c.Param("bucket")
	orgId, _ := c.Get("org_id")
	log.Printf("Deleting bucket %s for org: %s", reqBucketName, orgId)

	orgUUID, err := helpers.UUIDFromString(orgId.(string))
	bucket, err := m.Q.DeleteS3BucketByName(c.Request.Context(), sqlc.DeleteS3BucketByNameParams{
		Name: reqBucketName,
		OrganizationID: orgUUID,
	})
	bucketName := m.GetBucketName(orgId.(string), bucket.ID.String(), reqBucketName)

	err = m.Client.RemoveBucket(c.Request.Context(), bucketName)
	if err != nil {
		log.Printf("Error deleting bucket: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete bucket"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Bucket deleted successfully"})
}

func (m *MinioClient) GetBucketName(orgId, bucketId, bucketName string) string {
	return fmt.Sprintf("%s-%s-%s", orgId, bucketId, bucketName)
}

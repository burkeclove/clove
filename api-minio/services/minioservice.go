package services

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/gin-gonic/gin"
)

type MinioClient struct {
	Client *minio.Client	
}

func NewMinioClient() *MinioClient {
	endpoint := "localhost:9000"
	accessKeyID := "minioadmin"
	secretAccessKey := "minioadmin"
	useSSL := false

	// Initialize minio client object.
 	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		log.Fatalln(err)
		return nil
	}
	return &MinioClient{Client: minioClient}
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

func (m *MinioClient) CreateSigV4CredentialsGin(c *gin.Context) {
	
}

func (m *MinioClient) CreateSigV4Credentials(organizationId string) {
	
}

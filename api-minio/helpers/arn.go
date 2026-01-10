package helpers

import "fmt"

// BucketARN constructs an ARN for a bucket
func BucketARN(bucketId string) string {
	return fmt.Sprintf("arn:clove:s3:::%s", bucketId)
}

// ObjectARN constructs an ARN for an object
func ObjectARN(bucketId, objectKey string) string {
	return fmt.Sprintf("arn:clove:s3:::%s/%s", bucketId, objectKey)
}

// AllBucketsARN returns the wildcard ARN for all buckets
func AllBucketsARN() string {
	return "*"
}

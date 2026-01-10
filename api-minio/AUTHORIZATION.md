# MinIO API Authorization Pattern

This document describes how to implement endpoint-level authorization using IAM policies embedded in SigV4 session tokens.

## Architecture

### 1. Middleware (Session Token Validation)
The `SigV4Middleware` validates the session token and extracts:
- `org_id`: Organization identifier
- `policy`: IAM policy JSON string

These are set in the Gin context for endpoints to use.

### 2. Policy Evaluation (Per-Endpoint)
Each endpoint checks if the action is allowed using the policy evaluator helpers.

## Usage Patterns

### Pattern 1: Manual Policy Check in Endpoint

```go
func (s *MinioService) ListBuckets(c *gin.Context) {
    // Check if the action is allowed
    allowed, err := middleware.IsActionAllowed(c, "s3:ListAllMyBuckets", "*")
    if err != nil {
        c.JSON(500, gin.H{"error": "Failed to evaluate policy"})
        return
    }
    if !allowed {
        c.JSON(403, gin.H{"error": "s3:ListAllMyBuckets not allowed"})
        return
    }

    // Action is allowed, proceed with business logic
    orgId, _ := c.Get("org_id")
    buckets := s.listBucketsForOrg(orgId.(string))
    c.JSON(200, gin.H{"buckets": buckets})
}
```

### Pattern 2: Using RequireAction Middleware

```go
// In router setup
r := gin.Default()
r.Use(middleware.SigV4Middleware(q, authClient))

// ListBuckets - requires s3:ListAllMyBuckets on *
r.GET("/buckets",
    middleware.RequireAction("s3:ListAllMyBuckets", func(c *gin.Context) string {
        return "*"
    }),
    minioService.ListBuckets,
)

// GetObject - requires s3:GetObject on specific bucket/object
r.GET("/bucket/:bucketId/object/*objectKey",
    middleware.RequireAction("s3:GetObject", func(c *gin.Context) string {
        bucketId := c.Param("bucketId")
        objectKey := c.Param("objectKey")
        return fmt.Sprintf("arn:clove:s3:::%s/%s", bucketId, objectKey)
    }),
    minioService.GetObject,
)

// ListBucket - requires s3:ListBucket on bucket
r.GET("/bucket/:bucketId",
    middleware.RequireAction("s3:ListBucket", func(c *gin.Context) string {
        bucketId := c.Param("bucketId")
        return fmt.Sprintf("arn:clove:s3:::%s", bucketId)
    }),
    minioService.ListBucket,
)

// PutObject - requires s3:PutObject on bucket/prefix
r.PUT("/bucket/:bucketId/object/*objectKey",
    middleware.RequireAction("s3:PutObject", func(c *gin.Context) string {
        bucketId := c.Param("bucketId")
        objectKey := c.Param("objectKey")
        return fmt.Sprintf("arn:clove:s3:::%s/%s", bucketId, objectKey)
    }),
    minioService.PutObject,
)
```

### Pattern 3: Helper Function for Consistent Resource ARNs

```go
// helpers.go
package helpers

import "fmt"

func BucketARN(bucketId string) string {
    return fmt.Sprintf("arn:clove:s3:::%s", bucketId)
}

func ObjectARN(bucketId, objectKey string) string {
    return fmt.Sprintf("arn:clove:s3:::%s/%s", bucketId, objectKey)
}

// In routes
r.GET("/bucket/:bucketId",
    middleware.RequireAction("s3:ListBucket", func(c *gin.Context) string {
        return helpers.BucketARN(c.Param("bucketId"))
    }),
    minioService.ListBucket,
)
```

## Common S3 Actions

### Bucket-Level Actions
- `s3:ListAllMyBuckets` - List all buckets (resource: `*`)
- `s3:ListBucket` - List objects in a bucket (resource: `arn:clove:s3:::bucket-name`)
- `s3:CreateBucket` - Create a bucket (resource: `arn:clove:s3:::bucket-name`)
- `s3:DeleteBucket` - Delete a bucket (resource: `arn:clove:s3:::bucket-name`)

### Object-Level Actions
- `s3:GetObject` - Download an object (resource: `arn:clove:s3:::bucket-name/object-key`)
- `s3:PutObject` - Upload an object (resource: `arn:clove:s3:::bucket-name/object-key`)
- `s3:DeleteObject` - Delete an object (resource: `arn:clove:s3:::bucket-name/object-key`)
- `s3:GetObjectVersion` - Get specific object version
- `s3:ListMultipartUploadParts` - List parts of multipart upload

## Example Policy Evaluation

Given this policy:
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": ["s3:ListBucket", "s3:GetObject"],
      "Resource": [
        "arn:clove:s3:::my-bucket",
        "arn:clove:s3:::my-bucket/data/*"
      ]
    }
  ]
}
```

**Allowed:**
- `IsActionAllowed(c, "s3:ListBucket", "arn:clove:s3:::my-bucket")` ✅
- `IsActionAllowed(c, "s3:GetObject", "arn:clove:s3:::my-bucket/data/file.txt")` ✅

**Denied:**
- `IsActionAllowed(c, "s3:PutObject", "arn:clove:s3:::my-bucket/data/file.txt")` ❌ (action not in policy)
- `IsActionAllowed(c, "s3:GetObject", "arn:clove:s3:::my-bucket/config/file.txt")` ❌ (prefix mismatch)
- `IsActionAllowed(c, "s3:ListAllMyBuckets", "*")` ❌ (action not in policy)

## Benefits of This Approach

1. **Scalable**: Each endpoint independently validates permissions
2. **Flexible**: Different endpoints can require different permissions
3. **Stateless**: Policy is embedded in the token, no database lookups
4. **Fine-grained**: Supports prefix-based restrictions (e.g., `imagenet/v1/*`)
5. **Standard**: Uses AWS IAM policy format (familiar to developers)
6. **Testable**: Easy to unit test policy evaluation logic

## Best Practices

1. **Always check permissions**: Never skip authorization checks
2. **Use specific actions**: Don't use wildcards in permission checks
3. **Construct ARNs consistently**: Use helper functions
4. **Log authorization failures**: Help with debugging
5. **Fail closed**: Deny by default if policy evaluation fails
6. **Consider conditions**: Can extend to check IP allowlists, object sizes, etc.

## Future Enhancements

- Support for `Deny` statements (explicit denies override allows)
- Condition evaluation (IP address, object size, time-based)
- Multi-statement policies with complex logic
- Policy caching for performance
- Audit logging of authorization decisions

# MinIO API with SigV4 Authorization

This API provides S3-compatible object storage operations with fine-grained IAM policy-based authorization.

## Architecture

```
Client Request → SigV4Middleware (validates session token)
              → RequireAction (checks IAM policy for specific permission)
              → Handler (performs MinIO operation)
```

## Available Endpoints

All endpoints require a valid session token in the `Authorization` header.

### Bucket Operations

| Method | Endpoint | Required Permission | Description |
|--------|----------|-------------------|-------------|
| GET | `/api/buckets` | `s3:ListAllMyBuckets` on `*` | List all buckets |
| PUT | `/api/bucket/:bucket` | `s3:CreateBucket` on `arn:clove:s3:::bucket` | Create a bucket |
| DELETE | `/api/bucket/:bucket` | `s3:DeleteBucket` on `arn:clove:s3:::bucket` | Delete a bucket |
| GET | `/api/bucket/:bucket` | `s3:ListBucket` on `arn:clove:s3:::bucket` | List objects in bucket |

### Object Operations

| Method | Endpoint | Required Permission | Description |
|--------|----------|-------------------|-------------|
| GET | `/api/bucket/:bucket/object/*path` | `s3:GetObject` on `arn:clove:s3:::bucket/path` | Download an object |
| PUT | `/api/bucket/:bucket/object/*path` | `s3:PutObject` on `arn:clove:s3:::bucket/path` | Upload an object |
| DELETE | `/api/bucket/:bucket/object/*path` | `s3:DeleteObject` on `arn:clove:s3:::bucket/path` | Delete an object |

## Example Usage

### 1. Generate SigV4 Credentials

First, generate temporary credentials with specific permissions:

```bash
curl -X POST http://localhost/api/auth/sigv4/credentials \
  -H "Authorization: Bearer <your-jwt>" \
  -H "Content-Type: application/json" \
  -d '{
    "org_id": "b76a086a-b642-4791-adf5-4ce2d95163ed",
    "global_actions": ["s3:ListAllMyBuckets"],
    "buckets": [
      {
        "bucket_id": "my-bucket",
        "actions": ["s3:ListBucket", "s3:GetObject", "s3:PutObject"],
        "prefixes": ["data/"]
      }
    ]
  }'
```

**Response:**
```json
{
  "access_key": "AKIAabc123...",
  "secret_key": "secretkey123...",
  "session_token": "eyJhbGciOiJSUzI1NiIs...",
  "expires_at": "2026-01-10T23:00:00Z"
}
```

### 2. Use Session Token for API Requests

Use the `session_token` in the `Authorization` header:

```bash
# List all buckets
curl -X GET http://localhost/api/buckets \
  -H "Authorization: eyJhbGciOiJSUzI1NiIs..."

# List objects in bucket (with optional prefix filter)
curl -X GET "http://localhost/api/bucket/my-bucket?prefix=data/" \
  -H "Authorization: eyJhbGciOiJSUzI1NiIs..."

# Upload an object
curl -X PUT http://localhost/api/bucket/my-bucket/object/data/file.txt \
  -H "Authorization: eyJhbGciOiJSUzI1NiIs..." \
  -H "Content-Type: text/plain" \
  --data-binary @file.txt

# Download an object
curl -X GET http://localhost/api/bucket/my-bucket/object/data/file.txt \
  -H "Authorization: eyJhbGciOiJSUzI1NiIs..." \
  -o downloaded-file.txt

# Delete an object
curl -X DELETE http://localhost/api/bucket/my-bucket/object/data/file.txt \
  -H "Authorization: eyJhbGciOiJSUzI1NiIs..."

# Create a new bucket
curl -X PUT http://localhost/api/bucket/new-bucket \
  -H "Authorization: eyJhbGciOiJSUzI1NiIs..."

# Delete a bucket
curl -X DELETE http://localhost/api/bucket/old-bucket \
  -H "Authorization: eyJhbGciOiJSUzI1NiIs..."
```

## Authorization Flow

### Example: Downloading a File

**Request:**
```
GET /api/bucket/my-bucket/object/data/train.tar
Authorization: <session-token>
```

**Step 1: SigV4Middleware**
- Validates session token JWT signature
- Extracts and sets in context:
  - `org_id`: "b76a086a-b642-4791-adf5-4ce2d95163ed"
  - `policy`: IAM policy JSON

**Step 2: RequireAction Middleware**
- Checks: `IsActionAllowed("s3:GetObject", "arn:clove:s3:::my-bucket/data/train.tar")`
- Evaluates policy statements
- Result: ✅ Allowed (matches `"arn:clove:s3:::my-bucket/data/*"`)

**Step 3: Handler**
- Downloads object from MinIO
- Streams to client

## Policy Evaluation Examples

Given this policy:
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": ["s3:ListAllMyBuckets"],
      "Resource": "*"
    },
    {
      "Effect": "Allow",
      "Action": ["s3:ListBucket", "s3:GetObject", "s3:PutObject"],
      "Resource": [
        "arn:clove:s3:::my-bucket",
        "arn:clove:s3:::my-bucket/data/*"
      ]
    }
  ]
}
```

### Allowed Operations

✅ `GET /api/buckets` - ListAllMyBuckets on *
✅ `GET /api/bucket/my-bucket` - ListBucket on my-bucket
✅ `GET /api/bucket/my-bucket/object/data/file.txt` - GetObject on my-bucket/data/*
✅ `PUT /api/bucket/my-bucket/object/data/upload.tar` - PutObject on my-bucket/data/*

### Denied Operations

❌ `GET /api/bucket/my-bucket/object/config/secrets.txt` - Wrong prefix (not under data/)
❌ `DELETE /api/bucket/my-bucket/object/data/file.txt` - Action not in policy
❌ `GET /api/bucket/other-bucket` - Different bucket
❌ `PUT /api/bucket/new-bucket` - CreateBucket not in policy

## Configuration

### Environment Variables

```env
# MinIO connection
MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin
MINIO_USE_SSL=false

# Auth service connection
AUTH_CONNECTION=api-auth:50051

# Database
DATABASE_URL=postgres://cloveuser:clovepassword@postgres:5432/clovedb?sslmode=disable
```

## Error Responses

### 401 Unauthorized
```json
{
  "error": "Invalid session token"
}
```
**Cause:** Session token is missing, expired, or has invalid signature

### 403 Forbidden
```json
{
  "error": "Action s3:GetObject not allowed on resource arn:clove:s3:::bucket/file"
}
```
**Cause:** Policy doesn't permit the requested action on the resource

### 500 Internal Server Error
```json
{
  "error": "Failed to get object"
}
```
**Cause:** MinIO operation failed (bucket doesn't exist, network error, etc.)

## Benefits

1. **Fine-grained Control**: Restrict access to specific buckets and prefixes
2. **Stateless**: No database lookups per request
3. **Temporary**: Credentials expire after configured time
4. **Auditable**: All operations log org_id
5. **Standard**: Uses AWS IAM policy format
6. **Secure**: Policy embedded in signed JWT, cannot be tampered with

## Development

Run locally:
```bash
# Start MinIO
docker run -p 9000:9000 -p 9001:9001 \
  -e MINIO_ROOT_USER=minioadmin \
  -e MINIO_ROOT_PASSWORD=minioadmin \
  quay.io/minio/minio server /data --console-address ":9001"

# Start api-minio
cd api-minio
go run main.go
```

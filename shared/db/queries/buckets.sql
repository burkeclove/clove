-- name: CreateS3Bucket :one
INSERT INTO s3_buckets (name, organization_id)
VALUES ($1, $2)
RETURNING *;

-- name: GetBucketByName :one
SELECT FROM s3_buckets 
WHERE name = $1 and organization_id = $2;

-- name: DeleteS3BucketById :one
DELETE FROM s3_buckets
WHERE id = $1
RETURNING *;

-- name: DeleteS3BucketByName :one
DELETE FROM s3_buckets
WHERE name = $1
  AND organization_id = $2
RETURNING *;

-- name: CreateApiKey :one
INSERT INTO api_keys (name, organization_id)
VALUES ($1, $2)
RETURNING *;

-- name: GetApiKeyByID :one
SELECT * FROM api_keys WHERE id = $1;

-- name: ListApiKeys :many
SELECT * FROM api_keys ORDER BY created_at DESC LIMIT $1;

-- name: GetOrgFromApiKey :one
SELECT organizations.* FROM api_keys 
JOIN organizations on api_keys.organization_id = organizations.id
WHERE api_keys.key_hash = $1
LIMIT 1;

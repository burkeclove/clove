-- name: CreateOrganization :one
INSERT INTO organizations (name)
VALUES ($1)
RETURNING *;

-- -- name: GetOrganizationFromApiKey :one
SELECT organizations.* FROM api_keys
JOIN organizations ON organizations.id = api_keys.organization_id 
WHERE api_keys.key_hash = $1
LIMIT 1;

-- name: GetOrganizationsFromUserId :many
SELECT organizations.* FROM organization_users
JOIN organizations ON organizations.id = organization_users.organization_id 
WHERE organization_users.user_id = $1;

-- name: GetOrganizationByID :one
SELECT * FROM organizations WHERE id = $1;

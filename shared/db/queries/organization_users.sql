-- name: CheckOrganizationUserExists :one
SELECT EXISTS (
    SELECT 1 FROM organization_users 
    WHERE organization_id = $1 AND user_id = $2
);

-- name: GetOrganizationUser :one
SELECT * FROM organization_users WHERE organization_id = $1 AND user_id = $2 LIMIT 1;

-- name: CreateOrganizationUser :one
INSERT INTO organization_users (organization_id, user_id) 
VALUES ($1, $2)
RETURNING *;

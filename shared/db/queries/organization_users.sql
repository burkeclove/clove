-- name: CheckOrganizationUserExists :one
SELECT EXISTS (
    SELECT 1 FROM organization_users 
    WHERE organization_id = $1 AND user_id = $2
);

-- name: GetOrganizationUser :one
SELECT * FROM organization_users WHERE organization_id = $1 AND user_id = $2 LIMIT 1;

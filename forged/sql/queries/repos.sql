-- name: InsertRepo :one
INSERT INTO repos (group_id, name, description, contrib_requirements)
VALUES ($1, $2, $3, $4)
RETURNING id;

-- name: GetRepoByGroupAndName :one
SELECT id, name, COALESCE(description, '') AS description
FROM repos
WHERE group_id = $1 AND name = $2;

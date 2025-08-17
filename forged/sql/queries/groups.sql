-- name: GetRootGroups :many
SELECT name, COALESCE(description, '') FROM groups WHERE parent_group IS NULL;

-- name: GetGroupIDDescByPath :one
WITH RECURSIVE group_path_cte AS (
	SELECT
		id,
		parent_group,
		name,
		1 AS depth
	FROM groups
	WHERE name = ($1::text[])[1]
		AND parent_group IS NULL

	UNION ALL

	SELECT
		g.id,
		g.parent_group,
		g.name,
		group_path_cte.depth + 1
	FROM groups g
	JOIN group_path_cte ON g.parent_group = group_path_cte.id
	WHERE g.name = ($1::text[])[group_path_cte.depth + 1]
		AND group_path_cte.depth + 1 <= cardinality($1::text[])
)
SELECT c.id, COALESCE(g.description, '')
FROM group_path_cte c
JOIN groups g ON g.id = c.id
WHERE c.depth = cardinality($1::text[]);

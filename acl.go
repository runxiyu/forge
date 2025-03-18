// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileContributor: Runxi Yu <https://runxiyu.org>

package main

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

// get_path_perm_by_group_repo_key returns the filesystem path and direct
// access permission for a given repo and a provided ssh public key.
func get_path_perm_by_group_repo_key(ctx context.Context, group_path []string, repo_name, ssh_pubkey string) (repo_id int, filesystem_path string, access bool, contrib_requirements, user_type string, user_id int, err error) {
	err = database.QueryRow(ctx, `
WITH RECURSIVE group_path_cte AS (
	-- Start: match the first name in the path where parent_group IS NULL
	SELECT
		id,
		parent_group,
		name,
		1 AS depth
	FROM groups
	WHERE name = ($1::text[])[1]
		AND parent_group IS NULL

	UNION ALL

	-- Recurse: join next segment of the path
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
SELECT
	r.id,
	r.filesystem_path,
	CASE WHEN ugr.user_id IS NOT NULL THEN TRUE ELSE FALSE END AS has_role_in_group,
	r.contrib_requirements,
	COALESCE(u.type, ''),
	COALESCE(u.id, 0)
FROM group_path_cte g
JOIN repos r ON r.group_id = g.id
LEFT JOIN ssh_public_keys s ON s.key_string = $3
LEFT JOIN users u ON u.id = s.user_id
LEFT JOIN user_group_roles ugr ON ugr.group_id = g.id AND ugr.user_id = u.id
WHERE g.depth = cardinality($1::text[])
	AND r.name = $2
`, pgtype.FlatArray[string](group_path), repo_name, ssh_pubkey,
	).Scan(&repo_id, &filesystem_path, &access, &contrib_requirements, &user_type, &user_id)
	return
}

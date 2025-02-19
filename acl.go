package main

import (
	"context"
)

// get_path_perm_by_group_repo_key returns the filesystem path and direct
// access permission for a given repo and a provided ssh public key.
func get_path_perm_by_group_repo_key(ctx context.Context, group_name, repo_name, ssh_pubkey string) (repo_id int, filesystem_path string, access bool, contrib_requirements string, user_type string, user_id int, err error) {
	err = database.QueryRow(ctx,
		`SELECT 
		r.id,
		r.filesystem_path,
		CASE
			WHEN ugr.user_id IS NOT NULL THEN TRUE
			ELSE FALSE
		END AS has_role_in_group,
		r.contrib_requirements,
		COALESCE(u.type, ''),
		COALESCE(u.id, 0)
		FROM 
			groups g
		JOIN 
			repos r ON r.group_id = g.id
		LEFT JOIN 
			ssh_public_keys s ON s.key_string = $3
		LEFT JOIN 
			users u ON u.id = s.user_id
		LEFT JOIN 
			user_group_roles ugr ON ugr.group_id = g.id AND ugr.user_id = u.id
		WHERE 
			g.name = $1
		AND r.name = $2;`,
		group_name, repo_name, ssh_pubkey,
	).Scan(&repo_id, &filesystem_path, &access, &contrib_requirements, &user_type, &user_id)
	return
}

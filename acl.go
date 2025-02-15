package main

import (
	"context"
)

func get_path_perm_by_group_repo_key(ctx context.Context, group_name, repo_name, ssh_pubkey string) (filesystem_path string, access bool, err error) {
	err = database.QueryRow(ctx,
		`SELECT 
		r.filesystem_path,
		CASE
			WHEN ugr.user_id IS NOT NULL THEN TRUE
			ELSE FALSE
		END AS has_role_in_group
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
	).Scan(&filesystem_path, &access)
	return
}

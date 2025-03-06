// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileContributor: Runxi Yu <https://runxiyu.org>

package main

import (
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func handle_group_index(w http.ResponseWriter, r *http.Request, params map[string]any) {
	var group_path []string
	var repos []name_desc_t
	var subgroups []name_desc_t
	var err error
	var group_description string

	group_path = params["group_path"].([]string)

	// The group itself
	err = database.QueryRow(r.Context(), `
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
		SELECT COALESCE(g.description, '')
		FROM group_path_cte c
		JOIN groups g ON g.id = c.id
		WHERE c.depth = cardinality($1::text[])
	`,
		pgtype.FlatArray[string](group_path),
	).Scan(&group_description)

	if err == pgx.ErrNoRows {
		http.Error(w, "Group not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Error getting group: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Repos
	var rows pgx.Rows
	rows, err = database.Query(r.Context(), `
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
	SELECT r.name, COALESCE(r.description, '')
	FROM group_path_cte c
	JOIN repos r ON r.group_id = c.id
	WHERE c.depth = cardinality($1::text[])
	`,
		pgtype.FlatArray[string](group_path),
	)
	if err != nil {
		http.Error(w, "Error getting repos: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var name, description string
		if err = rows.Scan(&name, &description); err != nil {
			http.Error(w, "Error getting repos: "+err.Error(), http.StatusInternalServerError)
			return
		}
		repos = append(repos, name_desc_t{name, description})
	}
	if err = rows.Err(); err != nil {
		http.Error(w, "Error getting repos: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Subgroups
	rows, err = database.Query(r.Context(), `
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
	SELECT g.name, COALESCE(g.description, '')
	FROM group_path_cte c
	JOIN groups g ON g.parent_group = c.id
	WHERE c.depth = cardinality($1::text[])
	`,
		pgtype.FlatArray[string](group_path),
	)
	if err != nil {
		http.Error(w, "Error getting subgroups: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var name, description string
		if err = rows.Scan(&name, &description); err != nil {
			http.Error(w, "Error getting subgroups: "+err.Error(), http.StatusInternalServerError)
			return
		}
		subgroups = append(subgroups, name_desc_t{name, description})
	}
	if err = rows.Err(); err != nil {
		http.Error(w, "Error getting subgroups: "+err.Error(), http.StatusInternalServerError)
		return
	}

	params["repos"] = repos
	params["subgroups"] = subgroups
	params["description"] = group_description

	fmt.Println(group_path)

	render_template(w, "group", params)
}
